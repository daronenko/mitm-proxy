FROM golang:1.24.1-alpine AS builder

RUN adduser -u 1001 -D user

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" \
    go build -mod=vendor -o ./bin/proxy ./cmd/proxy/main.go

FROM scratch

WORKDIR /

COPY --from=builder /app/bin/proxy /proxy
COPY --from=builder /app/config/config.yaml /etc/proxy/config.yaml
COPY --from=builder /etc/passwd /etc/passwd

USER 1001

EXPOSE 8000
EXPOSE 8080

ENTRYPOINT [ "/proxy", "--config", "/etc/proxy/config.yaml" ]
