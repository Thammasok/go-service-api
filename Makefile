test:
	go test ./... -race -v

build:
	go build -o bin/app cmd/server/main.go

dev:
	go run ./cmd/server/main.go

run:
	./bin/app