BASE_STACK = docker compose -f docker-compose.yml
GARAGE_SETUP_SCRIPT_WIN = ./scripts/setup-garage-win.sh
GARAGE_SETUP_SCRIPT_LIN = ./scripts/setup-garage-lin.sh

compose-up-garage: ### Run docker compose(garage)
	mkdir -p garage/data garage/meta ; \
	$(BASE_STACK) up --build -d garage
.PHONY: compose-up-garage

compose-up-all: ### Run docker compose
	$(BASE_STACK) up --build -d
.PHONY: compose-up

compose-down: ### Down docker compose
	$(BASE_STACK) down
.PHONY: compose-down

down-n-clean: ### Down docker compose, delete volumes, delete garage data, metadata
	$(BASE_STACK) down -v ; \
	rm -rf ./garage/data/* ./garage/meta/*
.PHONY: down-n-clean

setup-garage-win: ### Run garage setup script for Windows
	$(GARAGE_SETUP_SCRIPT_WIN)
.PHONY: setup-garage

setup-garage-lin: ### Run garage setup script for Linux
	$(GARAGE_SETUP_SCRIPT_LIN)
.PHONY: setup-garage

deps: ### Deps tidy + verify
	go mod tidy && go mod verify
.PHONY: deps