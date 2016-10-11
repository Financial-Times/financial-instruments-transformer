package main

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

var testFIParser = &fiParserImpl{}

func wrapInReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func TestParseFIs(t *testing.T) {
	var tests = []struct {
		name         string
		securities   string
		secEntityMap string
		expected     map[string]rawFinancialInstrument
	}{
		// first line is ignored
		{
			securities:   `"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|1|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// fi is parsed with no org ID
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|1|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected: map[string]rawFinancialInstrument{
				"JBP7Z8-S": rawFinancialInstrument{
					securityID:       "JBP7Z8-S",
					fiType:           "EQ",
					securityName:     "Industrija Precizne Mehanike AD",
					primaryListingID: "WHV8G2-R",
					orgID:            "",
				},
			},
		},
		// fi is parsed with org ID
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|1|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: `FSYM_ID"|"FACTSET_ENTITY_ID"` + "\n" +
				`"JBP7Z8-S"|"092VYW-E"`,
			expected: map[string]rawFinancialInstrument{
				"JBP7Z8-S": rawFinancialInstrument{
					securityID:       "JBP7Z8-S",
					fiType:           "EQ",
					securityName:     "Industrija Precizne Mehanike AD",
					primaryListingID: "WHV8G2-R",
					orgID:            "092VYW-E",
				},
			},
		},

		// fi universe type is not equity
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|1|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"ET"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// fi is equity, but not a share
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|1|"PREF"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// fi is a share, but not active
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"WHV8G2-R"|0|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// fi is not a primary-level security
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"WHV8G2-R"|"RSD"|"Industrija Precizne Mehanike AD"|"JBP7Z8-S"|"M679DF-L"|1|"SHARE"|"BEL"|0|1|0|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// secID does not match primary equity ID
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z9-S"|"WHV8G2-R"|0|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
		// primary listing ID is missing
		{
			securities: `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"` + "\n" +
				`"JBP7Z8-S"|""|"Industrija Precizne Mehanike AD"|"JBP7Z9-S"|""|1|"SHARE"|""|0|0|1|"WHV8G2-R"|"JBP7Z8-S"|"EQ"`,
			secEntityMap: ``,
			expected:     map[string]rawFinancialInstrument{},
		},
	}

	for _, tc := range tests {
		fis, err := testFIParser.parseFIs(wrapInReadCloser(tc.securities), wrapInReadCloser(tc.secEntityMap))
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(tc.expected, fis) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, fis)
		}
	}
}

func TestParseListings(t *testing.T) {
	headerLine := `"FSYM_ID"|"CURRENCY"|"PROPER_NAME"|"FSYM_PRIMARY_EQUITY_ID"|"FSYM_PRIMARY_LISTING_ID"|"ACTIVE_FLAG"|"FREF_SECURITY_TYPE"|"FREF_LISTING_EXCHANGE"|"LISTING_FLAG"|"REGIONAL_FLAG"|"SECURITY_FLAG"|"FSYM_REGIONAL_ID"|"FSYM_SECURITY_ID"|"UNIVERSE_TYPE"`
	var testCases = []struct {
		listings     string
		secIDToRawFI map[string]rawFinancialInstrument
		expected     map[string]string
	}{
		// security record is not complete
		{
			listings: `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S"|`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN8-R",
				},
			},
			expected: map[string]string{},
		},
		// not a regional-level security
		{
			listings: `"H73FN8-L"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S"|"MLKNP9-L"|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN8-R",
				},
			},
			expected: map[string]string{},
		},
		// no primary equity
		{
			listings: `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|""|"MLKNP9-L"|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN8-R",
				},
			},
			expected: map[string]string{},
		},
		// no related FI exist
		{
			listings:     `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S"|"MLKNP9-L"|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{},
			expected:     map[string]string{},
		},
		// FI exist, but primary Listing ID does not match
		{
			listings: `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S""|"MLKNP9-L"|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN9-R",
				},
			},
			expected: map[string]string{},
		},
		// no primary listing ID
		{
			listings: `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S""|""|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN8-R",
				},
			},
			expected: map[string]string{},
		},
		// happy case
		{
			listings: `"H73FN8-R"|"GBP"|"Ralph Martindale & Company Ltd"|"GG9B0P-S""|"MLKNP9-L"|1|"SHARE"|"LON"|0|1|0|"H73FN8-R"|"GG9B0P-S"|"EQ"`,
			secIDToRawFI: map[string]rawFinancialInstrument{
				"GG9B0P-S": rawFinancialInstrument{
					securityID:       "GG9B0P-S",
					fiType:           "EQ",
					orgID:            "05MFLC-E",
					securityName:     "Ralph Martindale & Company Ltd",
					primaryListingID: "H73FN8-R",
				},
			},
			expected: map[string]string{
				"MLKNP9-L": "GG9B0P-S",
			},
		},
	}

	for _, tc := range testCases {
		actual := testFIParser.parseListings(wrapInReadCloser(headerLine+"\n"+tc.listings), tc.secIDToRawFI)

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, actual)
		}
	}

}

func TestParseFIGICodes(t *testing.T) {
	headerLine := `"FSYM_ID"|"BBG_ID"|"BBG_TICKER"`
	var testCases = []struct {
		figis    string
		listings map[string]string
		expected map[string]string
	}{
		{
			figis: `"M679DF-L"|"BBG000JPVHS1"|"IPMB SG"`,
			listings: map[string]string{
				"M679DF-L": "JBP7Z8-S",
			},
			expected: map[string]string{
				"BBG000JPVHS1": "JBP7Z8-S",
			},
		},
		{
			figis: `"M679DF-L"|"BBG000JPVHS1"|"IPMB SG"`,
			listings: map[string]string{
				"JBP7Z8-S": "M679DF-R",
			},
			expected: map[string]string{},
		},
	}

	for _, tc := range testCases {
		figis, err := testFIParser.parseFIGICodes(wrapInReadCloser(headerLine+"\n"+tc.figis), tc.listings)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(figis, tc.expected) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, figis)
		}
	}
}
