# Stage 1: Build Go Backend (Server & Worker)
# Use a Go version that matches go.mod requirement (>= 1.23)
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Download migrate CLI tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# Move migrate to a temporary location to avoid COPY issues later
RUN mkdir -p /tmp/bin && mv /go/bin/migrate /tmp/bin/migrate
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
COPY --from=builder /tmp/bin/migrate /usr/local/bin/migrate # Copy migrate CLI, specifying destination filename
# Copy frontend build from frontend-builder stage
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
# Copy migrations
COPY migrations /app/migrations
# Copy supervisor config
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
# Copy .env.example for reference, but actual .env should be mounted or managed externally on the server
# COPY .env.example .env.example
# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Expose the port the server listens on (adjust if needed)
EXPOSE 8080

# Use an entrypoint script to run migrations first
ENTRYPOINT ["/docker-entrypoint.sh"]
# CMD will be executed by the entrypoint script (supervisord)
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]