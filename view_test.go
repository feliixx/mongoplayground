package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestView(t *testing.T) {

	defer testServer.clearDatabases(t)

	viewTests := []struct {
		name         string
		params       url.Values
		url          string
		responseCode int
	}{
		// {
		// 	name:         "template parameters",
		// 	params:       templateParams,
		// 	url:          templateURL,
		// 	responseCode: http.StatusOK,
		// },
		{
			name: "new config",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"_id": 1}]`},
				"query":  {templateQuery},
			},
			url:          "p/DEz-pkpheLX",
			responseCode: http.StatusOK,
		},
		{
			name:         "non existing url",
			params:       templateParams,
			url:          "p/unknownURL",
			responseCode: http.StatusNotFound,
		},
	}

	// start by saving all needed playground
	for _, tt := range viewTests {
		httpBody(t, testServer.saveHandler, http.MethodPost, saveEndpoint, tt.params)
	}

	t.Run("parallel view", func(t *testing.T) {
		for _, tt := range viewTests {

			test := tt // capture range variable
			t.Run(test.name, func(t *testing.T) {

				t.Parallel()

				req, _ := http.NewRequest(http.MethodGet, "/"+test.url, nil)
				resp := httptest.NewRecorder()
				testServer.viewHandler(resp, req)

				if test.responseCode != resp.Code {
					t.Errorf("expected response code %d but got %d", test.responseCode, resp.Code)
				}
			})
		}
	})
}
