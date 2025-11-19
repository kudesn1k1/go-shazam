# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/main.go

FROM debian:bookworm-slim

# Install minimal Python + pip + ffmpeg
RUN apt-get update && \
	apt-get install -y --no-install-recommends \
	python3 \
	python3-pip \
	ffmpeg \
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

EXPOSE 5000

USER nonroot:nonroot

ENTRYPOINT ["/server"]