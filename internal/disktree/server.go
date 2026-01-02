package disktree

import (
	"embed"
	"encoding/json"
	"net/http"
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
		json.NewEncoder(w).Encode(rootNode)
	}
}
