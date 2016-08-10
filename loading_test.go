package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestFetchSecurities_FirstLineIsAnEquity_FirstLineIsIgnored(t *testing.T) {
	securities :=
		`"00000CY56"|"US00000CY568"|""|""|"D36R2K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||`

	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 0 {
		t.Errorf("Expected first line to be skipped!")
	}
}

func TestFetchSecurities_NoEquity_EmptyMapReturned(t *testing.T) {
	securities :=
		`"CUSIP"|"ISIN"|"FDS_PRIMARY_SEDOL"|"FDS_PRIMARY_TICKER_SYMBOL"|"FS_PERM_SEC_ID"|"FACTSET_ENTITY_ID"|"SECURITY_NAME"|"ISO_COUNTRY"|"ISSUE_TYPE"|"FDS_PRIMARY_MIC_EXCHANGE_CODE"|"INCEPTION_DATE"|"TERMINATION_DATE"|"CAP_GROUP"|"FDS_PRIMARY_ISO_CURRENCY"|"CIC_CODE"|"COUPON_RATE"|"MATURITY_DATE"
		"00000CY57"|"US00000CY568"|""|""|"D36R2K-E"|"0D1MLR-F"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"MB"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00010CY568"|""|""|"D36R4K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"MT"|""|1991-01-01||""|"USD"|"US65"||`
	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 0 {
		t.Errorf("Expected no equity record. Found: [%v]", fis)
	}
}

func TestFetchSecurities_OneEQAndOneMBSecurity_OnlyEquityIsReturned(t *testing.T) {
	securities :=
		`"CUSIP"|"ISIN"|"FDS_PRIMARY_SEDOL"|"FDS_PRIMARY_TICKER_SYMBOL"|"FS_PERM_SEC_ID"|"FACTSET_ENTITY_ID"|"SECURITY_NAME"|"ISO_COUNTRY"|"ISSUE_TYPE"|"FDS_PRIMARY_MIC_EXCHANGE_CODE"|"INCEPTION_DATE"|"TERMINATION_DATE"|"CAP_GROUP"|"FDS_PRIMARY_ISO_CURRENCY"|"CIC_CODE"|"COUPON_RATE"|"MATURITY_DATE"
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-E"|"0D1MLR-F"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"MB"|""|1991-01-01||""|"USD"|"US65"||`

	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 1 {
		t.Errorf("Expected one equity record. Found: [%d]", len(fis))
	}
	fi, present := fis["D36R2K-E"]
	if !present {
		t.Errorf("Expected equity with secID: [%s].", fi)
	}
	if len(fi) != 1 {
		t.Errorf("Expected one equity for this security id. Found: [%d]", len(fi))
	}
}

func TestFetchSecurities_TwoEQsWithSameSecID_OneEntryWithTwoFIsReturned(t *testing.T) {
	securities :=
		`"CUSIP"|"ISIN"|"FDS_PRIMARY_SEDOL"|"FDS_PRIMARY_TICKER_SYMBOL"|"FS_PERM_SEC_ID"|"FACTSET_ENTITY_ID"|"SECURITY_NAME"|"ISO_COUNTRY"|"ISSUE_TYPE"|"FDS_PRIMARY_MIC_EXCHANGE_CODE"|"INCEPTION_DATE"|"TERMINATION_DATE"|"CAP_GROUP"|"FDS_PRIMARY_ISO_CURRENCY"|"CIC_CODE"|"COUPON_RATE"|"MATURITY_DATE"
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-E"|"0D1MLR-F"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-E"|"0D1MLR-T"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"MB"|""|1991-01-01||""|"USD"|"US65"||`

	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 1 {
		t.Errorf("Expected one equity record. Found: [%d]", len(fis))
	}
	fi, present := fis["D36R2K-E"]
	if !present {
		t.Errorf("Expected equity with secID: [%s].", fi)
	}
	if len(fi) != 2 {
		t.Errorf("Expected two equity for this security id. Found: [%d]", len(fi))
	}
}

func TestFetchSecurities_TwoEQsWithDifferentSecID_OneEntryWithTwoFIsReturned(t *testing.T) {
	securities :=
		`"CUSIP"|"ISIN"|"FDS_PRIMARY_SEDOL"|"FDS_PRIMARY_TICKER_SYMBOL"|"FS_PERM_SEC_ID"|"FACTSET_ENTITY_ID"|"SECURITY_NAME"|"ISO_COUNTRY"|"ISSUE_TYPE"|"FDS_PRIMARY_MIC_EXCHANGE_CODE"|"INCEPTION_DATE"|"TERMINATION_DATE"|"CAP_GROUP"|"FDS_PRIMARY_ISO_CURRENCY"|"CIC_CODE"|"COUPON_RATE"|"MATURITY_DATE"
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-E"|"0D1MLR-F"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-T"|"0D1MLR-T"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||
		"00000CY56"|"US00000CY568"|""|""|"D36R2K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"MB"|""|1991-01-01||""|"USD"|"US65"||`

	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 2 {
		t.Errorf("Expected one equity record. Found: [%d]", len(fis))
	}
	for _, secID := range []string{"D36R2K-E", "D36R2K-T"} {
		fi, present := fis[secID]
		if !present {
			t.Errorf("Expected equity with secID: [%s].", fi)
		}
		if len(fi) != 1 {
			t.Errorf("Expected one equity for this security id. Found: [%d]", len(fi))
		}
	}
}

