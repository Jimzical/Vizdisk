package disktree

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

//go:embed index.html style.css app.js
var content embed.FS

var m *minify.M

func init() {
	m = minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	data, _ := content.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func HandleCSS(w http.ResponseWriter, r *http.Request) {
	data, _ := content.ReadFile("style.css")
	w.Header().Set("Content-Type", "text/css")

	mw := &bytes.Buffer{}
	if err := m.Minify("text/css", mw, bytes.NewReader(data)); err == nil {
		w.Write(mw.Bytes())
	} else {
		w.Write(data)
	}
}

func HandleJS(w http.ResponseWriter, r *http.Request) {
	data, _ := content.ReadFile("app.js")
	w.Header().Set("Content-Type", "application/javascript")

	mw := &bytes.Buffer{}
	if err := m.Minify("application/javascript", mw, bytes.NewReader(data)); err == nil {
		w.Write(mw.Bytes())
	} else {
		w.Write(data)
	}
}

func HandleData(rootNode *D3Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Requested-With") != "DiskTreeApp" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "text/plain")

		var output io.Writer = w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			output = gz
		}

		b64Encoder := base64.NewEncoder(base64.StdEncoding, output)
		defer b64Encoder.Close()

		json.NewEncoder(b64Encoder).Encode(rootNode)
	}
}
