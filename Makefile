.PHONY: fmt vet build

fmt:
	go fmt ./...

vet:	fmt
	go vet ./...

build: vet
	go build -o ./bin/gogym

run: build
	./bin/gogym

test:
	go test -v ./...
