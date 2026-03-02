# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o openmind-server ./cmd/server/main.go

# ---- Stage 2: Runtime ----
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/openmind-server .
COPY --from=builder /app/config/config.yaml ./config/config.yaml

EXPOSE 8080

ENTRYPOINT ["./openmind-server"]
