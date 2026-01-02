# VizDisk

VizDisk is a lightweight tool that visualizes your disk usage as an interactive tree map. It uses `ncdu` to scan directories efficiently and serves a D3.js visualization in your browser.

## Features

*   **Fast Scanning**: Leverages `ncdu` for efficient disk usage analysis.
*   **Interactive UI**: Explore directories and files using a D3.js tree visualization.
*   **Docker Support**: Run it anywhere without installing Go or ncdu on your host.
*   **Single Binary**: Compiles to a static binary with embedded assets.

## Prerequisites

### For Local Execution
*   **Go** (1.16+)
*   **ncdu** (must be installed and in your PATH)
    *   Ubuntu/Debian: `sudo apt install ncdu`
    *   macOS: `brew install ncdu`

### For Docker Execution
*   **Docker**

## Getting Started

### Running Locally

1.  Clone the repository:
    ```bash
    git clone https://github.com/jimzical/vizdisk.git
    cd vizdisk
    ```

2.  Run the application:
    ```bash
    # Scan the current directory
    go run main.go

    # Or scan a specific directory
    go run main.go /path/to/scan
    ```

3.  The browser should open automatically at `http://localhost:8810`.

### Building the Binary

1.  Build the static binary:
    ```bash
    go build -o disktree main.go
    ```

2.  Run the binary:
    ```bash
    # Scan the current directory
    ./disktree

    # Or scan a specific directory
    ./disktree /path/to/scan
    ```

### Running with Docker

1.  Build the image:
    ```bash
    docker build -t disktree .
    ```

2.  Run the container:
    **Note:** You must mount the directory you want to scan to `/scan` inside the container.

    ```bash
    # Scan the current directory
    docker run -p 8810:8810 -v $(pwd):/scan disktree

    # Scan a specific path (e.g., your home directory)
    docker run -p 8810:8810 -v /home/user:/scan disktree
    ```

3.  Open `http://localhost:8810` in your browser.

## Configuration

You can configure the port using an environment variable:

*   `NCDU_PORT`: Sets the HTTP server port (default: 8810).

Example:
```bash
export NCDU_PORT=9090
go run main.go
```