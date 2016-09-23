package main

type financialInstrument struct {
	figiCode     string
	securityID   string
	orgID        string //UPP UUID
	securityName string
}

// raw financial instrument model as it comes from Factset
type rawFinancialInstrument struct {
	securityID       string
	orgID            string
	fiType           string
	securityName     string
	primaryListingID string
}

type s3Config struct {
	accKey    string
	secretKey string
	bucket    string
	domain    string
}
