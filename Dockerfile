FROM golang:1.16-alpine

RUN apk update && apk add make

WORKDIR ./tinee
COPY . .

RUN go build -o ./build/tinee ./cmd/tinee/main.go
CMD ["./build/tinee"]
