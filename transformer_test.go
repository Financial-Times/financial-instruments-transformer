package main

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

type loaderMock struct {
	mockLoadResources func(name string) (io.ReadCloser, error)
}

type parserMock struct {
	mockParseFis       func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error)
	mockParseFigiCodes func(r io.ReadCloser) (map[string][]string, error)
}

func (l *loaderMock) LoadResource(name string) (io.ReadCloser, error) {
	return l.mockLoadResources(name)
}

func (p *parserMock) ParseFis(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
	return p.mockParseFis(r)
}

func (p *parserMock) ParseFigiCodes(r io.ReadCloser) (map[string][]string, error) {
	return p.mockParseFigiCodes(r)
}

var lm = &loaderMock{
	mockLoadResources: func(name string) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader("")), nil
	},
}

func TestTransform_OneEntryInBothMappings_TransformedFIIsCorrect(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"T621V4-S-KR": []rawFinancialInstrument{
					{
						securityID:      "T621V4-S-KR",
						orgID:           "0F03DX-E",
						inceptionDate:   "1995-03-12",
						terminationDate: "",
						securityName:    "LIG SPECIAL PURPOSE ACQ 2ND CO  ORD",
					}}}, nil
		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG007HW10F7": []string{"T621V4-S-KR"},
			}, nil
		},
	}

	fit := &fiTransformerImpl{
		parser: pm,
		loader: lm,
	}

	expectedKey := "e81c2fed-1fe3-3b38-9e82-0f64d3074281"
	expectedSecID := "T621V4-S-KR"
	expectedOrgID := "5a9c7643-31e4-3bad-b6ba-a7676f43da9f"
	expectedFigiCode := "BBG007HW10F7"
	expectedSecurityName := "LIG SPECIAL PURPOSE ACQ 2ND CO  ORD"

	fis := fit.Transform()

	fi, present := fis[expectedKey]
	if !present {
		t.Errorf("Expected fi with uuid: [%s]", expectedKey)
	}
	if fi.securityID != expectedSecID {
		t.Errorf("Expected secID: [%s]. Found: [%s]", expectedSecID, fi.securityID)
	}
	if fi.orgID != expectedOrgID {
		t.Errorf("Expected orgID: [%s]. Found: [%s]", expectedOrgID, fi.orgID)
	}
	if fi.figiCode != expectedFigiCode {
		t.Errorf("Expected FIGI [%s]. Found: [%s]", expectedFigiCode, fi.figiCode)
	}

	if fi.securityName != expectedSecurityName {
		t.Errorf("Expected security name [%s]. Found: [%s]", expectedSecurityName, fi.securityName)
	}
}

func TestTransform_TwoEntriesWithSameId_OnlyTheOneWithTerminationDateIsUsed(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"T621V4-S-KR": []rawFinancialInstrument{
					{
						securityID:      "T621V4-S-KR",
						orgID:           "0F03DX-T",
						inceptionDate:   "1994-07-18",
						terminationDate: "2005-07-31",
					},
					{
						securityID:      "T621V4-S-KR",
						orgID:           "0F03DX-E",
						inceptionDate:   "2005-07-18",
						terminationDate: "",
					}}}, nil
		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG007HW10F7": []string{"T621V4-S-KR"},
			}, nil
		},
	}

	fit := fiTransformerImpl{
		loader: lm,
		parser: pm,
	}

	expectedKey := "e81c2fed-1fe3-3b38-9e82-0f64d3074281"
	expectedOrgID := "5a9c7643-31e4-3bad-b6ba-a7676f43da9f"

	fis := fit.Transform()
	fi, present := fis[expectedKey]

	if !present {
		t.Errorf("Expected one financial instrument for this security id. Found: [%d]", len(fis))
	}

	if fi.orgID != expectedOrgID {
		t.Errorf("Expecting financial instrument with orgID: [%s]. Found [%s]", expectedOrgID, fi.orgID)
	}
}

func TestTransform_TwoEntriesWithSameId_OnlyTheLastIsUsed(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"T621V4-S-KR": []rawFinancialInstrument{
					{
						securityID:      "T621V4-S-KR",
						orgID:           "0F03DX-T",
						inceptionDate:   "1994-07-18",
						terminationDate: "",
					},
					{
						securityID:      "T621V4-S-KR",
						orgID:           "0F03DX-E",
						inceptionDate:   "2005-07-18",
						terminationDate: "",
					}}}, nil

		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG007HW10F7": []string{"T621V4-S-KR"},
			}, nil
		},
	}

	fit := fiTransformerImpl{
		loader: lm,
		parser: pm,
	}

	expectedKey := "e81c2fed-1fe3-3b38-9e82-0f64d3074281"
	expectedOrgID := "5a9c7643-31e4-3bad-b6ba-a7676f43da9f"

	fis := fit.Transform()
	fi := fis[expectedKey]

	if fi.orgID != expectedOrgID {
		t.Errorf("Expecting financial instrument with orgID: [%s]. Found [%s]", expectedOrgID, fi.orgID)
	}
}

