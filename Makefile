
build-frontend:
	echo "Building frontend"
	cd public/app/ && npm run build

build-backend:
	echo "Building backend"
	go build -o bin/tales cmd/tales/main.go


run-server:
	echo "Starting tales server ..."
	go run cmd/tales/main.go

run-frontend:
	echo "Starting tales frtontend ..."
	cd public/app/ && npm run dev

run: ; ${MAKE} -j4 run-server run-frontend

build:
	build-frontend
	build-backend
