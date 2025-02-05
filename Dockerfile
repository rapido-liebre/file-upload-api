# Use the official Go image as the base
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the application source code
COPY . .

# Install `swag` for OpenAPI documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
RUN swag init --dir ./pkg --output ./docs

# Build the application
RUN go build -o main ./pkg/main.go

CMD ["./main"]

