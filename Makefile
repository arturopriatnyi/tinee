.PHONY: build
protobuf:
	protoc -I api/pb --go_out=plugins=grpc:pkg/pb --go_opt=paths=source_relative api/pb/urx.proto

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

test:
	go test -cover -coverprofile=coverage.html -timeout 30s ./...

.PHONY: coverage
coverage:
	go tool cover -html=coverage.html
