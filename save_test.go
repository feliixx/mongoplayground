package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestSave(t *testing.T) {

	defer testServer.clearDatabases(t)

	saveTests := []struct {
		name      string
		params    url.Values
		result    string
		newRecord bool
		mode      byte
	}{
		{
			name:      "template config",
			params:    templateParams,
			result:    templateURL,
			newRecord: true,
			mode:      mgodatagenMode,
		},
		{
			name:      "template config with new query",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {"db.collection.find({\"k\": 10})"}},
			result:    "p/DYlGRQeO0bX",
			newRecord: true,
			mode:      mgodatagenMode,
		},
		{
			name:      "invalid config",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/EMmfQADkGcq",
			newRecord: true,
			mode:      mgodatagenMode,
		},
		{
			name:      "save existing playground",
			params:    templateParams,
			result:    templateURL,
			newRecord: false,
		},
		{
			name:      "template query with new config",
			params:    url.Values{"mode": {"bson"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/4cOeA7NGLru",
			newRecord: true,
			mode:      bsonSingleCollection,
		},
		{
			name:      "bson mutliple db",
			params:    url.Values{"mode": {"bson"}, "config": {`db={"c1":[{k:1}],"c2":[]}`}, "query": {templateQuery}},
			result:    "p/AHWNKW4GK50",
			newRecord: true,
			mode:      bsonMultipleCollection,
		},
		{
			name:      "bson unknown",
			params:    url.Values{"mode": {"bson"}, "config": {`unknown`}, "query": {templateQuery}},
			result:    "p/qXPeSYPpatw",
			newRecord: true,
			mode:      unknown,
		},
	}

	t.Run("parallel save", func(t *testing.T) {
		for _, tt := range saveTests {

			test := tt // capture range variable
			t.Run(tt.name, func(t *testing.T) {

				t.Parallel()

				buf := httpBody(t, testServer.saveHandler, http.MethodPost, saveEndpoint, test.params)

				if want, got := test.result, buf.String(); want != got {
					t.Errorf("expected %s, but got %s", want, got)
				}
			})
		}
	})

	nbMgoDatagen, nbBsonSingle, nbBsonMultiple, nbUnknown := 0, 0, 0, 0
	// save without run should not create any database
	nbMongoDatabases, nbBadgerRecords := 0, 0
	for _, tt := range saveTests {
		if tt.newRecord {
			nbBadgerRecords++
			switch tt.mode {
			case mgodatagenMode:
				nbMgoDatagen++
			case bsonSingleCollection:
				nbBsonSingle++
			case bsonMultipleCollection:
				nbBsonMultiple++
			case unknown:
				nbUnknown++
			}
		}
	}
	testStorageContent(t, nbMongoDatabases, nbBadgerRecords)

	testPlaygroundStats(t, nbMgoDatagen, nbBsonSingle, nbBsonMultiple, nbUnknown)
}

func testPlaygroundStats(t *testing.T, nbMgoDatagen, nbBsonSingle, nbBsonMultiple, nbUnknown int) {

	// reset saved playground metrics
	savedPlayground.Reset()
	err := testServer.computeSavedPlaygroundStats()
	if err != nil {
		t.Errorf("fail to get stats from saved playgrounds")
	}
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, metricsEndpoint, nil)
	promhttp.Handler().ServeHTTP(resp, req)

	want := fmt.Sprintf(`saved_playground_count{type="bson_multiple_collection"} %d
saved_playground_count{type="bson_single_collection"} %d
saved_playground_count{type="mgodatagen"} %d
saved_playground_count{type="unknown"} %d`, nbBsonMultiple, nbBsonSingle, nbMgoDatagen, nbUnknown)

	lines := make([]string, 0, 4)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "saved_playground_count") {
			lines = append(lines, scanner.Text())
		}
	}
	if got := strings.Join(lines, "\n"); want != got {
		t.Errorf("expected %s\n but got\n %s", want, got)
	}

}

func TestErrorOnSavePlaygroundTooBig(t *testing.T) {

	params := url.Values{
		"mode":   {"mgodatagen"},
		"config": {string(make([]byte, maxByteSize))},
		"query":  {"db.collection.find()"},
	}

	buf := httpBody(t, testServer.saveHandler, http.MethodPost, saveEndpoint, params)
	if want, got := errPlaygroundToBig, buf.String(); want != got {
		t.Errorf("expected %s, but got %s", want, got)
	}

}
