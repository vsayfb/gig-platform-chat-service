.PHONY: run down clean logs

run:
	@echo "Starting chat service environment..."
	docker compose up --build

down:
	@echo "Stopping chat service environment..."
	docker compose down

clean:
	@echo "Removing containers and volumes..."
	docker compose down -v

logs:
	docker compose logs -f