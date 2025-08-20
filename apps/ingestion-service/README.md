# Barclays-Grade Transaction Ingestion Service

A high-performance, enterprise-grade transaction ingestion service designed to handle 10,000+ transactions per second with bank-level security, reliability, and observability.

## 🚀 Key Features

### Security & Access Control
- **JWT Authentication**: OAuth2.0-style JWT tokens with role-based access control
- **RBAC**: Role-based permissions (teller, admin, auditor)
- **Account Isolation**: Users can only access their own accounts
- **mTLS Ready**: Prepared for mutual TLS implementation

### Idempotency & Reliability
- **Redis-backed Idempotency**: Prevents duplicate transaction processing
- **24-hour TTL**: Configurable idempotency window
- **Cached Responses**: Returns cached results for duplicate requests
- **Exactly-once Semantics**: End-to-end transaction deduplication

### Performance & Scalability
- **10,000+ TPS**: Designed for high-throughput banking operations
- **Kafka Partitioning**: Account-based partitioning for ordering guarantees
- **Async Publishing**: Non-blocking Kafka message publishing
- **Batch Processing**: Support for bulk transaction ingestion
- **Horizontal Scaling**: Stateless design for easy scaling

### Observability & Monitoring
- **Prometheus Metrics**: Comprehensive business and technical metrics
- **Real-time Monitoring**: Transaction rates, error rates, latency percentiles
- **Health Checks**: Built-in health monitoring endpoints
- **Structured Logging**: JSON-formatted logs for easy parsing

### Banking Domain Features
- **Account-based Partitioning**: Maintains transaction ordering per account
- **Multi-currency Support**: USD, EUR, GBP, INR, and more
- **Transaction Categories**: Groceries, utilities, entertainment, transport
- **Merchant Tracking**: Built-in merchant identification
- **Reference Numbers**: External reference tracking

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│  Ingestion API  │───▶│   Kafka Topic   │
│                 │    │                 │    │ transactions.raw│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Redis Cache   │
                       │  (Idempotency)  │
                       └─────────────────┘
```

## 📊 API Endpoints

### Authentication
- `POST /api/v1/auth/token` - Generate JWT token

### Transaction Ingestion
- `POST /api/v1/transactions` - Ingest single transaction
- `POST /api/v1/transactions/batch` - Ingest multiple transactions

### Monitoring
- `GET /health` - Service health check
- `GET /metrics` - Prometheus metrics

## 🔧 Configuration

Environment variables for configuration:

```bash
# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=transactions.raw

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24

# Security
RATE_LIMIT_PER_SECOND=10000
MAX_REQUEST_SIZE=1048576

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
```

## 🚀 Quick Start

### Prerequisites
- Go 1.25+
- Redis 7+
- Kafka 2.8+
- Docker & Docker Compose

### Local Development

1. **Start dependencies:**
```bash
docker-compose up -d redis
```

2. **Run the service:**
```bash
go run main.go
```

3. **Test with curl:**
```bash
# Generate a token
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test_user","account_id":"acc_123","roles":["teller"]}'

# Ingest a transaction
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Idempotency-Key: unique_key_123" \
  -d '{
    "idempotency_key": "unique_key_123",
    "account_id": "acc_123",
    "user_id": "user_456",
    "amount": 99.99,
    "currency": "USD",
    "type": "debit",
    "category": "groceries",
    "merchant": "Walmart"
  }'
```

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build manually
docker build -t ingestion-service .
docker run -p 8080:8080 -p 9090:9090 ingestion-service
```

## 📈 Load Testing

### Using k6

1. **Install k6:**
```bash
# macOS
brew install k6

# Windows
choco install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C22D63094710
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

2. **Run load test:**
```bash
k6 run load-test.js
```

3. **Customize test parameters:**
```bash
k6 run -e BASE_URL=http://your-service:8080 load-test.js
```

### Load Test Scenarios

The included load test script:
- Ramps up from 0 to 10,000 concurrent users
- Tests idempotency with duplicate requests
- Validates JWT authentication
- Measures response times and throughput
- Generates realistic transaction data

## 📊 Metrics & Monitoring

### Prometheus Metrics

Key metrics available at `/metrics`:

- **HTTP Metrics**: Request counts, durations, error rates
- **Business Metrics**: Transactions ingested, failed, by currency/type
- **Kafka Metrics**: Message publish rates, durations, failures
- **Redis Metrics**: Operation counts, durations, success rates

### Grafana Dashboard

Create a Grafana dashboard with these key panels:

1. **Throughput**: Transactions per second
2. **Latency**: P95, P99 response times
3. **Error Rates**: Failed transactions by reason
4. **Kafka Lag**: Consumer group lag
5. **Redis Performance**: Operation latency and throughput

## 🔒 Security Considerations

### Production Deployment

1. **JWT Secret**: Use strong, randomly generated secrets
2. **Redis Security**: Enable authentication and TLS
3. **Network Security**: Use mTLS, VPN, or private networks
4. **Rate Limiting**: Implement per-client rate limits
5. **Audit Logging**: Log all authentication and authorization events

### Compliance

- **PCI DSS**: Transaction data handling
- **GDPR**: Personal data protection
- **SOX**: Financial transaction audit trails
- **Basel III**: Risk management and capital adequacy

## 🚀 Performance Tuning

### Kafka Optimization

```yaml
# kafka-config.yaml
num.partitions: 50
default.replication.factor: 3
min.insync.replicas: 2
log.flush.interval.messages: 10000
log.flush.interval.ms: 1000
```

### Redis Optimization

```bash
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### Go Runtime

```bash
export GOMAXPROCS=8
export GOGC=100
export GOMEMLIMIT=2GiB
```

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### Load Tests
```bash
k6 run load-test.js
```

## 📚 API Documentation

### Transaction Request Schema

```json
{
  "idempotency_key": "string (required)",
  "account_id": "string (required)",
  "user_id": "string (required)",
  "amount": "number (required, > 0)",
  "currency": "string (required)",
  "type": "string (required: debit|credit)",
  "category": "string (required)",
  "merchant": "string (optional)",
  "reference": "string (optional)",
  "metadata": "object (optional)"
}
```

### Response Schema

```json
{
  "id": "string",
  "status": "string",
  "message": "string",
  "timestamp": "string (ISO 8601)"
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the documentation

---

**Note**: This service is designed for production banking environments. Ensure proper security review and testing before deployment in production.
