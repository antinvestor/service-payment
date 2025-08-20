
# Build stage
FROM golang:1.24 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Set environment variables for static binary and architecture compatibility
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the application binary
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o payment_binary ./cmd/main.go && chmod +x payment_binary

# Final stage with Distroless
FROM gcr.io/distroless/static:latest

# Install certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary to the image
COPY --from=builder /app/payment_binary /payment

# Copy the keys directory
COPY keys /keys

# Set entry point to the binary
ENTRYPOINT ["/payment"]

# Expose the application port
EXPOSE 8020
