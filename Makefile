GO ?= go
TEST_PKGS := ./services/api-gateway/internal/app ./services/auth-service/internal/app ./services/user-service/internal/app ./services/wallet-service/internal/app ./services/notification-service/internal/app ./shared/config ./shared/security ./shared/middleware ./shared/grpcx

.PHONY: deps fmt test test-all coverage build run-auth run-user run-wallet run-notification run-gateway docker-up docker-down docker-logs

deps:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

test:
	$(GO) test $(TEST_PKGS) -coverprofile coverage.out -covermode atomic

test-all:
	$(GO) test ./...

coverage:
	$(GO) tool cover -func coverage.out

build:
	$(GO) build ./...

run-auth:
	$(GO) run ./services/auth-service/cmd

run-user:
	$(GO) run ./services/user-service/cmd

run-wallet:
	$(GO) run ./services/wallet-service/cmd

run-notification:
	$(GO) run ./services/notification-service/cmd

run-gateway:
	$(GO) run ./services/api-gateway/cmd

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f --tail=200
