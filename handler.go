package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type id struct {
	ID string `json:"id"`
}

type uppFI struct {
	UUID           string         `json:"uuid"`
	PrefLabel      string         `json:"prefLabel"`
	AlternativeIDs alternativeIDs `json:"alternativeIdentifiers"`
	IssuedBy       string         `json:"issuedBy"`
}

type alternativeIDs struct {
	UUIDs     []string `json:"uuids"`
	FactsetID string   `json:"factsetIdentifier"`
	FIGI      string   `json:"figiCode"`
}

func (fi *fiHandler) count(w http.ResponseWriter, r *http.Request) {
	if fi.financialInstruments == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	_, err := w.Write([]byte(strconv.Itoa(len(fi.financialInstruments))))
	if err != nil {
		warnLogger.Printf("Could not write /count response: [%v]", err)
	}
}

func (fi *fiHandler) ids(w http.ResponseWriter, r *http.Request) {
	if fi.financialInstruments == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	for uid := range fi.financialInstruments {
		err := enc.Encode(id{uid})
		if err != nil {
			warnLogger.Printf("Could not encode uid: [%s]. Err: [%v]", uid, err)
			continue
		}
	}
}

func (fi *fiHandler) id(w http.ResponseWriter, r *http.Request) {
	if fi.financialInstruments == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	resource, present := fi.financialInstruments[id]
	if !present {
		infoLogger.Printf("FI with uuid [%s] does not exist", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	uppFI := uppFI{
		UUID:      id,
		PrefLabel: "Equity",
		AlternativeIDs: alternativeIDs{
			UUIDs:     []string{id},
			FactsetID: resource.securityID,
			FIGI:      resource.figiCode,
		},
		IssuedBy: resource.orgID,
	}
	err := json.NewEncoder(w).Encode(uppFI)
	if err != nil {
		warnLogger.Printf("Could not return fi with uuid [%s]. Resource: [%v]. Err: [%v]", id, resource, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
