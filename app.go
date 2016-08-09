package main

import (
	"bufio"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/pborman/uuid"
	"github.com/rlmcpherson/s3gof3r"
)

const bbgIDs = "edm_bbg_ids.txt"
const securityEntityMap = "edm_security_entity_map.txt"

type figiCodeToSecurityID map[string][]string
type securityIDtoRawFinancialInstruments map[string][]rawFinancialInstrument

type s3Config struct {
	accKey    string
	secretKey string
	bucket    string
	domain    string
}

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

type financialInstrument struct {
	figiCode   string
	securityID string
	//UPP UUID
	orgID string
}

type fiHandler struct {
	//uuid to financial instrument mapping
	financialInstruments map[string]financialInstrument
}

func init() {
	initLogs(os.Stdout, os.Stdout, os.Stderr)
}

func main() {
	app := cli.App("financial instrument transformer", "Transforms factset data into Financial Instruments")

	awsAccessKey := app.String(cli.StringOpt{
		Name:   "aws-access-key-id",
		Desc:   "s3 access key",
		EnvVar: "AWS_ACCESS_KEY_ID",
	})
	awsSecretKey := app.String(cli.StringOpt{
		Name:   "aws-secret-access-key",
		Desc:   "s3 secret key",
		EnvVar: "AWS_SECRET_ACCESS_KEY",
	})
	bucketName := app.String(cli.StringOpt{
		Name:   "bucket-name",
		Desc:   "bucket name of factset data",
		EnvVar: "BUCKET_NAME",
	})
	bucketDomain := app.String(cli.StringOpt{
		Name:   "bucket-domain",
		Value:  "s3-eu-west-1.amazonaws.com",
		Desc:   "domain of factset bucket",
		EnvVar: "BUCKET_DOMAIN",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "application port",
		EnvVar: "PORT",
	})

	app.Action = func() {
		s3 := s3Config{
			accKey:    *awsAccessKey,
			secretKey: *awsSecretKey,
			bucket:    *bucketName,
			domain:    *bucketDomain,
		}
		infoLogger.Printf("Config: [bucket: %s] [domain: %s]", s3.bucket, s3.domain)

		fih := &fiHandler{}
		go func(fih *fiHandler) {
			infoLogger.Println("Started loading FIs.")
			start := time.Now()
			fis, err := loadFIs(s3)
			if err != nil {
				errorLogger.Println(err)
				return
			}
			fih.financialInstruments = fis

			infoLogger.Printf("Loading FIs finished in [%v]", time.Since(start))
			infoLogger.Printf("Nr of FIs: [%v]", len(fis))

		}(fih)
		listen(fih, *port)
	}

	err := app.Run(os.Args)
	if err != nil {
		errorLogger.Printf("[%v]", err)
	}
}

func listen(fih *fiHandler, port int) {
	infoLogger.Println("Listening on port:", port)
	r := mux.NewRouter()
	r.HandleFunc("/transformers/financialinstruments/__count", fih.count).Methods("GET")
	r.HandleFunc("/transformers/financialinstruments/__ids", fih.ids).Methods("GET")
	r.HandleFunc("/transformers/financialinstruments/{id}", fih.id).Methods("GET")
	r.HandleFunc("/__health", fih.health()).Methods("GET")
	r.HandleFunc("/__gtg", fih.gtg).Methods("GET")
	err := http.ListenAndServe(":"+strconv.Itoa(port), r)
	if err != nil {
		errorLogger.Println(err)
	}
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

func fetchFIGICodes(r io.Reader) figiCodeToSecurityID {
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
					orgID:      r.orgID,
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
