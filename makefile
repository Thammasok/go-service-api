test:
	go test ./... -race -v

build:
	go build -o bin/app cmd/api/main.go