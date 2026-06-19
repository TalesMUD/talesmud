
build-frontend:
	echo "Building main frontend"
	cd public/app/ && npm run build
	echo "Copying frontend build into Go-embeddable dist folder"
	rm -rf pkg/webui/dist
	mkdir -p pkg/webui/dist
	cp -r public/app/dist/* pkg/webui/dist/

build-mud-client:
	echo "Building mud-client (game client)"
	cd public/mud-client/ && npm install && npm run build
	echo "Copying mud-client build into Go-embeddable dist folder"
	rm -rf pkg/webuiplay/dist
	mkdir -p pkg/webuiplay/dist
	cp -r public/mud-client/public/* pkg/webuiplay/dist/

prepare-embedded-assets:
	@if [ ! -e pkg/webui/dist/index.html ]; then \
		echo "Preparing fallback Creator UI embedded assets"; \
		mkdir -p pkg/webui/dist; \
		if [ -e public/app/dist/index.html ]; then \
			cp -r public/app/dist/* pkg/webui/dist/; \
		else \
			printf '%s\n' '<!doctype html><title>TalesMUD Creator</title><main>Run make build-frontend to embed the Creator UI.</main>' > pkg/webui/dist/index.html; \
		fi; \
	fi
	@if [ ! -e pkg/webuiplay/dist/index.html ]; then \
		echo "Preparing MUD client embedded assets"; \
		mkdir -p pkg/webuiplay/dist; \
		cp -r public/mud-client/public/* pkg/webuiplay/dist/; \
	fi

build-backend: prepare-embedded-assets
	echo "Building backend"
	go build -o bin/tales cmd/tales/main.go

build-dialogs-sandbox:
	echo "Building Dialogs sandbox"
	go build -o bin/dialog_sandbox cmd/dialog_sandbox/main.go

run-dialogs-sandbox:
	echo "Starting dialogs sandbox..."
	go run cmd/dialog_sandbox/main.go

run-server: prepare-embedded-assets
	echo "Starting tales server ..."
	go run cmd/tales/main.go

run-frontend:
	echo "Starting main frontend ..."
	cd public/app/ && npm run dev

run-mud-client:
	echo "Starting mud-client (game client) ..."
	cd public/mud-client/ && npm install && npm run dev

run: ; ${MAKE} -j4 run-server run-frontend run-mud-client

# Dev mode: Go backend (go run) + Vite/Rollup dev servers with hot reload.
# Main app: http://localhost:5173 (Vite proxies /api and /ws to Go backend)
# Game client: http://localhost:8080 (talks to Go backend via CORS)
# Go backend: http://localhost:8010
dev: ; ${MAKE} -j4 run-server run-frontend run-mud-client

build:
	echo "1. Building main frontend"
	cd public/app/ && npm run build
	echo "1a. Copying frontend build into Go-embeddable dist folder"
	rm -rf pkg/webui/dist
	mkdir -p pkg/webui/dist
	cp -r public/app/dist/* pkg/webui/dist/

	echo "2. Building mud-client (game client)"
	cd public/mud-client/ && npm install && npm run build
	echo "2a. Copying mud-client build into Go-embeddable dist folder"
	rm -rf pkg/webuiplay/dist
	mkdir -p pkg/webuiplay/dist
	cp -r public/mud-client/public/* pkg/webuiplay/dist/

	echo "3. Building backend"
	go build -o bin/tales cmd/tales/main.go
