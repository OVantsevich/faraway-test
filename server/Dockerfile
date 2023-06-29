FROM golang:1.20-alpine AS builder

ENV CGO_ENABLED=1 \
    GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

RUN apk add --no-cache --update go gcc g++

WORKDIR /app

COPY . .

WORKDIR /app/internal/ent
RUN go generate

WORKDIR /app
RUN go build -o /app/main

FROM alpine

WORKDIR /app

RUN mkdir db

COPY --from=builder /app/main .

EXPOSE 12345

CMD ["/app/main"]