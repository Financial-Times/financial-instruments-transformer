package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type fiParser interface {
	ParseFis(secReader io.ReadCloser, secOrgReader io.ReadCloser) (map[string][]rawFinancialInstrument, error)
	ParseListings(r io.ReadCloser, fis map[string][]rawFinancialInstrument) map[string]string
	ParseFigiCodes(r io.ReadCloser, listings map[string]string) (map[string][]string, error)
}

type fiParserImpl struct {
}

func (fip *fiParserImpl) ParseFis(secReader io.ReadCloser, secOrgReader io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
	rawFIs := make(map[string][]rawFinancialInstrument)
	scanner := bufio.NewScanner(secReader)
	scanner.Scan() // skip the first line (contains the column names)
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		securityID := record[0]
		universeType := record[13]
		activeFlag, err := strconv.Atoi(record[5])
		if err != nil {
			continue
		}
		primaryEquityId := record[3]
		primaryListingId := record[4]
		securityType := record[6]

		if universeType == "EQ" && strings.HasSuffix(securityID, "-S") && activeFlag == 1 && primaryEquityId == securityID && securityType == "SHARE" && primaryListingId != "" {
			equity := rawFinancialInstrument{
				securityID:       securityID,
				fiType:           universeType,
				securityName:     record[2],
				primaryListingID: record[4],
			}
			rawFIs[securityID] = append(rawFIs[securityID], equity)
		}
	}

	scanner = bufio.NewScanner(secOrgReader)
	scanner.Scan() // skip the first line (contains the column names)
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		securityID := record[0]
		orgID := record[1]
		rawFI := rawFIs[securityID]
		if rawFI != nil && len(rawFI) == 1 {
			rawFI[0].orgID = orgID
		}
	}
	infoLogger.Printf("Fetched securities. Nr of records: [%d]", len(rawFIs))

	return rawFIs, nil
}

func (fip *fiParserImpl) ParseFigiCodes(r io.ReadCloser, listings map[string]string) (map[string][]string, error) {
	figiCodes := make(map[string][]string)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		if securityID := listings[record[0]]; securityID != "" {
			figiCodes[record[1]] = append(figiCodes[record[1]], securityID)
		}
	}
	infoLogger.Printf("Fetched figi codes. Nr of records: [%d]", len(figiCodes))

	return figiCodes, nil
}

func (fip *fiParserImpl) ParseListings(r io.ReadCloser, fis map[string][]rawFinancialInstrument) map[string]string {
	listings := map[string]string{}
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		securityID := record[0]
		primaryEquityID := record[3]

		if strings.HasSuffix(securityID, "-R") && primaryEquityID != "" {
			primaryListingID := record[4]
			rawFis := fis[primaryEquityID]
			if len(rawFis) == 0 {
				continue
			}
			if len(rawFis) != 1 {
				warnLogger.Printf("Security with ID [%s] has multiple matches [%d]", primaryEquityID, len(rawFis))
				continue
			}
			rawFi := rawFis[0]
			if rawFi.primaryListingID == securityID {
				listings[primaryListingID] = primaryEquityID
			}
		}
	}

	return listings
}
