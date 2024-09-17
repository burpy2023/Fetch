# Build stage
FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .

# Download dependencies
RUN go mod download

# Run the tests
RUN go test ./... -v

# Build the application
RUN go build -o /go/bin/app

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
RUN chmod +x /app

# Expose port 8080 (since your server is listening on this port)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]

LABEL Name="receiptprocessor" Version="0.0.1"
