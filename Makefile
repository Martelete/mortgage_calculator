APP_NAME=mortgage_calculator

.PHONY: run build docker-build docker-run docker-stop clean

run:
	go run main.go

build:
	go build -o $(APP_NAME) main.go

docker-build-and-run:
	docker build -t $(APP_NAME) . && docker run --rm -p 8080:8080 $(APP_NAME)

docker-stop:
	docker compose down

clean:
	rm -f $(APP_NAME)
