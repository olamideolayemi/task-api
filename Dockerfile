# Use an official Go image
FROM golang:1.22-alpine

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
