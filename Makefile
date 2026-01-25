.PHONY: swag prod dev prod-build dev-build prod-down dev-down prod-logs dev-logs migrate-prod migrate-dev wix wix-logs crawl crawl-logs prod-wix prod-wix-logs prod-crawl prod-crawl-logs prod-output prod-output-logs


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

wix:
	docker compose -f docker-compose.dev.yml up wix

wix-logs:
	docker compose -f docker-compose.dev.yml logs -f wix

crawl:
	docker compose -f docker-compose.dev.yml up crawl

crawl-logs:
	docker compose -f docker-compose.dev.yml logs -f crawl

prod-wix:
	docker compose -f docker-compose.prod.yml up wix

prod-wix-logs:
	docker compose -f docker-compose.prod.yml logs -f wix

prod-crawl:
	docker compose -f docker-compose.prod.yml up crawl

prod-crawl-logs:
	docker compose -f docker-compose.prod.yml logs -f crawl

prod-output:
	docker compose -f docker-compose.prod.yml up output

prod-output-logs:
	docker compose -f docker-compose.prod.yml logs -f output