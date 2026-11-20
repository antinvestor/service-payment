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
COPY ./pkg ./pkg
COPY ./apps/integrations/jenga-api ./apps/integrations/jenga-api
# Build static binary for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
     -ldflags="-s -w \
         -X github.com/pitabwire/frame/v3/version.Repository=${REPOSITORY} \
         -X github.com/pitabwire/frame/v3/version.Version=${VERSION} \
         -X github.com/pitabwire/frame/v3/version.Commit=${REVISION} \
         -X github.com/pitabwire/frame/v3/version.Date=${BUILDTIME}" \
     -o /app/binary ./apps/integrations/jenga-api/cmd/main.go

# ---------- Final ----------
FROM cgr.dev/chainguard/static:latest
LABEL maintainer="Bwire Peter <bwire517@gmail.com>"

USER 65532:65532

EXPOSE 80

# Add OCI metadata labels
ARG REPOSITORY
ARG VERSION
ARG REVISION
ARG BUILDTIME

LABEL org.opencontainers.image.title="Payments Jenga API Service"
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.created=$BUILDTIME
LABEL org.opencontainers.image.source=$REPOSITORY

WORKDIR /

# Copy from builder stage
COPY --from=builder /app/binary /jenga-api

# The RSA signing key is never baked into the image: supply it through the
# JENGA_PRIVATE_KEY_PEM secret or mount a file and point JENGA_PRIVATE_KEY_PATH at it.

ENV LOG_LEVEL=info

# Run the binary
ENTRYPOINT ["/jenga-api"]
