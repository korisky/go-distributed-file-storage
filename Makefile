build:
	@go build -o bin/fs

run:
	@./bin/fs

test:
	@go test ./...