FROM golang:1.24-alpine AS builder
ARG GITHUB_ACCESS_TOKEN
WORKDIR /app
COPY . .
RUN apk add --no-cache ca-certificates
RUN update-ca-certificates

ENV GOCACHE=/root/.cache/go-build

ENV SQLITE_URL=file:./db.sqlite
ENV KV_SQLITE_URL=file:./kv.sqlite
WORKDIR /app
RUN --mount=type=cache,target="/root/.cache/go-build" go run ./cmd/scan_service/main.go --config ./cmd/scan_service/config.yaml
RUN CGO_ENABLED=0 GOOS=linux go build -o backline ./cmd/webapp/main.go

FROM scratch

COPY --from=builder /app/backline /
COPY --from=builder /app/cmd/webapp/config.yaml /
COPY --from=builder /app/db.sqlite /
COPY --from=builder /app/kv.sqlite /
COPY --from=builder /bluge-index /bluge-index
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/backline", "--config", "/config.yaml"]