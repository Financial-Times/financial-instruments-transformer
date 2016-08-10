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

type fiService struct {
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
	s3Domain := app.String(cli.StringOpt{
		Name:   "s3-domain",
		Value:  "s3-eu-west-1.amazonaws.com",
		Desc:   "s3 domain of factset bucket",
		EnvVar: "S3_DOMAIN",
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
			domain:    *s3Domain,
		}
		infoLogger.Printf("Config: [bucket: %s] [domain: %s]", s3.bucket, s3.domain)

		s := &fiService{}
		go func(s *fiService) {
			infoLogger.Println("Started loading FIs.")
			start := time.Now()
			fis, err := loadFIs(s3)
			if err != nil {
				errorLogger.Println(err)
				return
			}
			s.financialInstruments = fis

			infoLogger.Printf("Loading FIs finished in [%v]", time.Since(start))
			infoLogger.Printf("Nr of FIs: [%v]", len(fis))

		}(s)
		listen(s, *port)
	}

	err := app.Run(os.Args)
	if err != nil {
		errorLogger.Printf("[%v]", err)
	}
}

func listen(s *fiService, port int) {
	infoLogger.Println("Listening on port:", port)
	r := mux.NewRouter()
	r.HandleFunc("/transformers/financialinstruments/__count", s.count).Methods("GET")
	r.HandleFunc("/transformers/financialinstruments/__ids", s.ids).Methods("GET")
	r.HandleFunc("/transformers/financialinstruments/{id}", s.id).Methods("GET")
	r.HandleFunc("/__health", s.health()).Methods("GET")
	r.HandleFunc("/__gtg", s.gtg).Methods("GET")
	err := http.ListenAndServe(":"+strconv.Itoa(port), r)
	if err != nil {
		errorLogger.Println(err)
	}
}
