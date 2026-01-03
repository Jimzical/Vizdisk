package main

import (
	"context"
	"fmt"
	"github.com/jimzical/vizdisk/internal/vizdisk"
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

	// 5. Setup Server
	http.HandleFunc("/", vizdisk.HandleIndex)
	http.HandleFunc("/style.css", vizdisk.HandleCSS)
	http.HandleFunc("/app.js", vizdisk.HandleJS)
	http.HandleFunc("/data", vizdisk.HandleData(scanDir))

	port := os.Getenv("NCDU_PORT")
	if port == "" {
		port = "8810"
	}
	url := "http://localhost:" + port
	fmt.Printf("Serving at %s\n", url)

	// 6. Open Browser
	// Only try to open browser if not running in a container (simple heuristic)
	if os.Getenv("IS_DOCKER_CONTAINER") != "true" {
		vizdisk.OpenBrowser(url)
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
