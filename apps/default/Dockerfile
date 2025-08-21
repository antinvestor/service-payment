
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

# Set default environment variables for Kubernetes compatibility
ENV LOG_LEVEL=info
ENV SECURELY_RUN_SERVICE=false
ENV NATS_URL=nats://ant:secret@nats-server:4222?subject=
ENV PROMPT_TOPIC=initiate.prompt
ENV PAYMENT_LINK_TOPIC=create.payment.link
ENV DO_MIGRATION=false
ENV PROFILE_SERVICE_URI=127.0.0.1:7005
ENV PARTITION_SERVICE_URI=127.0.0.1:7003
ENV LEDGER_SERVICE_URI=127.0.0.1:7004

# Set entry point to the binary
ENTRYPOINT ["/payment"]

# Expose the application port
EXPOSE 80
