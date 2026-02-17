test:
	go test ./... -race -v

build:
	go build -o bin/app cmd/server/main.go