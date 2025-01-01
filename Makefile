run:
	go run -exec "env $$(cat .env | xargs)" cmd/main.go

bufgen:
	cd proto && rm -rf gen && buf generate

mongo-up:
	docker compose up -d mongo
mongo-down:
	docker compose down mongo