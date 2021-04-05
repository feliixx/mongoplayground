// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2017 Adrien Petel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

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

	defer clearDatabases(t)

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
			name:      "bson multiple db",
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

				got := httpBody(t, saveEndpoint, http.MethodPost, test.params)

				if want := test.result; want != got {
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
	savedPlaygroundSize.Reset()
	computeSavedPlaygroundStats(testStorage.kvStore)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, metricsEndpoint, nil)
	promhttp.Handler().ServeHTTP(resp, req)

	want := fmt.Sprintf(`saved_playground_size_count{type="bson_multiple_collection"} %d
saved_playground_size_count{type="bson_single_collection"} %d
saved_playground_size_count{type="mgodatagen"} %d
saved_playground_size_count{type="unknown"} %d`, nbBsonMultiple, nbBsonSingle, nbMgoDatagen, nbUnknown)

	lines := make([]string, 0, 4)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "saved_playground_size_count") {
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

	want := errPlaygroundToBig
	got := httpBody(t, saveEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s, but got %s", want, got)
	}

}
