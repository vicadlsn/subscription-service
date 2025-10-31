APP_NAME=app
MIGRATE_SERVICE=migrate
IMAGE_NAME=subservice

run:
	go run ./cmd/main.go

build:
	go build -o $(IMAGE_NAME) ./cmd/main.go

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --env-file .env -p ${SERVER_PORT}:${SERVER_PORT} $(IMAGE_NAME)

compose-up:
	docker-compose up --build

compose-down:
	docker-compose down

compose-migrate:
	docker-compose run --rm $(MIGRATE_SERVICE)

docker-clean:
	docker-compose down -v

logs-app:
	docker-compose logs -f $(APP_NAME)
