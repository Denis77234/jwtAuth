.SILENT:

build:
	docker compose build

run: build
	docker compose up

buildLocal:
	go build ./cmd/main/main.go

runLocal: buildLocal
	./main
