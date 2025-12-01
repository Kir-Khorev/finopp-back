# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
RUN go mod download && go mod tidy

# Copy source code
COPY . .
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/api /app/api

EXPOSE 8080

CMD ["/app/api"]

