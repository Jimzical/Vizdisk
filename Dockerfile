# Build stage
FROM golang:alpine AS builder

WORKDIR /app
COPY . .
# Build the binary. We disable CGO for a static binary, though not strictly necessary for Alpine->Alpine
RUN CGO_ENABLED=0 go build -o disktree main.go

# Final stage
FROM alpine:latest

# Install ncdu
RUN apk add --no-cache ncdu

# Set environment variable to prevent browser opening
ENV IS_DOCKER_CONTAINER=true
ENV NCDU_PORT=8810

WORKDIR /app
COPY --from=builder /app/disktree .

# Create a directory for mounting the volume to scan
WORKDIR /scan

# Expose the port
EXPOSE 8810

# Run the application
ENTRYPOINT ["/app/disktree", "."]

