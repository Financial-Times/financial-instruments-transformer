package main

import (
	"strings"
	"testing"
)

func TestFetchFactsetSecurities_OneEquityAndOneMBSecurity_OnlyEquityIsReturned(t *testing.T) {
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

func TestFetchFactsetSecurities_FirstLineIsAnEquity_FirstLineIsIgnored(t *testing.T) {
	securities :=
		`"00000CY56"|"US00000CY568"|""|""|"D36R2K-S"|"0D1MLR-E"|"GREENWICH CAP ACCEPTANCE  1991-B B1"|"US"|"EQ"|""|1991-01-01||""|"USD"|"US65"||`

	fis := fetchSecurities(strings.NewReader(securities))
	if len(fis) != 0 {
		t.Errorf("Expected first line to be skipped!")
	}
}
