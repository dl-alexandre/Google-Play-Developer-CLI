# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
        -X github.com/google-play-cli/gpd/pkg/version.Version=${VERSION} \
        -X github.com/google-play-cli/gpd/pkg/version.GitCommit=${COMMIT} \
        -X github.com/google-play-cli/gpd/pkg/version.BuildTime=${BUILD_TIME}" \
    -o gpd ./cmd/gpd

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' gpd
USER gpd

WORKDIR /home/gpd

# Copy binary from builder
COPY --from=builder /app/gpd /usr/local/bin/gpd

ENTRYPOINT ["gpd"]
CMD ["--help"]
