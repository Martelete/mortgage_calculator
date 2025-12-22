APP_NAME=mortgage_calculator

.PHONY: run build docker-build docker-run docker-stop clean

run:
	go run browser.go

build:
	go build -o $(APP_NAME) browser.go

docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker run --rm -p 8080:8080 $(APP_NAME)

docker-stop:
	docker compose down

clean:
	rm -f $(APP_NAME)
