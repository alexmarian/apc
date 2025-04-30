# Use the official Go image as the base image
FROM golang:1.23 as builder
# Set the working directory inside the container
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

ENV GOPROXY=direct

WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY .. .

# Build the Go application
RUN go build -o main .

# Use a minimal image for the final container
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy any additional files (e.g., SQL migrations, static files)
COPY sql/ ./sql/
COPY .env.s .env

# Expose the application port
EXPOSE 80

# Command to run the application
CMD ["./main"]