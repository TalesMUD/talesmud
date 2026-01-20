
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
	echo "Starting main frontend ..."
	cd public/app/ && npm run dev

run-mud-client:
	echo "Starting mud-client (game client) ..."
	cd public/mud-client/ && npm install && npm run dev

run: ; ${MAKE} -j4 run-server run-frontend run-mud-client

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