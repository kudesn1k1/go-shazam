# Stage 1: build the SPA so we can embed it into the Go binary
FROM node:24-alpine AS client-builder

WORKDIR /client

COPY client/package.json client/package-lock.json ./
RUN npm ci

COPY client/ .
RUN npm run build

# Stage 2: build Go binaries
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .

# Replace the dev stub at internal/web/dist with the freshly built SPA so
# go:embed picks up the real bundle.
RUN rm -rf ./internal/web/dist
COPY --from=client-builder /client/dist ./internal/web/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /worker ./cmd/worker/main.go

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Stage 3: minimal runtime
FROM debian:bookworm-slim

# Install minimal Python + pip + ffmpeg + bash + coreutils (for timeout)
RUN apt-get update && \
	apt-get install -y --no-install-recommends \
	python3 \
	python3-pip \
	ffmpeg \
	bash \
	coreutils \
	ca-certificates \
	wget && \
	# Install yt-dlp via pip (clean, no cache)
	pip3 install --no-cache-dir --break-system-packages yt-dlp && \
	# Clean up
	apt-get purge -y python3-pip && \
	rm -rf /var/lib/apt/lists/* /root/.cache/pip/*

# Create nonroot user/group
RUN groupadd -r nonroot && useradd -r -g nonroot nonroot

WORKDIR /app

COPY --from=builder /server /server
COPY --from=builder /worker /worker

# Copy goose binary (from builder's GOPATH)
COPY --from=builder /go/bin/goose /usr/local/bin/goose
RUN chmod +x /usr/local/bin/goose

# Copy entrypoint script and migrations
COPY server/docker-entrypoint.sh /docker-entrypoint.sh
COPY server/migrations ./migrations

# Make entrypoint executable
RUN chmod +x /docker-entrypoint.sh

EXPOSE 5000

USER nonroot:nonroot

ENTRYPOINT ["/docker-entrypoint.sh"]
