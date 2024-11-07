# Build stage
FROM golang:1.23 as builder

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
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o jenga_binary . && chmod +x jenga_binary

# Final stage with Alpine
FROM alpine:3.18

# Install certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary to the Alpine image
COPY --from=builder /app/jenga_binary /jenga

# Set entry point to the binary
ENTRYPOINT ["/jenga"]

# Expose the application port
EXPOSE 8080
