package main

import (
	"crypto/md5"
	"github.com/pborman/uuid"
	"io"
	"time"
)

const secToFIGIs = "sym_bbg"
const securities = "sym_coverage"
const securityEntityMap = "sym_sec_entity"
const entities = "ent_entity_coverage"

type fiTransformer interface {
	Transform() (map[string]financialInstrument, error)
	checkConnectivityToS3() error
}

type fiTransformerImpl struct {
	loader loader
	parser fiParser
}

type fiMappings struct {
	figiCodeToSecurityIDs               map[string]string
	securityIDtoRawFinancialInstruments map[string]rawFinancialInstrument
}

func (fit *fiTransformerImpl) Transform() (map[string]financialInstrument, error) {
	infoLogger.Println("Started loading FIs.")
	start := time.Now()

	mappings, err := getMappings(*fit)
	if err != nil {
		return map[string]financialInstrument{}, err
	}

	fis := transformMappings(mappings)
	infoLogger.Printf("Loading FIs finished in [%v]", time.Since(start))
	infoLogger.Printf("Nr of FIs: [%v]", len(fis))

	return fis, nil
}

func getMappings(fit fiTransformerImpl) (fiMappings, error) {
	latestResourcesFolderName, err := fit.loader.FindLatestResourcesFolder()
	if err != nil {
		return fiMappings{}, err
	}

	secReader, err := fit.loader.LoadResource(latestResourcesFolderName, securities)
	if err != nil {
		return fiMappings{}, err
	}
	defer secReader.Close()

	secToOrgReader, err := fit.loader.LoadResource(latestResourcesFolderName, securityEntityMap)
	if err != nil {
		return fiMappings{}, err
	}
	defer secToOrgReader.Close()

	fis, err := fit.parser.parseFIs(secReader, secToOrgReader)
	if err != nil {
		return fiMappings{}, err
	}
	// filter only if an entity parsing function exist
	if parseEntities := fit.parser.parseEntityFunc(); parseEntities != nil {
		entReader, err := fit.loader.LoadResource(latestResourcesFolderName, entities)
		if err != nil {
			return fiMappings{}, err
		}
		pubEnts := parseEntities(entReader)
		applyPublicEntityFilter(fis, pubEnts)
	}
	lisReader, err := fit.loader.LoadResource(latestResourcesFolderName, securities)
	if err != nil {
		return fiMappings{}, err
	}
	defer lisReader.Close()
	listings := fit.parser.parseListings(lisReader, fis)

	figiReader, err := fit.loader.LoadResource(latestResourcesFolderName, secToFIGIs)
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

func applyPublicEntityFilter(fis map[string]rawFinancialInstrument, pubEnts map[string]bool) {
	for k, fi := range fis {
		if _, present := pubEnts[fi.orgID]; !present {
			delete(fis, k)
		}
	}
	infoLogger.Println("Number of fis after filtering non-public companies:", len(fis))
}

func transformMappings(fiData fiMappings) map[string]financialInstrument {
	fis := make(map[string]financialInstrument)
	for figi, sID := range fiData.figiCodeToSecurityIDs {
		r := fiData.securityIDtoRawFinancialInstruments[sID]
		uid := uuid.NewMD5(uuid.UUID{}, []byte(r.securityID)).String()
		fis[uid] = financialInstrument{
			figiCode:     figi,
			orgID:        doubleMD5Hash(r.orgID),
			securityID:   r.securityID,
			securityName: r.securityName,
		}
	}
	return fis
}

//same as in org-transformer
func doubleMD5Hash(input string) string {
	h := md5.New()
	io.WriteString(h, input)
	return uuid.NewMD5(uuid.UUID{}, h.Sum(nil)).String()
}

func (fit *fiTransformerImpl) checkConnectivityToS3() error {
	_, err := fit.loader.BucketExists()
	if err != nil {
		return err
	}
	return nil
}
