# Stage 1: Build
FROM golang:1.22.2-alpine As builder

# Labels for the build stage
LABEL maintainer="obimbira60@gmail.com" \
    version="1.0" \
    description="Build stage for Go application"

# Install build dependencies for CGO and sqlite3
RUN apk add --no-cache gcc musl-dev sqlite sqlite-dev

# Set the working directory in the build stage
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the project code to the /app directory in the container
COPY . .

# Build the Go binary with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o mybinary

# Stage 2: Final Image
FROM alpine:latest

# Labels for the final image
LABEL maintainer="obimbira60@gmail.com" \
    version="1.0" \
    description="Final image with Go application"

# Install runtime dependencies for sqlite3
RUN apk add --no-cache sqlite-libs sqlite

# Create app directory
WORKDIR /app
RUN mkdir -p /app/data

# Copy the files from the builder stage
COPY --from=builder /app/mybinary /app/
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/uploads /app/uploads
COPY --from=builder /app/forum.db /app/forum.db

# Set permissions for the app directory
RUN chmod 755 /app && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

# Switch to non-root user for security
USER appuser

# Expose the port the application will run on
EXPOSE 8080

# Command to run the binary
CMD ["./mybinary"]