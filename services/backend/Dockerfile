# Stage 1: Build the Go binary
FROM golang:1.18-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy the data directory
# This is needed for the file-based storage
# In a later step, this will be replaced with a database
COPY data ./data

# Expose port 8081
EXPOSE 8081

# Command to run the application
CMD ["./main"]
