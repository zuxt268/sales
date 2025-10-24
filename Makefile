.PHONY: swag prod dev prod-build dev-build prod-down dev-down prod-logs dev-logs migrate-prod migrate-dev


swag:
	swag init -g cmd/sales/main.go

prod:
	docker compose -f docker-compose.prod.yml down
	docker image prune -f
	docker compose -f docker-compose.prod.yml build
	docker compose -f docker-compose.prod.yml up -d
	docker compose -f docker-compose.prod.yml ps

dev:
	docker compose -f docker-compose.dev.yml down
	docker image prune -f
	docker compose -f docker-compose.dev.yml build
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


prod-migrate-up:
	docker compose -f docker-compose.prod.yml exec app ./sql-migrate up -config=dbconfig.yml -env=production

dev-migrate-up:
	docker compose -f docker-compose.dev.yml exec app sql-migrate up -config=dbconfig.yml -env=development

prod-migrate-down:
	docker compose -f docker-compose.prod.yml exec app ./sql-migrate down -config=dbconfig.yml -env=production

dev-migrate-down:
	docker compose -f docker-compose.dev.yml exec app sql-migrate down -config=dbconfig.yml -env=development

token:
ifndef password
	@echo "Error: password is required. Usage: make token password=your_password"
	@exit 1
endif
	docker compose exec app ./token $(password)

