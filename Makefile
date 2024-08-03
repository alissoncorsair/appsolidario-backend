build:
	@go build -o bin/main cmd/main.go

build-docker-compose:
	@docker-compose build

test:
	@go test -v ./...
	
run: build
	@./bin/main

run-hotreload:
	@air -c .air.toml

run-docker-compose: build-docker-compose
	@docker-compose up -d

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down


# build:
# 	@docker compose build

# build-docker-compose:
# 	@docker compose build

# test:
# 	@docker compose run --rm app go test -v ./...

# run: build
# 	@docker compose up -d

# run-docker-compose: build-docker-compose
# 	@docker compose up -d

# migration:
# 	@docker compose run --rm app migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

# migrate-up:
# 	@docker compose run --rm app go run cmd/migrate/main.go up

# migrate-down:
# 	@docker compose run --rm app go run cmd/migrate/main.go down
