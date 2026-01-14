
build-frontend:
	echo "Building frontend"
	cd public/app/ && npm run build

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

	echo "2. Building backend"
	go build -o bin/tales cmd/tales/main.go