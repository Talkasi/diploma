# Сборка бинарника
FROM golang:1.21-alpine AS dev
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM golang:1.21-alpine AS builder
WORKDIR /build
RUN apk add --no-cache git ca-certificates
COPY --from=dev /src /build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o api ./cmd/api && go build -o importer ./cmd/importer

# Минимальный runtime-образ
FROM alpine:3.19
WORKDIR /app
# Устанавливаем корневые сертификаты и создаём непривилегированного пользователя
RUN apk add --no-cache ca-certificates \
    && addgroup -S app && adduser -S app -G app
COPY --from=builder /build/api ./api
COPY --from=builder /build/importer ./importer
USER app
EXPOSE 8080
ENTRYPOINT ["./api"]
