package main

import (
	"net/http"

	"github.com/Financial-Times/go-fthealth"
)

func (h *httpHandler) health() func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler("FinancialInstrumentsTransformer", "Financial Instrument Transformer healthcheck")
}

func (h *httpHandler) gtg() func(w http.ResponseWriter, r *http.Request) {
	return fthealth.Handler("FinancialInstrumentsTransformer", "Financial Instrument Transformer healthcheck")
}
