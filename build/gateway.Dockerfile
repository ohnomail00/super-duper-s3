# Build stage using Go 1.23 with CGO disabled
FROM golang:1.23 AS builder

WORKDIR /app

# Disable CGO to produce a fully static binary.
ENV CGO_ENABLED=0

# Copy dependency files and download modules.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire repository.
COPY . .

# Build the gateway binary from the cmd/gateway directory.
WORKDIR /app/cmd/gateway
RUN go build -o gateway .

# Copy the binary to a location accessible to the final stage.
RUN cp gateway /gateway-bin

# Final stage: use alpine as the runtime.
FROM alpine:latest

# Copy the gateway binary from the builder stage.
COPY --from=builder /gateway-bin /gateway-bin

# Start the gateway server.
ENTRYPOINT ["/gateway-bin"]
