build:
	@go build -o bin/dfs

run:
	@.go run ./bin/sfs

test:
	@go test ./... -v