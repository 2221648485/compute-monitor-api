# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=${GOPROXY} \
    GOSUMDB=${GOSUMDB}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/compute-monitor-api ./cmd/server

FROM alpine:3.22

ENV TZ=Asia/Shanghai \
    APP_ENV=prod

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/compute-monitor-api /app/compute-monitor-api
COPY configs/config.*.yaml /app/configs/

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/compute-monitor-api"]
