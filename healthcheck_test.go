package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthcheck(t *testing.T) {
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/_status/healthcheck", nil)

	testServer.healthcheckHandler(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("expected response code %v, got %v", http.StatusOK, resp.Code)
	}
	if want, got := string(statusOK), resp.Body.String(); want != got {
		t.Errorf("expected response %s, but got %s", want, got)
	}
}

func TestHealthcheckServerError(t *testing.T) {

	testServer.session.SetSocketTimeout(1 * time.Microsecond)
	defer testServer.session.SetSocketTimeout(100 * time.Millisecond)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/_status/healthcheck", nil)

	testServer.healthcheckHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected response code %v, got %v", http.StatusOK, resp.Code)
	}

	want := `{"status":"unexpected result:`
	got := resp.Body.String()

	if !strings.HasPrefix(got, want) {
		t.Errorf("expected response to start with %s, but got %s", want, got)
	}
}
