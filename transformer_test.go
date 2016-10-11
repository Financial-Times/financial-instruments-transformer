package main

import (
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

type loaderMock struct {
	mockLoadResources func(name string) (io.ReadCloser, error)
}

type parserMock struct {
	mockParseFIs       func() (map[string]rawFinancialInstrument, error)
	mockParseFIGICodes func() (map[string]string, error)
	mockParseListings  func() map[string]string
}

func (l *loaderMock) LoadResource(name string) (io.ReadCloser, error) {
	return l.mockLoadResources(name)
}

func (p *parserMock) parseFIs(r1, r2 io.ReadCloser) (map[string]rawFinancialInstrument, error) {
	return p.mockParseFIs()
}

func (p *parserMock) parseFIGICodes(r io.ReadCloser, m map[string]string) (map[string]string, error) {
	return p.mockParseFIGICodes()
}

func (p *parserMock) parseListings(r io.ReadCloser, m map[string]rawFinancialInstrument) map[string]string {
	return p.mockParseListings()
}

var lm = &loaderMock{
	mockLoadResources: func(name string) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader("")), nil
	},
}

type s3LoaderMock struct {
	mockLoad func(name string) (io.ReadCloser, error)
}

func (s3 *s3LoaderMock) LoadResource(name string) (io.ReadCloser, error) {
	return s3.mockLoad(name)
}

type errorCloser struct {
	reader io.Reader
}

func (errorCloser) Close() error {
	return closerError
}

var closerError = errors.New("Error while reading from the reader")
var loaderError = errors.New("Error loading resource")

func TestGetMappings(t *testing.T) {
	var tests = []struct {
		lm       loader
		pm       *parserMock
		err      error
		expected fiMappings
	}{
		// loader error
		{
			lm: &loaderMock{
				mockLoadResources: func(name string) (io.ReadCloser, error) {
					return nil, loaderError
				},
			},
			pm:       &parserMock{},
			err:      loaderError,
			expected: fiMappings{},
		},
		// edge cases
		{
			lm: lm,
			pm: &parserMock{
				mockParseFIs: func() (map[string]rawFinancialInstrument, error) {
					return nil, nil
				},
				mockParseListings: func() map[string]string {
					return nil
				},
				mockParseFIGICodes: func() (map[string]string, error) {
					return nil, nil
				},
			},
			err:      nil,
			expected: fiMappings{},
		},
		{
			lm: lm,
			pm: &parserMock{
				mockParseFIs: func() (map[string]rawFinancialInstrument, error) {
					return map[string]rawFinancialInstrument{}, nil
				},
				mockParseListings: func() map[string]string {
					return map[string]string{}
				},
				mockParseFIGICodes: func() (map[string]string, error) {
					return map[string]string{}, nil
				},
			},
			err: nil,
			expected: fiMappings{
				figiCodeToSecurityIDs:               map[string]string{},
				securityIDtoRawFinancialInstruments: map[string]rawFinancialInstrument{},
			},
		},
		// happy case
		{
			lm: lm,
			pm: &parserMock{
				mockParseFIs: func() (map[string]rawFinancialInstrument, error) {
					return map[string]rawFinancialInstrument{
						"ABCDEF-S": rawFinancialInstrument{
							securityID:       "ABCDEF-S",
							orgID:            "MNBVCX-E",
							fiType:           "EQ",
							securityName:     "foobar INC",
							primaryListingID: "LKJHHM-L",
						},
					}, nil
				},
				mockParseListings: func() map[string]string {
					return map[string]string{
						"LKJHHM-L": "ABCDEF-S",
					}
				},
				mockParseFIGICodes: func() (map[string]string, error) {
					return map[string]string{
						"BBG000123NMAV": "ABCDEF-S",
					}, nil
				},
			},
			err: nil,
			expected: fiMappings{
				figiCodeToSecurityIDs: map[string]string{
					"BBG000123NMAV": "ABCDEF-S",
				},
				securityIDtoRawFinancialInstruments: map[string]rawFinancialInstrument{
					"ABCDEF-S": rawFinancialInstrument{
						securityID:       "ABCDEF-S",
						orgID:            "MNBVCX-E",
						fiType:           "EQ",
						securityName:     "foobar INC",
						primaryListingID: "LKJHHM-L",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		m, err := getMappings(fiTransformerImpl{tc.lm, tc.pm})
		if err != tc.err {
			t.Errorf("Expected error: [%v]. Actual: [%v]", tc.err, err)
		}
		if !reflect.DeepEqual(m, tc.expected) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, m)
		}
	}
}

func TestTransformMappings(t *testing.T) {
	var tests = []struct {
		figisToSecIDs  map[string]string
		secIDstoRawFIs map[string]rawFinancialInstrument
		expected       map[string]financialInstrument
	}{
		// edge cases
		{
			figisToSecIDs:  nil,
			secIDstoRawFIs: nil,
			expected:       map[string]financialInstrument{},
		},
		{
			figisToSecIDs:  map[string]string{},
			secIDstoRawFIs: map[string]rawFinancialInstrument{},
			expected:       map[string]financialInstrument{},
		},
		// happy case
		{
			figisToSecIDs: map[string]string{
				"BBG000123NMAV": "ABCDEF-S",
			},
			secIDstoRawFIs: map[string]rawFinancialInstrument{
				"ABCDEF-S": rawFinancialInstrument{
					securityID:       "ABCDEF-S",
					orgID:            "MNBVCX-E",
					fiType:           "EQ",
					securityName:     "foobar INC",
					primaryListingID: "LKJHHM-L",
				},
			},
			expected: map[string]financialInstrument{
				"fd0d50ba-7031-3ebf-a594-4806b65a74bd": financialInstrument{
					figiCode:     "BBG000123NMAV",
					securityID:   "ABCDEF-S",
					orgID:        "6f2a22e5-2fb6-304e-b92b-1438f306dc94",
					securityName: "foobar INC",
				},
			},
		},
	}

	for _, tc := range tests {
		tcM := fiMappings{tc.figisToSecIDs, tc.secIDstoRawFIs}

		fis := transformMappings(tcM)
		if !reflect.DeepEqual(fis, tc.expected) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, fis)
		}
	}
}

func TestDoubleMD5Hash(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{
			"0F03DX-E",
			"5a9c7643-31e4-3bad-b6ba-a7676f43da9f",
		},
		{
			"0D1MLR-F",
			"385972c6-f8c1-3878-8e5f-7dd05a20f01b",
		},
	}

	for _, tc := range testCases {
		actual := doubleMD5Hash(tc.input)
		if tc.expected != actual {
			t.Errorf("Expected: [%s]. Actual: [%s]", tc.expected, actual)
		}
	}
}
