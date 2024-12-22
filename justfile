# migrate-db creates an sql file in ./migrations directory where migration queries should be written.
migrate-db FILENAME:
    goose -dir ./migrations create {{FILENAME}} sql

start:
	docker compose up -d

stop:
	docker compose down

build:
	docker compose build

rebuild: stop build start

test:
	go test ./...