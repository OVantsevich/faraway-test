ARG GOLANG_VERSION=1.20-alpine

FROM golang:${GOLANG_VERSION} AS build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app
COPY . .
RUN go mod download

RUN go build -o /app/main

FROM alpine

WORKDIR /app

COPY --from=builder /app/main ./main

CMD ["/app/main"]