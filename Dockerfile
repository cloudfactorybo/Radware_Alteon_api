# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/alteon-api    ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/alteon-admin  ./cmd/admin

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S alteon && adduser -S -G alteon alteon

COPY --from=build /out/alteon-api   /usr/local/bin/alteon-api
COPY --from=build /out/alteon-admin /usr/local/bin/alteon-admin

USER alteon
EXPOSE 8080
ENTRYPOINT ["alteon-api"]
