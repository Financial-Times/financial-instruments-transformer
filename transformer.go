package main

import (
	"crypto/md5"
	"io"
	"time"

	"github.com/pborman/uuid"
)

const secToFIGIs = "sym_bbg"
const securityEntityMap = "sym_sec_entity"
const securities = "sym_coverage"

type fiTransformer interface {
	Transform() map[string]financialInstrument
}

type fiTransformerImpl struct {
	loader loader
	parser fiParser
}

type fiMappings struct {
	figiCodeToSecurityIDs               map[string][]string
	securityIDtoRawFinancialInstruments map[string][]rawFinancialInstrument
	securityIDToEntityMapping           map[string]string
}

//same as in org-transformer
func doubleMD5Hash(input string) string {
	h := md5.New()
	io.WriteString(h, input)
	return uuid.NewMD5(uuid.UUID{}, h.Sum(nil)).String()
}

func (fit *fiTransformerImpl) Transform() map[string]financialInstrument {
	infoLogger.Println("Started loading FIs.")
	start := time.Now()

	fiData, err := getMappings(*fit)
	if err != nil {
		errorLogger.Println(err)
		return map[string]financialInstrument{}
	}
	//fmt.Println(fiData.securityIDtoRawFinancialInstruments)
	fis := make(map[string]financialInstrument)
	for figi, secIDs := range fiData.figiCodeToSecurityIDs {
		var rawFIsForFIGI []rawFinancialInstrument
		for _, sID := range secIDs {
			rawFIsForFIGI = append(rawFIsForFIGI, fiData.securityIDtoRawFinancialInstruments[sID]...)
		}
		count := 0
		for _, r := range rawFIsForFIGI {
			count++
			uid := uuid.NewMD5(uuid.UUID{}, []byte(r.securityID)).String()
			fis[uid] = financialInstrument{
				figiCode:     figi,
				orgID:        doubleMD5Hash(r.orgID),
				securityID:   r.securityID,
				securityName: r.securityName,
			}

		}
		if count > 1 {
			warnLogger.Printf("More raw fi mappings with empty termination date for FIGI: [%s]! using the last one [%v]", figi, rawFIsForFIGI)
		}

	}

	infoLogger.Printf("Loading FIs finished in [%v]", time.Since(start))
	infoLogger.Printf("Nr of FIs: [%v]", len(fis))

	return fis
}

func getMappings(fit fiTransformerImpl) (fiMappings, error) {
	secReader, err := fit.loader.LoadResource(securities)
	if err != nil {
		return fiMappings{}, err
	}
	defer secReader.Close()

	secToOrgReader, err := fit.loader.LoadResource(securityEntityMap)
	if err != nil {
		return fiMappings{}, err
	}
	defer secToOrgReader.Close()

	fis, err := fit.parser.parseFIs(secReader, secToOrgReader)
	if err != nil {
		return fiMappings{}, err
	}

	lisReader, err := fit.loader.LoadResource(securities)
	if err != nil {
		return fiMappings{}, err
	}
	defer lisReader.Close()
	listings := fit.parser.parseListings(lisReader, fis)

	figiReader, err := fit.loader.LoadResource(secToFIGIs)
	if err != nil {
		return fiMappings{}, err
	}
	defer figiReader.Close()

	figis, err := fit.parser.parseFIGICodes(figiReader, listings)
	if err != nil {
		return fiMappings{}, err
	}

	return fiMappings{
		securityIDtoRawFinancialInstruments: fis,
		figiCodeToSecurityIDs:               figis,
	}, nil
}
