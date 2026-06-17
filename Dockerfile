# Stage 1: Build
FROM docker.io/library/golang:latest AS builder

WORKDIR /workspace

ARG ARCH
ARG goproxy=https://proxy.golang.org
ENV GOPROXY=${goproxy}

# Install swag for Swagger doc gen
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod and sum files
COPY main-app/go.mod main-app/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy full source
COPY main-app/ .

# Generate Swagger docs
RUN swag init --generalInfo main.go --output docs

# Build the binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} \
    go build -ldflags "-s -w -extldflags '-static'" \
    -o design-carousel-service

# Stage 2: Run
FROM docker.io/library/ubuntu:latest

# Install CA certs for TLS (required to talk to GitHub)
RUN apt-get update && apt-get install -y ca-certificates curl wget jq && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /workspace/design-carousel-service .
COPY --from=builder /workspace/preload_images /app/preload_images
RUN mkdir -p /data/carousel_images \
  && chown -R 65532:65532 /app /data

# Non-root for K8s security policies
USER 65532

ENV CAROUSEL_DATA_DIR=/data \
    CAROUSEL_PRELOAD_DIR=/app/preload_images
EXPOSE 10022 9222

ENTRYPOINT ["/app/design-carousel-service"]

