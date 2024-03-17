# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.21-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

LABEL maintainer="Amangeldy Rakhimzhanov <amangeldyjk72@gmail.com>"

WORKDIR /var/www

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/api

EXPOSE 4000

CMD ["./main"]