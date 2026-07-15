# ---------- Builder ----------
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /app

ARG REPOSITORY
ARG VERSION=dev
ARG REVISION=none
ARG BUILDTIME

COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY ./apps/billing ./apps/billing
COPY ./apps/ledger ./apps/ledger
COPY ./apps/checkout/gen ./apps/checkout/gen
COPY ./pkg ./pkg
# billing/gen is produced by `make proto-generate-billing-collection` and checked in.


# Build static binary for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
     -ldflags="-s -w \
         -X github.com/pitabwire/frame/version.Repository=${REPOSITORY} \
         -X github.com/pitabwire/frame/version.Version=${VERSION} \
         -X github.com/pitabwire/frame/version.Commit=${REVISION} \
         -X github.com/pitabwire/frame/version.Date=${BUILDTIME}" \
     -o /app/binary ./apps/billing/cmd/main.go

# ---------- Final ----------
FROM cgr.dev/chainguard/static:latest
# Add Maintainer Info
LABEL maintainer="Bwire Peter <bwire517@gmail.com>"

USER 65532:65532

EXPOSE 80

# Add OCI metadata labels
ARG REPOSITORY
ARG VERSION
ARG REVISION
ARG BUILDTIME
LABEL org.opencontainers.image.title="Billing Service"
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.created=$BUILDTIME
LABEL org.opencontainers.image.source=$REPOSITORY

WORKDIR /

# Copy from builder stage
COPY --from=builder /app/binary /billing
COPY --from=builder /app/apps/billing/migrations /migrations

# Run the binary
ENTRYPOINT ["/billing"]
