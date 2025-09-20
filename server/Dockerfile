# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache ffmpeg && \
	addgroup -S nonroot && \
	adduser -S nonroot -G nonroot

WORKDIR /app

COPY --from=builder /server /server

EXPOSE 5000

USER nonroot:nonroot

ENTRYPOINT ["/server"]