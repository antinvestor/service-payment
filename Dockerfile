FROM golang:1.23 as builder

WORKDIR /

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the local package files to the container's workspace.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o payment_binary .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /payment_binary /payment
#COPY --from=builder /migrations /migrations

WORKDIR /

# Run the service command by default when the container starts.
ENTRYPOINT ["/payment"]

# Document the port that the service listens on by default.
EXPOSE 8020