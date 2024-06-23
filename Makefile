build:
	@go build -o bin/dfs

run:
	@./bin/dfs

test:
	@go test ./... -v