func TestTransform_DifferentSecIdSameFigiAndEntityId(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"LK1Q8B-S-HK": []rawFinancialInstrument{
					{
						securityID:      "LK1Q8B-S-HK",
						orgID:           "0BFB5F-E",
						inceptionDate:   "2011-01-27",
						terminationDate: "2016-03-29",
					},
					{
						securityID:      "LK1Q8B-S-HK",
						orgID:           "0BFB5F-E",
						inceptionDate:   "2016-03-29",
						terminationDate: "",
					}},
				"S67X9X-S-HK": []rawFinancialInstrument{
					{
						securityID:      "S67X9X-S-HK",
						orgID:           "0BFB5F-E",
						inceptionDate:   "2016-04-01",
						terminationDate: "",
					}}}, nil
		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG001D9T252": []string{"LK1Q8B-S-HK", "S67X9X-S-HK"},
			}, nil
		},
	}
	fit := fiTransformerImpl{
		loader: lm,
		parser: pm,
	}

	expectedIDs := []string{
		"a135ff6a-a466-32ba-a48c-ea280cf9df6f",
		"6372c1b4-0dee-3d99-b9a0-bef34d97d6c8",
	}

	fis := fit.Transform()

	if len(fis) != 2 {
		t.Errorf("Expected two financial instruments but found [%d]", len(fis))
	}

	for _, expectedID := range expectedIDs {
		_, present := fis[expectedID]
		if !present {
			t.Errorf("Expected to found financial instrument with uuid: [%s]", expectedID)
		}
	}
}

func TestTransform_SameFigiSameSecIDSameEntityNoTerminationDate(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"VCLG2C-S-IN": []rawFinancialInstrument{
					{
						securityID:      "VCLG2C-S-IN",
						orgID:           "05VD3Y-E",
						inceptionDate:   "2005-07-31",
						terminationDate: "",
					},
					{
						securityID:      "VCLG2C-S-IN",
						orgID:           "05VD3Y-E",
						inceptionDate:   "1994-07-18",
						terminationDate: "2005-07-31",
					},
					{
						securityID:      "VCLG2C-S-IN",
						orgID:           "05VD3Y-E",
						inceptionDate:   "2016-05-06",
						terminationDate: "",
					}}}, nil
		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG000CXGND1": []string{"VCLG2C-S-IN"},
			}, nil
		},
	}
	fit := fiTransformerImpl{
		loader: lm,
		parser: pm,
	}

	expectedID := "03a708f3-88ae-3c8a-a297-efe9fa0ad666"

	fis := fit.Transform()

	if len(fis) != 1 {
		t.Errorf("Expecting one financial instrument but found [%d]", len(fis))
	}

	if _, present := fis[expectedID]; !present {
		t.Errorf("Expected to found financial instrument with uuid: [%s]", expectedID)
	}
}

func TestTransform_SameFigiDifferentSecIDAndOrgIDNoTerminationDate(t *testing.T) {
	pm := &parserMock{
		mockParseFis: func(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
			return map[string][]rawFinancialInstrument{
				"B2GXQ6-S-CL": []rawFinancialInstrument{
					{
						securityID:      "B2GXQ6-S-CL",
						orgID:           "05W5BY-E",
						inceptionDate:   "2001-11-14",
						terminationDate: "",
					}},
				"P7RDT8-S-CA": []rawFinancialInstrument{
					{
						securityID:      "P7RDT8-S-CA",
						orgID:           "003XN9-E",
						inceptionDate:   "2014-12-01",
						terminationDate: "",
					}}}, nil
		},
		mockParseFigiCodes: func(r io.ReadCloser) (map[string][]string, error) {
			return map[string][]string{
				"BBG000BFPHG1": []string{"B2GXQ6-S-CL", "P7RDT8-S-CA"},
			}, nil
		},
	}

	fit := fiTransformerImpl{
		loader: lm,
		parser: pm,
	}

	expectedIDs := []string{
		"24d7f133-d30b-394f-970c-5a5e3ed66061",
		"f9ca6c8d-1b02-3492-ae76-5ef2231e6ae7",
	}

	fis := fit.Transform()

	if len(fis) != 2 {
		t.Errorf("Expecting two financial instruments but found [%d]", len(fis))
	}

	for _, expectedID := range expectedIDs {
		_, present := fis[expectedID]
		if !present {
			t.Errorf("Expected to found financial instrument with uuid: [%s]", expectedID)
		}
	}
}
