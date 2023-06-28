FROM golang:1.20 AS builder

RUN apt-get update && apt-get install -y gcc-aarch64-linux-gnu

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

RUN CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64


COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .


RUN go build -o /app/main

CMD [ "/main" ]

FROM alpine

WORKDIR /app

COPY --from=builder /app/main /app/main

RUN mkdir -p db

RUN chmod +x /app/main

CMD ["/app/main"]