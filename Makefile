build:
	@go build -o bin/alem ./cmd/alem/main.go

run: build
	@./bin/alem --config=./config/local.yaml

test:
	@go test -v ./...

migrate:
	@go run ./cmd/migrate/main.go --config=./config/local.yaml --migrations-path=./migrations
