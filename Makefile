.PHONY: build

build:
	docker compose build

run:
	docker compose up

clean:
	docker compose down -v

run_detach:
	docker compose up --detach

down:
	docker compose down