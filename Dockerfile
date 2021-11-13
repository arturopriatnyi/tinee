FROM golang:1.16-alpine

RUN apk update && apk add make

WORKDIR ./zen
COPY . .

RUN make build
