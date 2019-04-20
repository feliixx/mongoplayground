package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestView(t *testing.T) {

	testServer.clearDatabases(t)

	viewTests := []struct {
		name         string
		params       url.Values
		url          string
		responseCode int
		newRecord    bool
	}{
		{
			name:         "template parameters",
			params:       templateParams,
			url:          templateURL,
			responseCode: http.StatusOK,
			newRecord:    true,
		},
		{
			name: "new config",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"_id": 1}]`},
				"query":  {templateQuery},
			},
			url:          "p/DEz-pkpheLX",
			responseCode: http.StatusOK,
			newRecord:    true,
		},
		{
			name:         "non existing url",
			params:       templateParams,
			url:          "p/random",
			responseCode: http.StatusNotFound,
			newRecord:    false,
		},
	}

	nbBadgerRecords := 0
	for _, tt := range viewTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.responseCode == http.StatusOK {
				buf := httpBody(t, testServer.saveHandler, http.MethodPost, "/save", tt.params)

				if want, got := tt.url, buf.String(); want != got {
					t.Errorf("expected %s but got %s", want, got)
				}
			}
			req, _ := http.NewRequest(http.MethodGet, "/"+tt.url, nil)
			resp := httptest.NewRecorder()
			testServer.viewHandler(resp, req)

			if tt.responseCode != resp.Code {
				t.Errorf("expected response code %d but got %d", tt.responseCode, resp.Code)
			}
		})
		if tt.newRecord {
			nbBadgerRecords++
		}
	}

	testStorageContent(t, 0, nbBadgerRecords)

}
