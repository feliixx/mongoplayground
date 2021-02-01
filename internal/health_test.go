package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestHealthCheck(t *testing.T) {

	GitCommit = "af36b1ee99f0709d751fe7e70493b4e103560b2a"
	GitBranch = "dev"
	BuildDate = "2021-01-24T10:59:00"

	buf := httpBody(t, testServer.healthHandler, http.MethodGet, healthEndpoint, url.Values{})

	want := fmt.Sprintf(`{"Status":"UP","Services":[{"Name":"badger","Status":"UP"},{"Name":"mongodb","Version":"%s","Status":"UP"}],"BuildInfo":{"Commit":"af36b1ee99f0709d751fe7e70493b4e103560b2a","Branch":"dev","BuildDate":"2021-01-24T10:59:00"}}`, testServer.mongodbVersion)
	if got := buf.String(); want != got {
		t.Errorf("expected\n%s\nbut got\n%s", want, got)
	}
}