func TestFetchSecurities_OneEQ_RawFIModelBuiltCorrectly(t *testing.T) {
	securities :=
		`"CUSIP"|"ISIN"|"FDS_PRIMARY_SEDOL"|"FDS_PRIMARY_TICKER_SYMBOL"|"FS_PERM_SEC_ID"|"FACTSET_ENTITY_ID"|"SECURITY_NAME"|"ISO_COUNTRY"|"ISSUE_TYPE"|"FDS_PRIMARY_MIC_EXCHANGE_CODE"|"INCEPTION_DATE"|"TERMINATION_DATE"|"CAP_GROUP"|"FDS_PRIMARY_ISO_CURRENCY"|"CIC_CODE"|"COUPON_RATE"|"MATURITY_DATE"
		"FDS09ZAE9"|"KRFDS09ZAE94"|"BTHH0M8"|"208140-KR"|"T621V4-S-KR"|"0F03DX-E"|"LIG SPECIAL PURPOSE ACQ 2ND CO  ORD"|"KR"|"EQ"|"XKRX"|2013-11-21||"MICRO"|"KRW"|"KR31"||`
	secID := "T621V4-S-KR"
	orgID := "0F03DX-E"
	fiType := "EQ"
	inceptionDate := "2013-11-21"
	terminationDate := ""
	fis := fetchSecurities(strings.NewReader(securities))
	fi, present := fis[secID]
	if !present {
		t.Errorf("Expected to find financial instrument with secID [%s], but does not exist.", secID)
	}
	eq := fi[0]

	if eq.securityID != secID ||
		eq.orgID != orgID ||
		eq.fiType != fiType ||
		eq.inceptionDate != inceptionDate ||
		eq.terminationDate != terminationDate {
		t.Errorf("Expected secID=[%s], orgID=[%s], fiType=[%s], inceptionDate=[%s], terminationDate=[%s].\nFound: [%v]", secID, orgID, fiType, terminationDate, inceptionDate, fi)
	}
}

func TestFetchFIGICodes(t *testing.T) {
	headerLine := `"FS_PERM_SEC_ID"|"BBG_ID"`
	var testCases = []struct {
		figis    string
		expected figiCodeToSecurityIDs
	}{
		{
			`"B000BB-S"|"BBG000Y1HJT8"` + "\n" + `"B000CC-S"|"BBG000798XK9"`,
			map[string][]string{
				"BBG000Y1HJT8": []string{"B000BB-S"},
				"BBG000798XK9": []string{"B000CC-S"},
			},
		},
		{
			`"B000BB-S"|"BBG000Y1HJT8"` + "\n" + `"B000CC-S"|"BBG000Y1HJT8"`,
			map[string][]string{
				"BBG000Y1HJT8": []string{"B000BB-S", "B000CC-S"},
			},
		},
	}

	for _, tc := range testCases {
		figis := fetchFIGICodes(strings.NewReader(headerLine + "\n" + tc.figis))
		if !reflect.DeepEqual(figis, tc.expected) {
			t.Errorf("Expected: [%v]. Actual: [%v]", tc.expected, figis)
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

func TestTransform_OneEntryInBothMappings_TransformedFIIsCorrect(t *testing.T) {
	rawFIs := map[string][]rawFinancialInstrument{
		"T621V4-S-KR": []rawFinancialInstrument{
			rawFinancialInstrument{
				securityID:      "T621V4-S-KR",
				orgID:           "0F03DX-E",
				inceptionDate:   "1995-03-12",
				terminationDate: "",
			}}}
	figis := map[string][]string{
		"BBG007HW10F7": []string{"T621V4-S-KR"},
	}
	expectedKey := "e81c2fed-1fe3-3b38-9e82-0f64d3074281"
	expectedSecID := "T621V4-S-KR"
	expectedOrgID := "5a9c7643-31e4-3bad-b6ba-a7676f43da9f"
	expectedFigiCode := "BBG007HW10F7"

	fis := transform(rawFIs, figis)

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
}
