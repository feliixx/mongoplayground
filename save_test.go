package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSave(t *testing.T) {

	testServer.clearDatabases(t)

	saveTests := []struct {
		name      string
		params    url.Values
		result    string
		newRecord bool
	}{
		{
			name:      "template config",
			params:    templateParams,
			result:    templateURL,
			newRecord: true,
		},
		{
			name:      "template config with new query",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.collection.find({\"k\": 10})"}},
			result:    "p/DYlGRQeO0bX",
			newRecord: true,
		},
		{
			name:      "invalid config",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/EMmfQADkGcq",
			newRecord: true,
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
		},
	}

	nbBadgerRecords := 0
	for _, tt := range saveTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := httpBody(t, testServer.saveHandler, http.MethodPost, "/save", tt.params)

			if want, got := tt.result, buf.String(); want != got {
				t.Errorf("expected %s, but got %s", want, got)
			}
		})
		if tt.newRecord {
			nbBadgerRecords++
		}
	}

	testStorageContent(t, 0, nbBadgerRecords)

}
