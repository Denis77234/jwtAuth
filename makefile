.SILENT:

build:
	docker compose build

run: build
	docker compose up

buildLocal:
	go build ./cmd/main/main.go

runLocal: buildLocal
	./main

setDefault:
	export MONGO_URI=mongodb://localhost:27017
	export SERVER_PORT=:4000
	export ACCESS_SECRET=secret
	export REFRESH_SECRET=secret