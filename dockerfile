# syntax=docker/dockerfile:1

FROM golang:1.17.6-alpine

WORKDIR /app 

RUN apk add build-base

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY .hash ./

RUN go build -o /app/scbcb

EXPOSE 8080

CMD [ "/app/scbcb" ]
