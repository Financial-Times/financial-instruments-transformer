package main

type financialInstrument struct {
	figiCode     string
	securityID   string
	orgID        string //UPP UUID
	securityName string
}

// raw financial instrument model as it comes from Factset
type rawFinancialInstrument struct {
	securityID      string //FS_PERM_SEC_ID
	orgID           string //FACTSET_ENTITY_ID
	fiType          string //ISSUE_TYPE
	securityName    string //SECURITY_NAME
	inceptionDate   string
	terminationDate string
}

type s3Config struct {
	accKey    string
	secretKey string
	bucket    string
	domain    string
}
