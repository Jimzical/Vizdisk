package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetupRouter_Endpoints(t *testing.T) {
	mux := setupRouter(t.TempDir())

	tests := []struct {
		name           string
		method         string
		target         string
		headers        map[string]string
		wantStatusCode int
		wantHeaderSub  map[string]string
	}{
		{
			name:           "index route",
			method:         http.MethodGet,
			target:         "/",
			wantStatusCode: http.StatusOK,
			wantHeaderSub:  map[string]string{"Content-Type": "text/html"},
		},
		{
			name:           "css route",
			method:         http.MethodGet,
			target:         "/style.css",
			wantStatusCode: http.StatusOK,
			wantHeaderSub:  map[string]string{"Content-Type": "text/css"},
		},
		{
			name:           "js route",
			method:         http.MethodGet,
			target:         "/app.js",
			wantStatusCode: http.StatusOK,
			wantHeaderSub:  map[string]string{"Content-Type": "application/javascript"},
		},
		{
			name:           "data route forbidden without header",
			method:         http.MethodGet,
			target:         "/data",
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:   "data route routed with header",
			method: http.MethodGet,
			target: "/data",
			headers: map[string]string{
				"X-Requested-With": "DiskTreeApp",
			},
			// If ncdu runs successfully, returns 200 OK.
			// If ncdu fails or is missing, returns 500 InternalServerError.
			// In either case, it routes to HandleData and is not 404 StatusNotFound.
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if tt.name == "data route routed with header" {
				if rec.Code == http.StatusNotFound {
					t.Fatalf("expected route /data to be handled, got 404 Not Found")
				}
				if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
					t.Fatalf("expected status 200 or 500 for /data handler, got %d", rec.Code)
				}
				return
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			for headerKey, substr := range tt.wantHeaderSub {
				gotHeader := rec.Header().Get(headerKey)
				if !strings.Contains(gotHeader, substr) {
					t.Errorf("header %s: got %q, expected to contain %q", headerKey, gotHeader, substr)
				}
			}

			body := rec.Body.String()
			if len(body) == 0 {
				t.Errorf("expected non-empty response body for %s", tt.target)
			}
		})
	}
}

func TestSetupRouter_DoesNotModifyGlobalDefaultServeMux(t *testing.T) {
	// Call setupRouter to initialize an isolated router instance.
	_ = setupRouter(t.TempDir())

	routes := []string{"/", "/style.css", "/app.js", "/data"}

	for _, route := range routes {
		t.Run("default servemux: "+route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			rec := httptest.NewRecorder()

			http.DefaultServeMux.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("expected global http.DefaultServeMux to return 404 for route %q, got %d", route, rec.Code)
			}
		})
	}
}

func TestSetupServer_HandlerRouting(t *testing.T) {
	srv := setupServer(t.TempDir(), "8810")
	if srv.Handler == nil {
		t.Fatal("expected srv.Handler to be non-nil")
	}

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	client := ts.Client()

	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if !strings.Contains(string(body), "<!DOCTYPE html>") && !strings.Contains(string(body), "<html") {
		t.Errorf("expected HTML body, got: %s", string(body))
	}
}
