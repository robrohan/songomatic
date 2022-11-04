.PHONY: build clean test

hash = $(shell git log --pretty=format:'%h' -n 1)

include .env
export

DOCKER_CONTAINER=legendary-doodle

# List all targets in thie file
list:
	@echo ""
	@echo "*^-. Go App Template .-^*"
	@echo ""
	@grep -B 1 '^[^#[:space:]].*:' Makefile
	@echo ""

# Run all go unit tests
test:
	go test ./...

# Runs the localhost server
start:
	go run cmd/server/main.go

clean:
	rm -rf build

build: clean
	mkdir -p build
	go build -o build/server -ldflags "-X main.build=${hash}" cmd/server/main.go
	cp -R static build/
	cp -R templates build/
	cp -R migrations build/

build.docker: build
	docker build -t $(DOCKER_CONTAINER) .

#	Using a different enviroment variable set for prod
start.docker:
	docker run --env-file=.env.production -p 8080:3000 $(DOCKER_CONTAINER)

#	Grab a base css that styles form elements with some basic style
fetch_base_css:
	curl https://raw.githubusercontent.com/robrohan/pho-ui/main/src/pho-ui.css > templates/pho-ui.css

start_db:
	docker run --name postgres \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-e PGDATA=$(DB_PATH) \
		-p 5432:5432 \
		-d postgres:latest
