# =====================================================
# Database Migration Makefile
# make migrate-up / make migrate-down / make migrate-create name=add_user_table
# =====================================================

include .env
export

# --------------------------
# config
# --------------------------
DB_URL=postgres://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_NAME}?sslmode=${PG_SSLMODE}
MIGRATIONS_DIR=infra/pg/migrations

MIGRATE_CMD=docker run --rm \
	-v $(PWD)/$(MIGRATIONS_DIR):/migrations \
	--network host migrate/migrate \
	-path=/migrations -database "$(DB_URL)"

# --------------------------
# commands
# --------------------------

migrate-up:
	@echo "===> Running migrations UP..."
	$(MIGRATE_CMD) up

migrate-down:
	@echo "===> Rolling back one migration..."
	$(MIGRATE_CMD) down 1

migrate-reset:
	@echo "===> Rolling back all migrations..."
	$(MIGRATE_CMD) down

migrate-version:
	@echo "===> Current migration version:"
	$(MIGRATE_CMD) version

migrate-force:
ifndef v
	$(error Please provide version, example: make migrate-force v=1)
endif
	@echo "===> Forcing migration version to $(v)..."
	$(MIGRATE_CMD) force $(v)

## create new migration file
## make migrate-create name=add_user_table
migrate-create:
ifndef name
	$(error please provide migration name, example: make migrate-create name=init_schema)
endif
	@echo "===> Creating new migration: $(name)"
	docker run --rm -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate create -seq -ext sql -dir /migrations $(name)

migrate-clean:
	@echo "===> Cleaning Docker migrate container..."
	docker rmi migrate/migrate || true

