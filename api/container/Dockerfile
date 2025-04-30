# Use the official Go image as the base image
FROM golang:1.24 AS builder
# Set the working directory inside the container
# Set the working directory inside the container
WORKDIR /app

# Install certificates
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Use the public Go proxy instead of direct connections
ENV GOPROXY=https://proxy.golang.org,direct

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Use a minimal image for the final container
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy any additional files (e.g., SQL migrations, static files)
COPY .env.s .env
COPY static ./static
COPY apc.db ./apc.db
# Expose the application port
EXPOSE 80

# Command to run the application
CMD ["./main"]