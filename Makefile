run: ## Run locally
	go run ./cmd/api
test:
	go test ./...
docker:
	docker build -t admira-api .
up:
	docker compose up --build
