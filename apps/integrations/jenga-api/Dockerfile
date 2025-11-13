
FROM golang:1.25 AS builder

# Set working directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY ./apps/integrations/jenga-api ./apps/integrations/jenga-api
#COPY ./internal ./internal

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' -o binary ./apps/integrations/jenga-api/cmd/main.go

# Final minimal image
FROM gcr.io/distroless/static:nonroot

# Copy from builder stage
COPY --from=builder /app/binary /jenga-api

# Copy the keys directory
COPY --chown=65532:65532 keys /keys

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
ENV DATABASE_URL=postgres://ant:secret@payment_db:5432/service_payment?sslmode=disable


# Run as non-root user
USER 65532:65532

# Document ports
EXPOSE 80
EXPOSE 50051

# Set working directory
WORKDIR /

# Run the binary
ENTRYPOINT ["/jenga-api"]

