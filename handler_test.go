package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCount_FinancialInstrumentsMapIsNil_ServiceUnavailableStatusCode(t *testing.T) {
	fi := fiHandler{}

	req, err := http.NewRequest("GET", "http://fiTransformer/__count", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	fi.count(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

func TestCount_FinancialInstrumentsMapIsNotNil_OkStatusCodeAndResponseBodyShowsCorrectNrOfItems(t *testing.T) {
	var testCases = []struct {
		fiMap map[string]financialInstrument
		count string
	}{
		{
			fiMap: make(map[string]financialInstrument),
			count: "0",
		},
		{
			fiMap: map[string]financialInstrument{"foo": financialInstrument{}},
			count: "1",
		},
		{
			fiMap: map[string]financialInstrument{"foo": financialInstrument{}, "bar": financialInstrument{}},
			count: "2",
		},
	}

	req, err := http.NewRequest("GET", "http://fiTransformer/__count", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	for _, tc := range testCases {
		fi := fiHandler{tc.fiMap}
		w := httptest.NewRecorder()
		fi.count(w, req)
		if w.Code != 200 {
			t.Errorf("Expected statusCode [%d]. Actual: [%d]", 200, w.Code)
		}
		resp := w.Body.String()
		if resp != tc.count {
			t.Errorf("Expected resp [%s]. Actual: [%s]", tc.count, resp)
		}
	}
}

func TestIds_FinancialInstrumentsMapIsNil_ServiceUnavailableStatusCode(t *testing.T) {
	fi := fiHandler{}

	req, err := http.NewRequest("GET", "http://fiTransformer/__ids", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	fi.ids(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

func TestIds_FinancialInstrumentsMapIsNotNil_IdsInStreamingJsonFormatReturned(t *testing.T) {
	var testCases = []struct {
		fiMap    map[string]financialInstrument
		response string
	}{
		{
			fiMap:    make(map[string]financialInstrument),
			response: "",
		},
		{
			fiMap:    map[string]financialInstrument{"foo": financialInstrument{}},
			response: `{"id":"foo"}` + "\n",
		},
		{
			fiMap:    map[string]financialInstrument{"foo": financialInstrument{}, "bar": financialInstrument{}},
			response: `{"id":"foo"}` + "\n" + `{"id":"bar"}` + "\n",
		},
	}

	req, err := http.NewRequest("GET", "http://fiTransformer/__ids", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	for _, tc := range testCases {
		fi := fiHandler{tc.fiMap}
		w := httptest.NewRecorder()
		fi.ids(w, req)
		if w.Code != 200 {
			t.Errorf("Expected statusCode [%d]. Actual: [%d]", 200, w.Code)
		}
		resp := w.Body.String()
		if resp != tc.response {
			t.Errorf("Expected resp [%s]. Actual: [%s]", tc.response, resp)
		}
	}
}

func TestId_FinancialInstrumentsMapIsNil_StatusServiceUnavailable(t *testing.T) {
	fi := fiHandler{}

	req, err := http.NewRequest("GET", "http://fiTransformer/{id}", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	fi.id(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

func TestId_RequestedFinancialInstrumentDoesNotExist_StatusNotFound(t *testing.T) {
	fi := fiHandler{
		map[string]financialInstrument{
			"foo": financialInstrument{},
		},
	}
	r := mux.NewRouter()
	r.HandleFunc("/{id}", fi.id)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/bar")
	if err != nil {
		t.Fatalf("Failure: [%v]", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 404, resp.StatusCode)
	}
}

func TestId_FinancialInstrumentExists_OkStatusAndCorrectFIReturned(t *testing.T) {
	fi := fiHandler{
		map[string]financialInstrument{
			"foo": financialInstrument{
				figiCode:   "BBG01234",
				securityID: "TVKI-123",
				orgID:      "012AF-E",
			},
		},
	}
	r := mux.NewRouter()
	r.HandleFunc("/{id}", fi.id)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/foo")
	if err != nil {
		t.Fatalf("Failure: [%v]", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 200, resp.StatusCode)
	}
	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failure: [%v]", err)
	}
	expected := `{"uuid":"foo","prefLabel":"Equity","alternativeIdentifiers":{"uuids":["foo"],"factsetIdentifier":"TVKI-123","figiCode":"BBG01234"},"issuedBy":"012AF-E"}` + "\n"
	actual := string(rBody)

	if actual != expected {
		t.Errorf("Expected: [%s]. Actual: [%s]", expected, actual)
	}
}
