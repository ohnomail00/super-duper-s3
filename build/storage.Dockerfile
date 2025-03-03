# Build stage using Go 1.23 with CGO disabled
FROM golang:1.23 AS builder

WORKDIR /app

# Disable CGO for a fully static build.
ENV CGO_ENABLED=0

# Copy dependency files and download modules.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire repository.
COPY . .

# Build the storage binary from the cmd/storage directory.
WORKDIR /app/cmd/storage
RUN go build -o storage .

# Create a dedicated storage directory and set permissions.
RUN mkdir -p /storage

# Copy the storage binary to a location for final copying.
RUN cp storage /storage-bin

# Final stage: use alpine as the runtime.
FROM alpine:latest

# Copy the storage directory and binary from the builder stage.
COPY --from=builder /storage /storage
COPY --from=builder /storage-bin /storage-bin

# Start the storage server.
ENTRYPOINT ["/storage-bin"]
