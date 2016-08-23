package main

import (
	"bufio"
	"io"
	"strings"
)

type fiParser interface {
	ParseFis(r io.ReadCloser) (map[string][]rawFinancialInstrument, error)
	ParseFigiCodes(r io.ReadCloser) (map[string][]string, error)
}

type fiParserImpl struct {
}

func (fip *fiParserImpl) ParseFis(r io.ReadCloser) (map[string][]rawFinancialInstrument, error) {
	rawFIs := FetchSecurities(r)
	err := r.Close()
	if err != nil {
		return nil, err
	}
	infoLogger.Printf("Fetched securities. Nr of records: [%d]", len(rawFIs))

	return rawFIs, nil
}

func FetchSecurities(r io.Reader) map[string][]rawFinancialInstrument {
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
				securityName:    record[6],
			}
			rawFIs[securityID] = append(rawFIs[securityID], equity)
		}
	}

	return rawFIs
}

func (fip *fiParserImpl) ParseFigiCodes(r io.ReadCloser) (map[string][]string, error) {
	figiCodes := FetchFIGICodes(r)
	err := r.Close()
	if err != nil {
		return nil, err
	}
	infoLogger.Printf("Fetched figi codes. Nr of records: [%d]", len(figiCodes))

	return figiCodes, nil
}

func FetchFIGICodes(r io.Reader) map[string][]string {
	figiCodes := make(map[string][]string)
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first line
	for scanner.Scan() {
		record := strings.Split(strings.Replace(scanner.Text(), `"`, ``, -1), "|")
		figiCodes[record[1]] = append(figiCodes[record[1]], record[0])
	}
	return figiCodes
}
