.PHONY: swag run


swag:
	swag init -g cmd/sales/main.go


run:
	docker compose down
	docker compose build app --no-cache
	docker compose up --build -d
	docker compose ps


token:
ifndef password
	@echo "Error: password is required. Usage: make token password=your_password"
	@exit 1
endif
	docker compose exec app ./token $(password)