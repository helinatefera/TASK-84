.PHONY: build run test migrate-up migrate-down migrate-create tidy certs

DB_DSN ?= "root:password@tcp(localhost:3306)/local_insights?parseTime=true&charset=utf8mb4&loc=UTC"

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

tidy:
	go mod tidy

certs:
	mkdir -p certs
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout certs/server.key -out certs/server.crt \
		-subj "/C=US/ST=Local/L=Local/O=LocalInsights/CN=localhost"

migrate-up:
	go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path ./migrations -database "mysql://$(DB_DSN)" up

migrate-down:
	go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path ./migrations -database "mysql://$(DB_DSN)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		create -ext sql -dir ./migrations -seq $$name

test:
	go test ./... -v -count=1
