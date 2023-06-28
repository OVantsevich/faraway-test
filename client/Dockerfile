FROM golang:1.20 AS build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /app/main

CMD [ "/main" ]

FROM alpine

WORKDIR /app

COPY --from=build /app/main /app/main

CMD ["/app/main"]