package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

//go:embed index.html
var content embed.FS

// D3Node represents the structure D3.js expects
type D3Node struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Value    int64     `json:"value"` // Size in bytes
	Type     string    `json:"type"`  // "file" or "directory"
	Children []*D3Node `json:"children,omitempty"`
}

func main() {
	// 1. Read ncdu.json
	file, err := os.Open("ncdu.json")
	if err != nil {
		log.Fatalf("Could not open ncdu.json: %v", err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var raw []interface{}
	if err := json.Unmarshal(byteValue, &raw); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// ncdu format: [major, minor, metadata, root]
	if len(raw) < 4 {
		log.Fatal("Invalid ncdu.json format")
	}

	rootRaw := raw[3]
	
	// 2. Transform Data
	rootNode := parseNode(rootRaw, "")

	// 3. Setup Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, _ := content.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rootNode)
	})

	port := "8080"
	url := "http://localhost:" + port
	fmt.Printf("Serving at %s\n", url)

	// 4. Open Browser
	openBrowser(url)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func parseNode(raw interface{}, parentPath string) *D3Node {
	// Case 1: Directory (Array) -> [ {metadata}, child1, child2... ]
	if list, ok := raw.([]interface{}); ok && len(list) > 0 {
		meta, _ := list[0].(map[string]interface{})
		name := meta["name"].(string)
		currentPath := parentPath + "/" + name
		if parentPath == "" {
			currentPath = name // Root case
		}

		node := &D3Node{
			Name: name,
			Path: currentPath,
			Type: "directory",
		}

		// Process children (items 1 to end)
		var totalSize int64 = 0
		for _, childRaw := range list[1:] {
			childNode := parseNode(childRaw, currentPath)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				totalSize += childNode.Value
			}
		}
		node.Value = totalSize
		return node
	}

	// Case 2: File (Object) -> { "name": "foo", "asize": 123 ... }
	if meta, ok := raw.(map[string]interface{}); ok {
		name := meta["name"].(string)
		size := int64(0)
		if s, ok := meta["asize"].(float64); ok {
			size = int64(s)
		}
		
		return &D3Node{
			Name:  name,
			Path:  parentPath + "/" + name,
			Value: size,
			Type:  "file",
		}
	}

	return nil
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}
