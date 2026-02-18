# ============================================================
# Stage 1: Build the Creator UI (Vite + Svelte)
# ============================================================
FROM node:20-alpine AS build-frontend

WORKDIR /src/public/app
COPY public/app/package.json public/app/package-lock.json* ./
RUN npm ci --legacy-peer-deps
COPY public/app/ ./
RUN npm run build

# ============================================================
# Stage 2: Build the Game Client (Rollup + Svelte)
# ============================================================
FROM node:20-alpine AS build-mud-client

WORKDIR /src/public/mud-client
COPY public/mud-client/package.json public/mud-client/package-lock.json* ./
RUN npm ci --legacy-peer-deps
COPY public/mud-client/ ./
RUN npm run build

# ============================================================
# Stage 3: Build the Go binary
# ============================================================
FROM golang:1.24-alpine AS build-backend

WORKDIR /src

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Copy frontend builds into the Go embed directories
COPY --from=build-frontend /src/public/app/dist/ pkg/webui/dist/
COPY --from=build-mud-client /src/public/mud-client/public/ pkg/webuiplay/dist/

# Build statically-linked binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tales cmd/tales/main.go

# ============================================================
# Stage 4: Minimal runtime image
# ============================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# Non-root user for security
RUN adduser -D -h /app tales
WORKDIR /app

COPY --from=build-backend /tales /app/tales

# Ensure writable dirs for SQLite and uploads
RUN mkdir -p /app/data /app/uploads /app/backups && \
    chown -R tales:tales /app

USER tales

EXPOSE 8010

ENTRYPOINT ["/app/tales"]
