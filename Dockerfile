# Step 1: Build the Go binary
FROM golang:1.19-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o nexus .

# Step 2: Create a minimal image with only the binary
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the pre-built binary from the build stage
COPY --from=build /app/nexus .

RUN apk add jq bash

# Expose the necessary port
EXPOSE 8545

# Command to run the binary
ENTRYPOINT ["./nexus"]
