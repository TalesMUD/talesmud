
build-frontend:
	echo "Building frontend"
	cd public/app/ && npm run build
	echo "Copying Svelte Flow CSS"
	cp public/app/node_modules/@xyflow/svelte/dist/style.css public/app/public/svelte-flow.css
	echo "Copying frontend build into Go-embeddable dist folder"
	rm -rf pkg/webui/dist
	mkdir -p pkg/webui/dist
	cp -r public/app/public/* pkg/webui/dist/

build-backend:
	echo "Building backend"
	go build -o bin/tales cmd/tales/main.go

build-dialogs-sandbox:
	echo "Building Dialogs sandbox"
	go build -o bin/dialog_sandbox cmd/dialog_sandbox/main.go

run-dialogs-sandbox:
	echo "Starting dialogs sandbox..."
	go run cmd/dialog_sandbox/main.go

run-server:
	echo "Starting tales server ..."
	go run cmd/tales/main.go

run-frontend:
	echo "Starting tales frtontend ..."
	cd public/app/ && npm run dev

run: ; ${MAKE} -j4 run-server run-frontend

build:
	echo "1. Building frontend"
	cd public/app/ && npm run build
	echo "1a. Copying Svelte Flow CSS"
	cp public/app/node_modules/@xyflow/svelte/dist/style.css public/app/public/svelte-flow.css
	echo "1b. Copying frontend build into Go-embeddable dist folder"
	rm -rf pkg/webui/dist
	mkdir -p pkg/webui/dist
	cp -r public/app/public/* pkg/webui/dist/

	echo "2. Building backend"
	go build -o bin/tales cmd/tales/main.go