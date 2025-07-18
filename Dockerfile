# Build stage
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fm-actions .

# Runtime stage
FROM alpine:3.20

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S cloudbees && adduser -S cloudbees -G cloudbees

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/fm-actions .

# Change ownership to cloudbees user
RUN chown cloudbees:cloudbees /app/fm-actions

# Switch to non-root user
USER cloudbees

# Set the binary as executable
RUN chmod +x /app/fm-actions

# Entry point
ENTRYPOINT ["/app/fm-actions"] 