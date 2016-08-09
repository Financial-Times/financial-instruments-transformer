package main

import (
	"net/http"

	"github.com/Financial-Times/go-fthealth"
)

func (fih *fiHandler) health() func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler("FinancialInstrumentsTransformer", "Financial Instrument Transformer healthcheck")
}

func (fih *fiHandler) gtg(w http.ResponseWriter, r *http.Request) {
	if fih.financialInstruments == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
}
