.PHONY: swag prod dev prod-build dev-build prod-down dev-down prod-logs dev-logs


swag:
	swag init -g cmd/sales/main.go

prod:
	docker compose -f docker-compose.prod.yml down
	docker compose -f docker-compose.prod.yml build --no-cache
	docker compose -f docker-compose.prod.yml up -d
	docker compose -f docker-compose.prod.yml ps

dev:
	docker compose -f docker-compose.dev.yml down
	docker compose -f docker-compose.dev.yml build --no-cache
	docker compose -f docker-compose.dev.yml up -d
	docker compose -f docker-compose.dev.yml ps

prod-build:
	docker compose -f docker-compose.prod.yml build --no-cache

dev-build:
	docker compose -f docker-compose.dev.yml build --no-cache

prod-down:
	docker compose -f docker-compose.prod.yml down

dev-down:
	docker compose -f docker-compose.dev.yml down

prod-logs:
	docker compose -f docker-compose.prod.yml logs -f

dev-logs:
	docker compose -f docker-compose.dev.yml logs -f


token:
ifndef password
	@echo "Error: password is required. Usage: make token password=your_password"
	@exit 1
endif
	docker compose exec app ./token $(password)