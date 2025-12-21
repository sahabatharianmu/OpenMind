FROM oven/bun:1.0 AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/bun.lockb ./
RUN bun install --frozen-lockfile
COPY web/ .
RUN bun run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git make
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o openmind-server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=backend-builder /app/openmind-server .
COPY --from=backend-builder /app/config/config.yaml ./config/config.yaml

COPY --from=frontend-builder /app/web/dist ./web/dist
COPY --from=frontend-builder /app/web/public ./web/public

EXPOSE 8080

ENTRYPOINT ["./openmind-server"]
