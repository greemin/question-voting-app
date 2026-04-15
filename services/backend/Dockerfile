# Stage 1: Build the Go binary
# $BUILDPLATFORM runs this stage natively on the CI host; TARGETARCH is passed
# in so Go cross-compiles the binary for the actual target platform.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETARCH

WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o main ./cmd/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8081
EXPOSE 8081

# Command to run the application
CMD ["./main"]
