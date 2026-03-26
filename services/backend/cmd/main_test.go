package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"question-voting-app/internal/handlers"
	"question-voting-app/internal/testutil"
	"strings"
	"testing"
)

func TestSetupRouter_CORS(t *testing.T) {
	storer := testutil.NewMockStorer()
	api := handlers.New(storer, false, nil)
	mux := SetupRouter(api, "http://test-origin.com")

	req := httptest.NewRequest(http.MethodOptions, "/api/session", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "http://test-origin.com" {
		t.Errorf("Expected CORS origin 'http://test-origin.com', got %q", origin)
	}
	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected OPTIONS request to return 200 OK, got %d", w.Code)
	}
}

func TestSetupRouter_Routes(t *testing.T) {
	storer := testutil.NewMockStorer()
	api := handlers.New(storer, false, nil)
	mux := SetupRouter(api, "*")

	tests := []struct {
		name   string
		method string
		path   string
		expect int
	}{
		{"Create Session", http.MethodPost, "/api/session", http.StatusCreated},
		{"Root Wrong Method", http.MethodGet, "/api/session", http.StatusMethodNotAllowed},
		{"Get Session (Not Found Session)", http.MethodGet, "/api/session/123", http.StatusNotFound},
		{"End Session (No Content/Not Found silently succeeds)", http.MethodDelete, "/api/session/123", http.StatusNoContent},
		{"Delete Question (Not Found Session)", http.MethodDelete, "/api/session/123/questions/456", http.StatusNotFound},
		{"Check Admin", http.MethodGet, "/api/session/123/check-admin", http.StatusOK},
		{"Unknown Route", http.MethodPatch, "/api/session/123/unknown", http.StatusNotFound},
		{"Missing Handler Path", http.MethodGet, "/api/session/123/invalid-suffix", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.name == "Create Session" {
				body = strings.NewReader("{}")
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expect {
				t.Errorf("Expected %d, got %d", tt.expect, w.Code)
			}
		})
	}
}
