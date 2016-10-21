package main

import (
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"net/http"
)

func (h *httpHandler) amazonS3Healthcheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to read the latest dataset from S3",
		Name:             "Check connectivity to Amazon S3",
		PanicGuide:       "TODO",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Amazon S3 bucket to read the latest Factset dataset",
		Checker:          h.checkConnectivityToS3,
	}
}

func (h *httpHandler) checkConnectivityToS3() (string, error) {
	err := h.fiService.checkConnectivity()
	if err != nil {
		return fmt.Sprintf("Healthcheck: Unable to connect to Amazon S3: %v", err.Error()), err
	}
	return "", nil
}

func (h *httpHandler) goodToGo(w http.ResponseWriter, r *http.Request) {
	if _, err := h.checkConnectivityToS3(); err != nil {
		errorLogger.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
}
