# DiskTree

DiskTree is a lightweight, interactive disk usage visualizer written in Go. It wraps the powerful `ncdu` utility to scan your directories and presents the results in a beautiful, zoomable treemap directly in your web browser.

## Features

- **Automated Scanning**: Automatically runs `ncdu` to scan directories—no manual export steps required.
- **Interactive Treemap**: Visualizes disk usage using D3.js.
  - **Zoomable**: Click on folders to zoom in, click the breadcrumbs to zoom out.
  - **Tooltips**: Hover over blocks to see detailed file information and sizes.
  - **Smart Visibility**: Automatically handles tiny files to ensure the visualization remains readable.
- **Single Binary**: The HTML frontend is embedded into the Go binary, making it a single, portable executable.
- **Cross-Platform Support**: Automatically opens your default browser on Linux, macOS, and Windows.

## Prerequisites

1. **Go**: You need Go installed to build or run the project.
2. **ncdu**: This tool relies on `ncdu` for efficient disk scanning.
   - **Linux (Debian/Ubuntu)**: `sudo apt install ncdu`
   - **macOS**: `brew install ncdu`
   - **Windows**: Ensure `ncdu` is installed and available in your system PATH.

## Installation

Clone the repository and build the binary:

```bash
# Clone the repo (if applicable)
# git clone ...

# Navigate to the directory
cd DiskTree

# Build the binary
go build -o disktree
```

## Usage

You can run the tool directly with Go, or use the built binary.

### Scan Current Directory
By default, DiskTree scans the current working directory.

```bash
./disktree
# OR
go run main.go
```

### Scan Specific Directory
Pass the directory path as an argument.

```bash
./disktree /home/user/Documents
# OR
go run main.go /home/user/Documents
```

## How It Works

1. The Go program executes `ncdu -o - [directory]` to generate a JSON stream of the directory structure.
2. It parses the JSON output into a tree structure compatible with D3.js.
3. It starts a local HTTP server (default port 8080) and serves the embedded `index.html`.
4. It automatically opens your default web browser to the visualization.

