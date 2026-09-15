.PHONY: backend-build backend-test backend-run frontend-install frontend-dev frontend-build docker-up docker-down

backend-build:
	cd backend && go build -o ../bin/void ./cmd/void

backend-test:
	cd backend && go test ./...

backend-run: backend-build
	./bin/void serve

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
