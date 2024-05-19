# Use the official Golang image to create a build artifact
FROM golang:1.16 AS builder
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker's cache
COPY go.mod go.sum ./
RUN go mod download

# Run go mod tidy to ensure all dependencies are correctly included
RUN go mod tidy

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o main .

# Use a smaller base image for the runtime
FROM golang:1.16-alpine
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]
