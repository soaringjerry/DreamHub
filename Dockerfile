# Stage 1: Build Go Backend (Server & Worker)
# Use a Go version that matches go.mod requirement (>= 1.23)
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server cmd/server/main.go
# Build worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /worker cmd/worker/main.go

# Stage 2: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
# Use ci for potentially faster and more reliable installs in CI
RUN npm ci
COPY frontend/ ./
# Add build-time args if needed, e.g., ARG VITE_API_URL
# RUN npm run build -- --base=./ --mode production -- VITE_API_URL=$VITE_API_URL
RUN npm run build

# Stage 3: Final Image
FROM alpine:latest
# Install supervisor and any other runtime dependencies
RUN apk update && apk add --no-cache supervisor ca-certificates tzdata && rm -rf /var/cache/apk/*
WORKDIR /app
# Copy Go binaries from builder stage
COPY --from=builder /server /app/server
COPY --from=builder /worker /app/worker
# Copy frontend build from frontend-builder stage
# Assuming the server serves frontend files from a 'public' or 'static' directory
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
# Copy supervisor config
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
# Copy .env.example for reference, but actual .env should be mounted or managed externally on the server
# COPY .env.example .env.example

# Expose the port the server listens on (adjust if needed)
EXPOSE 8080

# Set the entrypoint to supervisor
ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
# CMD is defined in supervisord.conf