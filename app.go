package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
)

type s3Config struct {
	accKey    string
	secretKey string
	bucket    string
	domain    string
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
