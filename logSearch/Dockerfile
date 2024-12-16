# Use the official Golang image as the base image
# Stage 1: Build the Go app for Linux x86 architecture
FROM golang:alpine AS builder

# Install dependencies
RUN apk update && apk add --no-cache git gcc libc-dev musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first, and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code to the working directory
COPY . .

# Build the Go application as a statically linked binary for Linux x86_64 architecture
ENV CGO_ENABLED 1
ENV GOOS linux
ENV GOARCH amd64
# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -tags musl -o main .
FROM alpine:latest

# Set up necessary certificates (for HTTPS communication, if required)
RUN apk add --no-cache ca-certificates

# Set the working directory in the second stage
WORKDIR /root/

# Copy the compiled Go binary from the first stage
COPY --from=builder /app/main .

# Expose the port that the app will run on (e.g., 9000)
EXPOSE 3000

# Run the Go application
# CMD ["./mail"]

# # Expose the port the application will run on
# EXPOSE 3000

# Command to run the application
CMD ["./main"]