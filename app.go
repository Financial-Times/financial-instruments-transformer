package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
)

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
		Value:  "s3.amazonaws.com",
		Desc:   "s3 domain of factset bucket",
		EnvVar: "S3_DOMAIN",
	})
	baseURL := app.String(cli.StringOpt{
		Name:   "base-url",
		Value:  "http://localhost:8080/transformers/financial-instruments/",
		Desc:   "Base url",
		EnvVar: "BASE_URL",
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

		s3Loader, err := news3Loader(s3)
		if err != nil {
			errorLogger.Printf("[%v]", err)
		}

		fiParser := fiParserImpl{}
		fit := fiTransformerImpl{
			loader: &s3Loader,
			parser: &fiParser,
		}
		fis := fiServiceImpl{
			fit:     &fit,
			config:  s3,
			baseUrl: *baseURL,
		}
		go func() {
			fis.Init()
		}()

		httpHandler := &httpHandler{fiService: &fis}
		listen(httpHandler, *port)
	}

	err := app.Run(os.Args)
	if err != nil {
		errorLogger.Printf("[%v]", err)
	}
}

func listen(h *httpHandler, port int) {
	infoLogger.Println("Listening on port:", port)
	r := mux.NewRouter()
	r.HandleFunc("/transformers/financial-instruments/__count", h.Count).Methods("GET")
	r.HandleFunc("/transformers/financial-instruments/__ids", h.IDs).Methods("GET")
	r.HandleFunc("/transformers/financial-instruments", h.getFinancialInstruments).Methods("GET")
	r.HandleFunc("/transformers/financial-instruments/{id}", h.Read).Methods("GET")
	r.HandleFunc("/__health", v1a.Handler("Financial Instruments Transformer Healthchecks", "Checks for accessing Amazon S3 bucket", h.amazonS3Healthcheck()))
	r.HandleFunc("/__gtg", h.goodToGo)
	err := http.ListenAndServe(":"+strconv.Itoa(port), r)
	if err != nil {
		errorLogger.Println(err)
	}
}
