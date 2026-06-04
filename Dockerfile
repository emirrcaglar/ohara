# syntax=docker/dockerfile:1

FROM node:22-bookworm-slim AS frontend-builder
WORKDIR /src

COPY src/frontend/package*.json ./frontend/
RUN npm --prefix frontend ci

COPY src/frontend ./frontend
RUN npm --prefix frontend run build:embed \
    && test -f /src/backend/ui/dist/index.html

FROM golang:1.24-bookworm AS backend-builder
WORKDIR /src/backend

COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download

COPY src/backend ./
COPY --from=frontend-builder /src/backend/ui/dist ./ui/dist
RUN test -f ./ui/dist/index.html

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/ohara \
    ./cmd

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S ohara \
    && adduser -S -D -H -h /var/lib/ohara -s /sbin/nologin -G ohara ohara \
    && mkdir -p /var/lib/ohara /var/cache/ohara /etc/ohara \
    && chown -R ohara:ohara /var/lib/ohara /var/cache/ohara /etc/ohara

COPY --from=backend-builder /out/ohara /usr/local/bin/ohara

USER ohara
WORKDIR /var/lib/ohara
VOLUME ["/var/lib/ohara"]
EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/ohara"]
