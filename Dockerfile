# Stage 1: Build the React frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend source
COPY frontend/ ./

# Install dependencies and build the Vite React application
# Note: This assumes package.json exists.
RUN npm install && npm run build

# Stage 2: Build the Go backend
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app

# Copy Go module files and download dependencies
# This is done before copying the rest of the code to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the built frontend assets from the previous stage
# This ensures the Go embed directive picks up the actual production build
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Build the statically linked Go binary
# -w -s removes debugging information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /vaultify-server ./cmd/server

# Stage 3: Minimal Runtime
FROM gcr.io/distroless/static-debian12 AS runtime

WORKDIR /

# Copy the compiled binary from the backend-builder stage
COPY --from=backend-builder /vaultify-server /vaultify-server

# Copy the database migrations
COPY --from=backend-builder /app/db /db

# Ensure we run as a non-root user (standard in distroless images)
USER nonroot:nonroot

# Expose the API port
EXPOSE 8080

# Start the server
ENTRYPOINT ["/vaultify-server"]
