FROM golang:1.20 AS builder

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY . .

RUN go build -o /app/main

FROM alpine

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]