.PHONY: build
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
