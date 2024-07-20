build:
	@go build -o bin/fs

run:
	@go run .

test:
	@go test ./...