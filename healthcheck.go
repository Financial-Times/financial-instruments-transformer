package main

import (
	"net/http"

	"github.com/Financial-Times/go-fthealth"
)

func (s *fiService) health() func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler("FinancialInstrumentsTransformer", "Financial Instrument Transformer healthcheck")
}

func (s *fiService) gtg(w http.ResponseWriter, r *http.Request) {
	if s.financialInstruments == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
}
