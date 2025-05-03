# Variables
APP_NAME = converterd
DOCKER_IMAGE = ghcr.io/alex4108/$(APP_NAME):latest

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building the application..."
	go build -o $(APP_NAME) .

# Build the Docker container
.PHONY: build-container
build-container:
	@echo "Building the Docker container..."
	docker build -t $(DOCKER_IMAGE) .

# Run the application locally
.PHONY: run
run: build
	@echo "Running the application locally..."
	mkdir -p /tmp/converterd
	CHECK_SECONDS=1 WATCH_FOLDERS=/tmp/converterd ./$(APP_NAME)

# Run the application in a Docker container
.PHONY: run-container
run-container: build-container
	@echo "Running the application in a Docker container..."
	docker run --rm -it $(DOCKER_IMAGE)