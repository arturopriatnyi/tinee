FROM golang:1.16-alpine

RUN apk update && apk add make

WORKDIR ./urx
COPY . .

RUN go build -o ./build/urx ./cmd/urx/main.go
CMD ["./build/urx"]
