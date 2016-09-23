package main

import (
	"crypto/md5"
	"io"
	"time"

	"github.com/pborman/uuid"
)

const bbgIDs = "sym_bbg"
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
	fisReader, err := fit.loader.LoadResource(securities)
	if err != nil {
		errorLogger.Println(err)
		return fiMappings{}, err
	}
	defer fisReader.Close()

	figiReader, err := fit.loader.LoadResource(bbgIDs)
	if err != nil {
		errorLogger.Println(err)
		return fiMappings{}, err
	}
	defer figiReader.Close()

	secOrgReader, err := fit.loader.LoadResource(securityEntityMap)
	if err != nil {
		errorLogger.Println(err)
		return fiMappings{}, err
	}
	defer secOrgReader.Close()

	sec, err := fit.parser.ParseFis(fisReader, secOrgReader)
	if err != nil {
		return fiMappings{}, err
	}

	//todo redundant, this file is already loaded, just reset reader
	lisReader, err := fit.loader.LoadResource(securities)
	if err != nil {
		errorLogger.Println(err)
		return fiMappings{}, err
	}
	defer lisReader.Close()
	listings := fit.parser.ParseListings(lisReader, sec)


	figis, err := fit.parser.ParseFigiCodes(figiReader, listings)
	if err != nil {
		return fiMappings{}, err
	}

	return fiMappings{
		securityIDtoRawFinancialInstruments: sec,
		figiCodeToSecurityIDs:               figis,
	}, nil
}
