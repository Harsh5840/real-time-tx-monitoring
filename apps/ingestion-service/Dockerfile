# apps/ingestion-service/Dockerfile

# ---- Build Stage ----
    FROM golang:1.22 AS builder

    WORKDIR /app
    ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
    
    # Cache deps
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy source
    COPY . .
    
    # Build the binary from the module root (main.go at .)
    RUN go build -trimpath -ldflags="-s -w" -o /app/ingestion-service .
    
    # ---- Run Stage ----
    FROM gcr.io/distroless/static-debian12
    
    WORKDIR /app
    COPY --from=builder /app/ingestion-service /app/ingestion-service
    
    USER nonroot:nonroot
    EXPOSE 8080
    ENTRYPOINT ["/app/ingestion-service"]