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

# Build the application binary from the main package
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o jenga_binary ./cmd/main.go && chmod +x jenga_binary

# Final stage with Distroless
FROM gcr.io/distroless/static:latest

# Install certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary to the image
COPY --from=builder /app/jenga_binary /jenga

# Copy the keys directory
COPY keys /keys

# Set default environment variables for Kubernetes compatibility
ENV LOG_LEVEL=info
ENV SECURELY_RUN_SERVICE=false
ENV JENGA_PRIVATE_KEY_PATH=/app/keys/privatekey.pem
ENV JENGA_API_KEY=SZq0WmmtX6mfo3fARW7yHeEzhfs3sOiEj2TgS2jb9gFz80JPfvTF1g4nr1uziA1meg3uFB1/Cm+ZXdTDob4z0Q==
ENV JENGA_CONSUMER_SECRET=JZkt2pAIiS4F4RP4x6zQ97f1dn9j1N
ENV JENGA_MERCHANT_CODE=8503993262
ENV JENGA_CALLBACK_URL=http://localhost:8080/callback
ENV JENGA_ENV=https://uat.finserve.africa
ENV PROFILE_SERVICE_URI=127.0.0.1:7005
ENV PARTITION_SERVICE_URI=127.0.0.1:7003
ENV PAYMENT_SERVICE_URI=localhost:50051
ENV NATS_URL=nats://ant:secret@nats-server:4222?subject=
ENV DATABASE_URL=postgres://ant:secret@payment_db:5432/service_payment?sslmode=disable

# Set entry point to the binary
ENTRYPOINT ["/jenga"]

# Expose the application port
EXPOSE 80
