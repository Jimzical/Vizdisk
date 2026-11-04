package vizdisk

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tdewolff/minify/v2"
)

func TestHandleCSS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()

	HandleCSS(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/css" {
		t.Fatalf("expected Content-Type %q, got %q", "text/css", contentType)
	}

	rawCSS, err := content.ReadFile("style.css")
	if err != nil {
		t.Fatalf("failed to read embedded style.css: %v", err)
	}

	expectedBuf := &bytes.Buffer{}
	if err := m.Minify("text/css", expectedBuf, bytes.NewReader(rawCSS)); err != nil {
		t.Fatalf("failed to minify CSS in test: %v", err)
	}

	if !bytes.Equal(rec.Body.Bytes(), expectedBuf.Bytes()) {
		t.Errorf("HandleCSS response body did not match expected minified CSS")
	}

	if rec.Body.Len() >= len(rawCSS) {
		t.Errorf("expected minified CSS length (%d) to be smaller than raw CSS (%d)", rec.Body.Len(), len(rawCSS))
	}
}

func TestHandleJS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()

	HandleJS(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/javascript" {
		t.Fatalf("expected Content-Type %q, got %q", "application/javascript", contentType)
	}

	rawJS, err := content.ReadFile("app.js")
	if err != nil {
		t.Fatalf("failed to read embedded app.js: %v", err)
	}

	expectedBuf := &bytes.Buffer{}
	if err := m.Minify("application/javascript", expectedBuf, bytes.NewReader(rawJS)); err != nil {
		t.Fatalf("failed to minify JS in test: %v", err)
	}

	if !bytes.Equal(rec.Body.Bytes(), expectedBuf.Bytes()) {
		t.Errorf("HandleJS response body did not match expected minified JS")
	}

	if rec.Body.Len() >= len(rawJS) {
		t.Errorf("expected minified JS length (%d) to be smaller than raw JS (%d)", rec.Body.Len(), len(rawJS))
	}
}

func TestHandleCSS_MinificationFallback(t *testing.T) {
	origM := m
	defer func() { m = origM }()

	m = minify.New()
	m.AddFunc("text/css", func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
		return errors.New("simulated CSS minification error")
	})

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()

	HandleCSS(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/css" {
		t.Fatalf("expected Content-Type %q, got %q", "text/css", contentType)
	}

	rawCSS, err := content.ReadFile("style.css")
	if err != nil {
		t.Fatalf("failed to read embedded style.css: %v", err)
	}

	if !bytes.Equal(rec.Body.Bytes(), rawCSS) {
		t.Errorf("HandleCSS fallback response body did not match raw unminified CSS")
	}
}

func TestHandleJS_MinificationFallback(t *testing.T) {
	origM := m
	defer func() { m = origM }()

	m = minify.New()
	m.AddFunc("application/javascript", func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
		return errors.New("simulated JS minification error")
	})

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()

	HandleJS(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/javascript" {
		t.Fatalf("expected Content-Type %q, got %q", "application/javascript", contentType)
	}

	rawJS, err := content.ReadFile("app.js")
	if err != nil {
		t.Fatalf("failed to read embedded app.js: %v", err)
	}

	if !bytes.Equal(rec.Body.Bytes(), rawJS) {
		t.Errorf("HandleJS fallback response body did not match raw unminified JS")
	}
}

type errorResponseWriter struct {
	header http.Header
}

func (e *errorResponseWriter) Header() http.Header {
	return e.header
}

func (e *errorResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("simulated write error")
}

func (e *errorResponseWriter) WriteHeader(statusCode int) {}

func TestServeStaticAsset_MissingAsset(t *testing.T) {
	rec := httptest.NewRecorder()
	serveStaticAsset(rec, "nonexistent.file", "text/plain")

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for missing file, got %d bytes", rec.Body.Len())
	}
}

func TestServeStaticAsset_WriterError(t *testing.T) {
	ew := &errorResponseWriter{header: make(http.Header)}
	// Should log error without panicking
	serveStaticAsset(ew, "style.css", "text/css")

	// Also test writer error during fallback path
	origM := m
	defer func() { m = origM }()
	m = minify.New()
	m.AddFunc("text/css", func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
		return errors.New("minification failure")
	})

	serveStaticAsset(ew, "style.css", "text/css")
}
