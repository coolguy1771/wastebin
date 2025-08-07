# syntax=docker/dockerfile:1@sha256:9857836c9ee4268391bb5b09f9f157f3c91bb15821bb77969642813b0d00518d

# Build arguments for multi-platform builds
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG TARGETVARIANT=""

# ================================
# Frontend Build Stage
# ================================
FROM node:22-alpine@sha256:5539840ce9d013fa13e3b9814c9353024be7ac75aca5db6d039504a56c04ea59 AS frontend

# Install security updates and cleanup in one layer
RUN apk update && apk upgrade && apk add --no-cache \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# Create non-root user for build
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

WORKDIR /app

# Copy package files first for better layer caching
COPY web/package*.json ./

# Install dependencies with npm ci for faster, reliable builds
RUN npm ci && npm cache clean --force

# Copy source code
COPY web/ ./

# Change ownership to non-root user
RUN chown -R nextjs:nodejs /app
USER nextjs

# Build the frontend
RUN npm run build

# ================================
# Backend Build Stage  
# ================================
FROM golang:1.24-alpine@sha256:daae04ebad0c21149979cd8e9db38f565ecefd8547cf4a591240dc1972cf1399 AS backend

# Build arguments
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Set Go environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

# Install build dependencies and security updates
RUN apk update && apk upgrade && apk add --no-cache \
    ca-certificates \
    tini-static \
    gcc \
    musl-dev \
    git \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S golang && \
    adduser -S golang -u 1001

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Change ownership and switch to non-root user for build
RUN chown -R golang:golang /build
USER golang

# Build the application with optimizations
RUN go build -a \
    -tags netgo \
    -ldflags '-w -s -extldflags "-static"' \
    -trimpath \
    -o wastebin \
    ./cmd/wastebin

# ================================
# Final Runtime Stage
# ================================
FROM gcr.io/distroless/static:nonroot-amd64@sha256:3a31438b71f1e1c7d4210da01a4a5b9bc55deaba115788fdf8b2bf0269ec1d69

# Add metadata labels
LABEL \
    org.opencontainers.image.title="wastebin" \
    org.opencontainers.image.description="A fast, secure pastebin service" \
    org.opencontainers.image.source="https://github.com/coolguy1771/wastebin" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.vendor="coolguy1771" \
    maintainer="coolguy1771"

# Use non-root user
USER nonroot:nonroot

# Copy frontend build artifacts
COPY --from=frontend --chown=nonroot:nonroot /app/dist /web

# Copy backend binary and certificates
COPY --from=backend --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend --chown=nonroot:nonroot /build/wastebin /wastebin
COPY --from=backend --chown=nonroot:nonroot /sbin/tini-static /tini

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/wastebin", "health"] || exit 1

# Expose port (for documentation)
EXPOSE 3000

# Use tini as init system
ENTRYPOINT ["/tini", "--"]
CMD ["/wastebin"]
