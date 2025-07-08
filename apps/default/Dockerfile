FROM golang:1.24 as builder

WORKDIR /

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the local package files to the container's workspace.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o payment_binary .

FROM gcr.io/distroless/static:latest

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /payment_binary /payment
#COPY --from=builder /migrations /migrations

WORKDIR /

# Run the payment binary
ENTRYPOINT ["/payment"]

# Document the port that the service listens on by default.
EXPOSE 8020
