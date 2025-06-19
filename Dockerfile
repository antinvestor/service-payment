FROM golang:1.24 as builder

WORKDIR /

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the local package files to the container's workspace.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o payment_binary .

FROM alpine:3.18

# Install netcat for wait-for-db and certificates for HTTPS support
RUN apk add --no-cache netcat-openbsd

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /payment_binary /payment
#COPY --from=builder /migrations /migrations

# Copy the wait-for-db script
COPY wait-for-db.sh /wait-for-db.sh
RUN chmod +x /wait-for-db.sh

WORKDIR /

# Run the wait-for-db script before starting the service
ENTRYPOINT ["/wait-for-db.sh", "payment_db", "5432", "/payment"]

# Document the port that the service listens on by default.
EXPOSE 8020
