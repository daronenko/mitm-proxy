FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" \
    go build -mod=vendor -o ./bin/mitmproxy ./cmd/mitmproxy/main.go

FROM ubuntu:latest

RUN apt-get update && apt-get install -y openssl ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /

VOLUME [ "/certs" ]

COPY --from=builder /app/bin/mitmproxy /mitmproxy
COPY --from=builder /app/config/config.yaml /etc/mitmproxy/config.yaml
COPY --from=builder /app/scripts/ /scripts/
COPY --from=builder /etc/passwd /etc/passwd

COPY /certs/ca.crt /usr/local/share/ca-certificates/ca.crt
RUN update-ca-certificates

EXPOSE 8000
EXPOSE 8080

ENTRYPOINT [ "/proxy", "--config", "/etc/mitmproxy/config.yaml" ]
