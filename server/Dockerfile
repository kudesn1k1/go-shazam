# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /worker ./cmd/worker/main.go

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm-slim

# Install minimal Python + pip + ffmpeg + bash + coreutils (for timeout)
RUN apt-get update && \
	apt-get install -y --no-install-recommends \
	python3 \
	python3-pip \
	ffmpeg \
	bash \
	coreutils \
	ca-certificates && \
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
COPY docker-entrypoint.sh /docker-entrypoint.sh
COPY migrations ./migrations

# Make entrypoint executable
RUN chmod +x /docker-entrypoint.sh

EXPOSE 5000

USER nonroot:nonroot

ENTRYPOINT ["/docker-entrypoint.sh"]
