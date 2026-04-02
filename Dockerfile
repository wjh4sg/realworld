FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/apiserver ./apiserver

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=builder /out/apiserver /app/apiserver
COPY config.yaml /app/config.yaml

ENV CONFIG_PATH=/app/config.yaml

EXPOSE 8080

USER app

ENTRYPOINT ["/app/apiserver"]
