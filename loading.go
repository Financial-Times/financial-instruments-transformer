package main

import (
	"bufio"
	"crypto/md5"
	"io"
	"strings"

	"github.com/pborman/uuid"
	"github.com/rlmcpherson/s3gof3r"
)

const bbgIDs = "edm_bbg_ids.txt"
const securityEntityMap = "edm_security_entity_map.txt"

type figiCodeToSecurityIDs map[string][]string
type securityIDtoRawFinancialInstruments map[string][]rawFinancialInstrument

// raw financial instrument model as it comes from Factset
type rawFinancialInstrument struct {
	//FS_PERM_SEC_ID
	securityID string
	//FACTSET_ENTITY_ID
	orgID string
	//ISSUE_TYPE
	fiType string

	inceptionDate   string
	terminationDate string
}

func loadFIs(c s3Config) (map[string]financialInstrument, error) {
	k, err := s3gof3r.EnvKeys()
	if err != nil {
		return nil, err
	}
	s3 := s3gof3r.New(c.domain, k)
	b := s3.Bucket(c.bucket)

	r, _, err := b.GetReader(securityEntityMap, nil)
	if err != nil {
		return nil, err
	}
	rawFIs := fetchSecurities(r)
	err = r.Close()
	if err != nil {
		return nil, err
	}
	infoLogger.Printf("Fetched securities. Nr of records: [%d]", len(rawFIs))

	r, _, err = b.GetReader(bbgIDs, nil)
	if err != nil {
		return nil, err
	}
	figiCodes := fetchFIGICodes(r)
	err = r.Close()
	if err != nil {
		return nil, err
	}
	infoLogger.Printf("Fetched figi codes. Nr of records: [%d]", len(figiCodes))

	return transform(rawFIs, figiCodes), nil
}

func fetchSecurities(r io.Reader) securityIDtoRawFinancialInstruments {
	rawFIs := make(map[string][]rawFinancialInstrument)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip the first line (contains the column names)
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		securityID := record[4]
		if record[8] == "EQ" && securityID != "" {
			equity := rawFinancialInstrument{
				securityID:      securityID,
				orgID:           record[5],
				inceptionDate:   record[10],
				terminationDate: record[11],
				fiType:          record[8],
			}
			rawFIs[securityID] = append(rawFIs[securityID], equity)
		}
	}

	return rawFIs
}

//same as in org-transformer
func doubleMD5Hash(input string) string {
	h := md5.New()
	io.WriteString(h, input)
	return uuid.NewMD5(uuid.UUID{}, h.Sum(nil)).String()
}

func fetchFIGICodes(r io.Reader) figiCodeToSecurityIDs {
	figiCodes := make(map[string][]string)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		figiCodes[record[1]] = append(figiCodes[record[1]], record[0])
	}
	return figiCodes
}

func transform(rawFIs map[string][]rawFinancialInstrument, figiCodes map[string][]string) map[string]financialInstrument {
	fis := make(map[string]financialInstrument)
	for figi, secIDs := range figiCodes {
		var rawFIsForFIGI []rawFinancialInstrument
		for _, sID := range secIDs {
			rawFIsForFIGI = append(rawFIsForFIGI, rawFIs[sID]...)
		}
		count := 0
		for _, r := range rawFIsForFIGI {
			if r.terminationDate == "" {
				count++
				uid := uuid.NewMD5(uuid.UUID{}, []byte(r.securityID)).String()
				fis[uid] = financialInstrument{
					figiCode:   figi,
					orgID:      doubleMD5Hash(r.orgID),
					securityID: r.securityID,
				}
			}
		}
		if count > 1 {
			warnLogger.Printf("More raw fi mappings with empty termination date for FIGI: [%s]! using the last one [%v]", figi, rawFIsForFIGI)
		}

	}
	return fis
}
