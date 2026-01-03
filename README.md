# VizDisk

VizDisk is a lightweight tool that visualizes your disk usage as an interactive tree map. It uses `ncdu` to scan directories efficiently and serves a D3.js visualization in your browser.

## Why Vizdisk?

I wanted something like WinDirStat / WizTree, but with a modern interface and the ability to run it on a headless server. It’s basically a quick "what’s eating my disk?" dashboard for a homeserver that I can open in a browser.

## Features

*   **Fast Scanning**: Leverages `ncdu` for efficient disk usage analysis.
*   **Interactive UI**: Explore directories and files using a D3.js tree visualization.
*   **Docker Support**: Run it anywhere without installing Go or ncdu on your host.
*   **Single Binary**: Compiles to a static binary with embedded assets.

## Screenshots

#### Map View
<img width="1920" height="1020" alt="VizDisk map view showing disk usage as a colored treemap of directories and files" src="https://github.com/user-attachments/assets/3f099747-02de-40fe-bbd1-138ba10edfb7" />
<img width="1920" height="1020" alt="VizDisk map view zoomed into a selected directory within the treemap visualization" src="https://github.com/user-attachments/assets/eed0200b-a66a-4120-9a6d-9c19ace95ede" />

#### List View
<img width="1920" height="1020" alt="VizDisk list view displaying directories and files with sizes and usage details" src="https://github.com/user-attachments/assets/5277526c-80e1-40db-8d8d-8bff547a7910" />
<img width="1920" height="1020" alt="VizDisk list view with expanded directory details for disk usage analysis" src="https://github.com/user-attachments/assets/defc315f-86b4-4558-86b7-d717155a69e2" />

## Prerequisites

### For Local Execution
*   **Go** (1.24+)
*   **ncdu**
    *   Ubuntu/Debian: `sudo apt install ncdu`
    *   macOS: `brew install ncdu`

### For Docker Execution
*   **Docker**

## Quickstart (Docker)

VizDisk scans whatever is mounted at `/scan` inside the container.

### Docker Compose

Create a `compose.yml`:

```yaml
services:
    vizdisk:
        image: ghcr.io/jimzical/vizdisk:latest
        ports:
            - "8810:8810"
        volumes:
            # Mount the directory you want to scan to /scan in the container.
            # Example: scan your whole machine (read-only):
            - "/:/scan:ro"
            # Example: scan only your home folder (read-only):
            # - "/home:/scan:ro"
            # Example (macOS):
            # - "/Users:/scan:ro"
        environment:
            # Optional: change the HTTP port the server listens on.
            # If you change this, also update the host-side port mapping above.
            NCDU_PORT: "8810"
```

Start it:
```bash
docker compose -f compose.yml up
```

Open `http://localhost:8810`.

To scan a different folder, edit the volume mapping in `compose.yml` (mount your target directory to `/scan:ro`).

### Docker (one-liner)

Scan your whole machine (read-only):
```bash
docker run --rm -p 8810:8810 -v /:/scan:ro ghcr.io/jimzical/vizdisk:latest
```

Scan only your home directory (read-only):
```bash
docker run --rm -p 8810:8810 -v $HOME:/scan:ro ghcr.io/jimzical/vizdisk:latest
```

## Installation

### Option 1: Download Binary
Download the latest binary for your OS from the [Releases Page](https://github.com/jimzical/vizdisk/releases).

> Note: The downloaded binary still requires `ncdu` to be installed (see [Prerequisites → For Local Execution](#for-local-execution)).

### Option 2: Build from Source
```bash
git clone https://github.com/jimzical/vizdisk.git
cd vizdisk
go build -o vizdisk main.go
```

## Usage
### Running Locally
```bash
# Scan the current directory
./vizdisk

# Or scan a specific directory
./vizdisk /path/to/scan
```

The browser should open automatically at `http://localhost:8810`.

### Building a Docker Image Locally

```bash
docker build -t vizdisk:local .
docker run --rm -p 8810:8810 -v $(pwd):/scan:ro vizdisk:local
```
## Configuration

You can configure the port using an environment variable:

*   `NCDU_PORT`: Sets the HTTP server port (default: 8810).

Example:
```bash
export NCDU_PORT=9090
go run main.go
```

## TODOs
- [ ] Update screenshots in README.md
- [ ] Add basic authentication to restrict access
- [ ] Add ability to restrict to localhost only (bind host option)
- [ ] Add filtering options (e.g., exclude certain file types or directories)
- [ ] Add option to export reports (e.g., CSV, JSON)
- [ ] Add a scan progress indicator / "scanning..." status
- [ ] Cache scan results (avoid re-scanning on every page refresh)
- [ ] Add an option to rescan on demand