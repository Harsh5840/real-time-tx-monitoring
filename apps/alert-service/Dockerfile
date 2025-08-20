# apps/alert-service/Dockerfile

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
RUN go build -trimpath -ldflags="-s -w" -o /app/alert-service .

# ---- Run Stage ----
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/alert-service /app/alert-service

USER nonroot:nonroot
ENTRYPOINT ["/app/alert-service"]

