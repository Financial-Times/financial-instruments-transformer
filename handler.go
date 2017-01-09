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

type apiUrl struct {
	APIURL string `json:"apiUrl"`
}

type httpHandler struct {
	fiService fiService
	baseUrl   string
}

func (h *httpHandler) Count(w http.ResponseWriter, r *http.Request) {
	s := h.fiService

	if !s.IsInitialised() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	_, err := w.Write([]byte(strconv.Itoa(s.Count())))
	if err != nil {
		warnLogger.Printf("Could not write /count response: [%v]", err)
	}
}

func (h *httpHandler) IDs(w http.ResponseWriter, r *http.Request) {
	s := h.fiService

	if !s.IsInitialised() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	for _, uid := range s.IDs() {
		err := enc.Encode(id{ID: uid})
		if err != nil {
			warnLogger.Printf("Could not encode uid: [%s]. Err: [%v]", uid, err)
			continue
		}
	}
}

func (h *httpHandler) Read(w http.ResponseWriter, r *http.Request) {
	s := h.fiService

	if !s.IsInitialised() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	fi, present := s.Read(id)

	if !present {
		infoLogger.Printf("FI with uuid [%s] does not exist", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	uppFi := uppFI{
		UUID:      id,
		PrefLabel: fi.securityName,
		AlternativeIDs: alternativeIDs{
			UUIDs:     []string{id},
			FactsetID: fi.securityID,
			FIGI:      fi.figiCode,
		},
		IssuedBy: fi.orgID,
	}
	err := json.NewEncoder(w).Encode(uppFi)
	if err != nil {
		warnLogger.Printf("Could not return fi with uuid [%s]. Resource: [%v]. Err: [%v]", id, fi, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *httpHandler) getFinancialInstruments(w http.ResponseWriter, r *http.Request) {
	s := h.fiService

	if !s.IsInitialised() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	var apiUrls = []apiUrl{}
	for _, uuid := range s.IDs() {
		apiUrl := apiUrl{APIURL: h.baseUrl + uuid}
		apiUrls = append(apiUrls, apiUrl)
	}

	err := json.NewEncoder(w).Encode(apiUrls)

	if err != nil {
		warnLogger.Printf("Error on json encoding=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
