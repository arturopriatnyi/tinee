.PHONY: build
protobuf:
	protoc -I api/pb --go_out=plugins=grpc:pkg/pb --go_opt=paths=source_relative api/pb/urx.proto

build:
	go build -o ./build/urx ./cmd/urx/main.go

run:
	go run ./cmd/urx/main.go

start:
	./build/urx

test:
	go test -cover -coverprofile=coverage.html -timeout 30s ./...

.PHONY: coverage
coverage:
	go tool cover -html=coverage.html
