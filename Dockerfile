# syntax=docker/dockerfile:1.7

# --- build stage ---------------------------------------------------------
FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION}" \
    -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w" \
    -o /out/importobsidian ./cmd/importobsidian

# --- runtime stage -------------------------------------------------------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app

WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/importobsidian /app/importobsidian
COPY web /app/web

RUN mkdir -p /app/storage /app/tokens && chown -R app:app /app
VOLUME ["/app/storage", "/app/tokens"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1

USER app

ENTRYPOINT ["/app/server"]
