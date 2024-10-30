.PHONY: fmt vet build

fmt:
	go fmt ./...

vet:	fmt
	go vet ./...

build: vet
	go build -o ./bin/gogym ./cmd/api

run: build
	./bin/gogym

test:
	go test -v ./...
