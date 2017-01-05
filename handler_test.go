package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCount_FinancialInstrumentsMapIsNil_ServiceUnavailableStatusCode(t *testing.T) {
	fis := &fiServiceImpl{}
	h := httpHandler{fiService: fis}

	req, err := http.NewRequest("GET", "http://fiTransformer/__count", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	h.Count(w, req)

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
			fiMap: map[string]financialInstrument{"foo": {}},
			count: "1",
		},
		{
			fiMap: map[string]financialInstrument{"foo": {}, "bar": {}},
			count: "2",
		},
	}

	req, err := http.NewRequest("GET", "http://fiTransformer/__count", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	for _, tc := range testCases {
		fi := &fiServiceImpl{financialInstruments: tc.fiMap}
		h := httpHandler{fiService: fi}
		w := httptest.NewRecorder()
		h.Count(w, req)
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
	fis := &fiServiceImpl{}
	h := httpHandler{fiService: fis}

	req, err := http.NewRequest("GET", "http://fiTransformer/__ids", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	h.IDs(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

func TestIds_FinancialInstrumentsMapIsNotNil_IdsInStreamingJsonFormatAreReturned(t *testing.T) {
	var testCases = []struct {
		fiMap    map[string]financialInstrument
		expected []string
	}{
		{
			fiMap:    make(map[string]financialInstrument),
			expected: []string{""},
		},
		{
			fiMap:    map[string]financialInstrument{"foo": {}},
			expected: []string{`{"id":"foo"}` + "\n"},
		},
		{
			fiMap: map[string]financialInstrument{"foo": {}, "bar": {}},
			//map order is random, hence the multiple valid responses
			expected: []string{
				`{"id":"foo"}` + "\n" + `{"id":"bar"}` + "\n",
				`{"id":"bar"}` + "\n" + `{"id":"foo"}` + "\n",
			},
		},
	}

	req, err := http.NewRequest("GET", "http://fiTransformer/__ids", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	for _, tc := range testCases {
		fi := &fiServiceImpl{financialInstruments: tc.fiMap}
		h := httpHandler{fiService: fi}
		w := httptest.NewRecorder()
		h.IDs(w, req)
		if w.Code != 200 {
			t.Errorf("Expected statusCode [%d]. Actual: [%d]", 200, w.Code)
		}
		actual := w.Body.String()
		if !equals(actual, tc.expected) {
			t.Errorf("Expected resp [%s]. Actual: [%s]", tc.expected, actual)
		}
	}
}

func TestRead_FinancialInstrumentsMapIsNil_StatusServiceUnavailable(t *testing.T) {
	fis := &fiServiceImpl{}
	h := httpHandler{fiService: fis}

	req, err := http.NewRequest("GET", "http://fiTransformer/{id}", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	h.Read(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

// mux package doesn't provide a way to mock path params, therefore we have to set up a test server with a router

func TestId_RequestedFinancialInstrumentDoesNotExist_StatusNotFound(t *testing.T) {
	s := &fiServiceImpl{
		financialInstruments: map[string]financialInstrument{
			"foo": {},
		},
	}
	h := httpHandler{fiService: s}

	r := mux.NewRouter()
	r.HandleFunc("/{id}", h.Read)

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
	s := &fiServiceImpl{
		financialInstruments: map[string]financialInstrument{
			"foo": {
				figiCode:     "BBG01234",
				securityID:   "TVKI-123",
				orgID:        "012AF-E",
				securityName: "LIG SPECIAL PURPOSE ACQ 2ND CO  ORD",
			},
		},
	}
	r := mux.NewRouter()
	h := httpHandler{fiService: s}

	r.HandleFunc("/{id}", h.Read)

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
	expected := `{"uuid":"foo","prefLabel":"LIG SPECIAL PURPOSE ACQ 2ND CO  ORD","alternativeIdentifiers":{"uuids":["foo"],"factsetIdentifier":"TVKI-123","figiCode":"BBG01234"},"issuedBy":"012AF-E"}` + "\n"
	actual := string(rBody)

	if actual != expected {
		t.Errorf("Expected: [%s]. Actual: [%s]", expected, actual)
	}
}

func TestGetFinancialInstruments_FinancialInstrumentsMapIsNil_ServiceUnavailableStatusCode(t *testing.T) {
	fis := &fiServiceImpl{}
	h := httpHandler{fiService: fis}

	req, err := http.NewRequest("GET", "http://fiTransformer", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	w := httptest.NewRecorder()

	h.getFinancialInstruments(w, req)

	if w.Code != 503 {
		t.Errorf("Expected: [%d]. Actual: [%d]", 503, w.Code)
	}
}

func TestGetFinancialInstruments_FinancialInstrumentsMapIsNotNil_ApiIdsInStreamingJsonFormatAreReturned(t *testing.T) {
	baseUrl := "fiAppURL/"
	uuid1 := "foo"
	uuid2 := "bar"
	var testCases = []struct {
		fiMap    map[string]financialInstrument
		expected []string
	}{
		{
			fiMap:    make(map[string]financialInstrument),
			expected: []string{""},
		},
		{
			fiMap:    map[string]financialInstrument{uuid1: {}},
			expected: []string{`{"apiUrl":"` + baseUrl + uuid1 + `"}` + "\n"},
		},
		{
			fiMap: map[string]financialInstrument{uuid1: {}, uuid2: {}},
			//map order is random, hence the multiple valid responses
			expected: []string{
				`{"apiUrl":"` + baseUrl + uuid1 + `"}` +  "\n" + `{"apiUrl":"` + baseUrl + uuid2 + `"}` + "\n",
				`{"apiUrl":"` + baseUrl + uuid2 + `"}` + "\n" + `{"apiUrl":"` + baseUrl + uuid1 + `"}` + "\n",
			},
		},
	}

	req, err := http.NewRequest("GET", "http://fiTransformer", nil)
	if err != nil {
		t.Fatalf("Failure in setting up the test request: [%v]", err)
	}

	for _, tc := range testCases {
		fi := &fiServiceImpl{financialInstruments: tc.fiMap, baseUrl: baseUrl}
		h := httpHandler{fiService: fi}
		w := httptest.NewRecorder()
		h.getFinancialInstruments(w, req)
		if w.Code != 200 {
			t.Errorf("Expected statusCode [%d]. Actual: [%d]", 200, w.Code)
		}
		actual := w.Body.String()
		if !equals(actual, tc.expected) {
			t.Errorf("Expected resp [%s]. Actual: [%s]", tc.expected, actual)
		}
	}
}

func equals(actual string, expected []string) bool {
	for _, e := range expected {
		if e == actual {
			return true
		}
	}
	return false
}
