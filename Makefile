build:
	@go build -o bin/ecom cmd/main.go

build-docker-compose:
	@docker-compose build

test:
	@go test -v ./...
	
run: build
	@./bin/ecom

run-docker-compose: build-docker-compose
	@docker-compose up -d

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down