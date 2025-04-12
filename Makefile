.PHONY: build run dev clean kill pull help db up down force

# Değişkenler
PORT=8080
BINARY_NAME=guideofdubai-blog
BUILD_DIR=./build
AIR_PATH=$(HOME)/go/bin/air

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
db:
	@test -n "${n}" || (echo "Error: 'n' (name) is not set. Use 'make db n=yourfilename'"; exit 1)
	migrate create -ext sql -dir database/migrations -seq ${n}

# Migration up
up:
	@echo "Running migration up..."
	go run cmd/migrate/up/main.go

# Migration down
down:
	@echo "Running migration down..."
	go run cmd/migrate/down/main.go

# Migration force (belirli bir versiyona zorla)
force:
	@test -n "${v}" || (echo "Error: 'v' (version) is not set. Use 'make force v=desired_version'"; exit 1)
	@echo "Forcing migration to version ${v}..."
	go run cmd/migrate/force/main.go ${v}

# Yardım
help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Build and run the application"
	@echo "  make dev         - Run with hot-reload using Air"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make kill        - Kill process running on port $(PORT)"
	@echo "  make pull        - Pull latest changes and restart service"
	@echo "  make db n=name   - Create a new migration file with the given name"
	@echo "  make up          - Run migration up"
	@echo "  make down        - Run migration down"
	@echo "  make force v=ver - Force migration to a specific version"
	@echo "  make help        - Display this help message"
