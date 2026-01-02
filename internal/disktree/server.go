package disktree

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed index.html
var content embed.FS

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	data, _ := content.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func HandleData(rootNode *D3Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			json.NewEncoder(gz).Encode(rootNode)
			return
		}

		json.NewEncoder(w).Encode(rootNode)
	}
}
