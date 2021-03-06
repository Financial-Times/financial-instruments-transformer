package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

const publicEntity = "PUB"

type fiParser interface {
	parseFIs(secReader io.Reader, secOrgReader io.Reader) (map[string]rawFinancialInstrument, error)
	parseListings(r io.Reader, fis map[string]rawFinancialInstrument) map[string]string
	parseFIGICodes(r io.Reader, listings map[string]string) (map[string]string, error)
	parseEntityFunc() func(r io.ReadCloser) map[string]bool
}

type fiParserImpl struct{}

func (fip *fiParserImpl) parseFIs(secReader io.Reader, secOrgReader io.Reader) (map[string]rawFinancialInstrument, error) {
	infoLogger.Println("Starting security parsing.")
	rawFIs := make(map[string]rawFinancialInstrument)
	scanner := bufio.NewScanner(secReader)
	scanner.Scan() // skip the first line (contains the column names)
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		if len(record) < 14 {
			infoLogger.Println("Skip raw fi:", record)
			continue
		}
		securityID := record[0]
		universeType := record[13]
		activeFlag, err := strconv.Atoi(record[5])
		if err != nil {
			errorLogger.Println(err)
			continue
		}

		primaryEquityID := record[3]
		primaryListingID := record[4]
		securityType := record[6]

		if universeType == "EQ" &&
			strings.HasSuffix(securityID, "-S") &&
			activeFlag == 1 &&
			primaryEquityID == securityID &&
			securityType == "SHARE" &&
			primaryListingID != "" {

			equity := rawFinancialInstrument{
				securityID:       securityID,
				fiType:           universeType,
				securityName:     record[2],
				primaryListingID: record[4],
			}
			rawFIs[securityID] = equity
		}
	}

	infoLogger.Println("Starting sec-org mapping parsing.")
	scanner = bufio.NewScanner(secOrgReader)
	scanner.Scan() // skip the first line (contains the column names)
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		if len(record) < 2 {
			infoLogger.Println("Skip sec-org mapping:", record)
			continue
		}
		securityID := record[0]
		orgID := record[1]
		if fi, ok := rawFIs[securityID]; ok {
			fi.orgID = orgID
			rawFIs[securityID] = fi
		}
	}

	infoLogger.Printf("Fetched securities. Nr of records: [%d]", len(rawFIs))

	return rawFIs, nil
}

func (fip *fiParserImpl) parseFIGICodes(r io.Reader, listings map[string]string) (map[string]string, error) {
	infoLogger.Println("Starting FIGI code parsing.")
	figiCodes := make(map[string]string)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		if len(record) < 2 {
			infoLogger.Println("Skip figi code:", record)
			continue
		}
		if securityID, ok := listings[record[0]]; ok {
			figiCodes[record[1]] = securityID
		}
	}
	infoLogger.Printf("Fetched figi codes. Nr of records: [%d]", len(figiCodes))

	return figiCodes, nil
}

func (fip *fiParserImpl) parseListings(r io.Reader, fis map[string]rawFinancialInstrument) map[string]string {
	infoLogger.Println("Starting listings parsing.")
	listings := make(map[string]string)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		if len(record) < 5 {
			infoLogger.Println("Skip listing:", record)
			continue
		}
		securityID := record[0]
		primaryEquityID := record[3]

		if strings.HasSuffix(securityID, "-R") && primaryEquityID != "" {
			primaryListingID := record[4]
			if primaryListingID == "" {
				continue
			}

			rawFi, ok := fis[primaryEquityID]
			if !ok || rawFi.primaryListingID != securityID {
				continue
			}
			listings[primaryListingID] = primaryEquityID
		}
	}
	infoLogger.Printf("Fetched listings. Nr of records: [%v]", len(listings))
	return listings
}

func (fip *fiParserImpl) parseEntityFunc() func(r io.ReadCloser) map[string]bool {
	return func(r io.ReadCloser) map[string]bool {
		infoLogger.Println("Starting entity parsing.")
		entities := make(map[string]bool)
		scanner := bufio.NewScanner(r)
		scanner.Scan() // skip first line
		for scanner.Scan() {
			record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
			if len(record) < 12 {
				infoLogger.Println("Skip entity:", record)
				continue
			}
			entityID := record[0]
			entityType := record[11]
			if entityType == publicEntity {
				entities[entityID] = true
			}
		}
		infoLogger.Printf("Fetched public entities. Nr of records: [%v]", len(entities))
		return entities
	}
}
