build:
	@go build -o bin/dfs

run:
	@go run .

test:
	@go test ./... -v