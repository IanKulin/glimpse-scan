# Dockerfile for 'glimpse-scan'

# Stage 1: Build the Go binary
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Set environment variables for cross-compilation to AMD64/Linux
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0

# Build the Go binary
RUN go build -o glimpse-scan .

# Stage 2: Create a minimal final image
FROM busybox:latest

# Set a working directory for the final image
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/glimpse-scan .

# Ensure the data directory exists
RUN mkdir -p /app/data

# Define a bind mount for persistent data storage
VOLUME /app/data

# Run the Go binary
ENTRYPOINT ["/app/glimpse-scan"]
