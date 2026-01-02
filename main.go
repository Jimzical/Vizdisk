package main

import (
	"context"
	"disktree/internal/disktree"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Determine Scan Directory, default to current directory
	scanDir := "."
	if len(os.Args) > 1 {
		scanDir = os.Args[1]
	}

	// Check if ncdu is installed
	if _, err := exec.LookPath("ncdu"); err != nil {
		log.Fatal("Error: 'ncdu' command not found. Please install it (e.g., sudo apt install ncdu) or ensure it's in your PATH.")
	}

	fmt.Printf("Scanning '%s' with ncdu... (this may take a moment)\n", scanDir)

	// Using CommandContext so the scan can be interrupted
	cmd := exec.CommandContext(ctx, "ncdu", "-o", "-", "-x", "--exclude-kernfs", scanDir)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.Canceled {
			fmt.Println("\nScan cancelled by user.")
			return
		}
		log.Fatalf("Error running ncdu: %v", err)
	}

	fmt.Println("Scan complete. Parsing data...")

	var raw []any
	if err := json.Unmarshal(output, &raw); err != nil {
		log.Fatalf("Error parsing JSON output from ncdu: %v", err)
	}

	// ncdu format: [major, minor, metadata, root]
	if len(raw) < 4 {
		log.Fatal("Invalid ncdu output format")
	}

	rootRaw := raw[3]

	// 4. Transform Data
	rootNode := disktree.ParseNode(rootRaw, "")

	// 5. Setup Server
	http.HandleFunc("/", disktree.HandleIndex)
	http.HandleFunc("/data", disktree.HandleData(rootNode))

	port := os.Getenv("NCDU_PORT")
	if port == "" {
		port = "8810"
	}
	url := "http://localhost:" + port
	fmt.Printf("Serving at %s\n", url)

	// 6. Open Browser
	// Only try to open browser if not running in a container (simple heuristic)
	if os.Getenv("IS_DOCKER_CONTAINER") != "true" {
		disktree.OpenBrowser(url)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // Uses DefaultServeMux
	}

	// Run server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	fmt.Println("\nShutting down server...")

	// Create a deadline to wait for active requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting")
}
