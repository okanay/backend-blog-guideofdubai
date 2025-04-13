.PHONY: build run dev clean kill pull help migrate-db migrate-build migrate-up migrate-down migrate-up-all migrate-down-all migrate-goto migrate-force migrate-version migrate-clean

# Değişkenler
PORT=8080
BINARY_NAME=guideofdubai-blog
BUILD_DIR=./build
AIR_PATH=$(HOME)/go/bin/air
MIGRATE_TOOL_SRC := ./cmd/migrate/main.go
MIGRATE_TOOL_BIN := ./bin/migrate_tool
MIGRATIONS_PATH := ./database/migrations
BACKUP_PATH := ./database/backups

# Ana komutlar
build:
	@echo "Building application..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

run: build
	@echo "Running application..."
	$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@echo "Starting development server with Air..."
	$(AIR_PATH) || (echo "Air not found. Installing..." && go install github.com/cosmtrek/air@latest && $(HOME)/go/bin/air)

# Temizleme
clean:
	@echo "Cleaning build files..."
	rm -rf $(BUILD_DIR) tmp
	go clean

# Port işlemleri
kill:
	@echo "Killing process on port $(PORT)..."
	-lsof -ti:$(PORT) | xargs kill -9 || npx kill-port $(PORT)

# Güncelleme
pull:
	@echo "Pulling latest changes..."
	git pull && go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go && sudo systemctl restart api-menuarts.service

# Migration file oluştur
migrate-db:
	@test -n "${n}" || (echo "Error: 'n' (name) is not set. Use 'make db n=yourfilename'"; exit 1)
	migrate create -ext sql -dir database/migrations -seq ${n}

migrate-build:
	@echo ">> Migration tool is being built..."
	@mkdir -p $(shell dirname ${MIGRATE_TOOL_BIN}) # create bin directory
	@go build -o ${MIGRATE_TOOL_BIN} ${MIGRATE_TOOL_SRC}
	@echo ">> Build completed: ${MIGRATE_TOOL_BIN}"

# One step forward
migrate-up: migrate-build
	@echo ">> Migration: Up (1 step)"
	@${MIGRATE_TOOL_BIN} up

# One step backward
migrate-down: migrate-build
	@echo ">> Migration: Down (1 step)"
	@${MIGRATE_TOOL_BIN} down

# All steps forward
migrate-up-all: migrate-build
	@echo ">> Migration: Up All"
	@${MIGRATE_TOOL_BIN} up_all

# All steps backward
migrate-down-all: migrate-build
	@echo ">> Migration: Down All"
	@${MIGRATE_TOOL_BIN} down_all

# Migrate to a specific version
migrate-goto: migrate-build
	@if [ -z "$(V)" ]; then \
		echo "ERROR: You must specify a version number with the 'V' variable."; \
		echo "Example: make migrate-goto V=3"; \
		exit 1; \
	fi
	@echo ">> Migration: Goto version $(V)"
	@${MIGRATE_TOOL_BIN} goto $(V)

# Force migration to a specific version
migrate-force: migrate-build
	@if [ -z "$(V)" ]; then \
		echo "ERROR: You must specify a version number with the 'V' variable."; \
		echo "Example: make migrate-force V=2"; \
		exit 1; \
	fi
	@echo ">> Migration: Force version $(V)"
	@read -p "WARNING: This will change the database record but not run the files. Are you sure? (y/N) " choice; \
	if [ "$${choice}" = "y" ] || [ "$${choice}" = "Y" ]; then \
		${MIGRATE_TOOL_BIN} force $(V); \
	else \
		echo "Operation cancelled."; \
		exit 1; \
	fi

# Show current version
migrate-version: migrate-build
	@echo ">> Migration: Version"
	@${MIGRATE_TOOL_BIN} version

# Clean built files
migrate-clean:
	@echo ">> Cleaning..."
	@rm -f ${MIGRATE_TOOL_BIN}
	@echo ">> Clean completed."

# Backup database
migrate-backup-pull: migrate-build
	@echo ">> Migration: Backup Pull (Creating database backup)"
	@# BACKUP_PATH environment değişkeninin migrate_tool tarafından okunabilmesi için export ediyoruz
	@export BACKUP_PATH=$(BACKUP_PATH); \
	 ${MIGRATE_TOOL_BIN} backup-pull

# Restore database from latest backup
migrate-backup-push: migrate-build
	@echo ">> Migration: Backup Push (Restoring from latest backup)"
	@# BACKUP_PATH environment değişkeninin migrate_tool tarafından okunabilmesi için export ediyoruz
	@export BACKUP_PATH=$(BACKUP_PATH); \
	 ${MIGRATE_TOOL_BIN} backup-push

# Yardım
help:
	@echo "Available commands:"
	@echo "  make build               - Build the application"
	@echo "  make run                 - Build and run the application"
	@echo "  make dev                 - Run with hot-reload using Air"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make kill                - Kill process running on port $(PORT)"
	@echo "  make pull                - Pull latest changes and restart service"
	@echo "  make help                - Display this help message"
	@echo ""
	@echo "Migration Commands:"
	@echo "  make migrate-build       - Build the migration tool (${MIGRATE_TOOL_BIN})"
	@echo "  make migrate-db n=name   - Create a new migration file with the given name"
	@echo "  make migrate-up          - Apply one step forward migration"
	@echo "  make migrate-down        - Revert one step backward migration"
	@echo "  make migrate-up-all      - Apply all pending migrations"
	@echo "  make migrate-down-all    - Revert all migrations"
	@echo "  make migrate-goto V=<version> - Migrate to the specified <version>"
	@echo "  make migrate-force V=<version> - Force migration to <version> (use with caution!)"
	@echo "  make migrate-version     - Show the current migration version"
	@echo "  make migrate-clean       - Clean built files"
	@echo "  make migrate-backup-pull  - Create a backup of the current database"
	@echo "  make migrate-backup-push  - Restore the database from the latest backup (use with caution!)"
	@echo "  make migrate-clean        - Clean built files"
