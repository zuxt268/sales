.PHONY: swag run


swag:
	swag init -g cmd/sales/main.go


rebuild:
	docker compose down
	docker compose build --no-cache
	docker compose up -d
	docker compose ps


token:
ifndef password
	@echo "Error: password is required. Usage: make token password=your_password"
	@exit 1
endif
	docker compose exec app ./token $(password)