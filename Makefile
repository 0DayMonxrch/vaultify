.PHONY: dev test lint migrate sqlc build

dev:
	docker-compose up -d

test:
	go test ./...

lint:
	golangci-lint run

migrate:
	golang-migrate -path ./db/migrations -database "$$DATABASE_URL" up

sqlc:
	sqlc generate

build:
	go build -o server.exe ./cmd/server
	go build -o vaultify.exe ./cmd/vaultify
