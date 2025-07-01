.PHONY: run build clean inspector build-inspector package

# Build the RPG demo
build:
	go build -o rpg_demo *.go

# Build the inspector tool
build-inspector:
	go build -o aseprite-inspector cmd/inspector/main.go

# Build both tools
build-all: build build-inspector

# Run the demo
run:
	go run *.go

# Run the inspector tool
inspector:
	@echo "Usage: make inspect FILE=<aseprite-file>"
	@echo "Example: make inspect FILE=assets/Soldier.aseprite"

# Inspect a specific file
inspect:
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE parameter is required"; \
		echo "Usage: make inspect FILE=<aseprite-file>"; \
		echo "Example: make inspect FILE=assets/Soldier.aseprite"; \
		exit 1; \
	fi
	go run cmd/inspector/main.go $(FILE)

# Clean build artifacts
clean:
	rm -f rpg_demo aseprite-inspector
	rm -rf rpg_demo.app

# Build for different platforms
build-windows:
	GOOS=windows GOARCH=amd64 go build -o rpg_demo.exe *.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o rpg_demo-linux *.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o rpg_demo-mac *.go
