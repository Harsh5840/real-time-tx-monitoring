import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    // Ramp up to 1000 users over 30 seconds
    { duration: '30s', target: 1000 },
    // Stay at 1000 users for 2 minutes
    { duration: '2m', target: 1000 },
    // Ramp up to 5000 users over 1 minute
    { duration: '1m', target: 5000 },
    // Stay at 5000 users for 3 minutes
    { duration: '3m', target: 5000 },
    // Ramp up to 10000 users over 1 minute
    { duration: '1m', target: 10000 },
    // Stay at 10000 users for 5 minutes
    { duration: '5m', target: 10000 },
    // Ramp down to 0 users over 30 seconds
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate should be below 1%
    http_req_rate: ['rate>8000'],     // Should maintain at least 8000 req/s
  },
};

// Test data generation
function generateTransaction() {
  const currencies = ['USD', 'EUR', 'GBP', 'INR'];
  const types = ['debit', 'credit'];
  const categories = ['groceries', 'utilities', 'entertainment', 'transport'];
  const merchants = ['Walmart', 'Amazon', 'Netflix', 'Uber'];
  
  return {
    idempotency_key: `test_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    account_id: `acc_${Math.floor(Math.random() * 10000)}`,
    user_id: `user_${Math.floor(Math.random() * 100000)}`,
    amount: parseFloat((Math.random() * 1000).toFixed(2)),
    currency: currencies[Math.floor(Math.random() * currencies.length)],
    type: types[Math.floor(Math.random() * types.length)],
    category: categories[Math.floor(Math.random() * categories.length)],
    merchant: merchants[Math.floor(Math.random() * merchants.length)],
    reference: `ref_${Math.random().toString(36).substr(2, 9)}`,
    metadata: {
      source: 'load_test',
      test_run: 'barclays_grade',
      timestamp: new Date().toISOString(),
    },
  };
}

// Generate JWT token for authentication
function generateToken() {
  const tokenPayload = {
    user_id: 'load_test_user',
    account_id: 'load_test_account',
    roles: ['teller'],
  };
  
  // Note: In a real scenario, you'd get this from your auth endpoint
  // For load testing, we'll use a pre-generated token
  return 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoibG9hZF90ZXN0X3VzZXIiLCJhY2NvdW50X2lkIjoibG9hZF90ZXN0X2FjY291bnQiLCJyb2xlcyI6WyJ0ZWxsZXIiXSwiZXhwIjoxNzM1NjgwMDAwLCJpYXQiOjE3MzU2NzY0MDAsIm5iZiI6MTczNTY3NjQwMH0.example_signature';
}

// Main test function
export default function () {
  const baseUrl = __ENV.BASE_URL || 'http://localhost:8080';
  const token = generateToken();
  
  // Generate transaction data
  const transaction = generateTransaction();
  
  // Set headers
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
    'Idempotency-Key': transaction.idempotency_key,
  };
  
  // Make request to ingest transaction
  const response = http.post(
    `${baseUrl}/api/v1/transactions`,
    JSON.stringify(transaction),
    { headers }
  );
  
  // Check response
  const success = check(response, {
    'status is 202': (r) => r.status === 202,
    'response has transaction id': (r) => r.json('id') !== undefined,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'idempotency cache header present': (r) => r.headers['X-Idempotency-Cache'] !== undefined,
  });
  
  // Record error rate
  errorRate.add(!success);
  
  // Add some randomness to simulate real-world usage
  sleep(Math.random() * 0.1); // Random sleep between 0-100ms
}

// Setup function to test authentication endpoint
export function setup() {
  const baseUrl = __ENV.BASE_URL || 'http://localhost:8080';
  
  // Test health endpoint
  const healthResponse = http.get(`${baseUrl}/health`);
  check(healthResponse, {
    'health check passed': (r) => r.status === 200,
  });
  
  // Test metrics endpoint
  const metricsResponse = http.get(`${baseUrl}/metrics`);
  check(metricsResponse, {
    'metrics endpoint accessible': (r) => r.status === 200,
  });
  
  console.log('Setup completed. Load test ready to begin.');
}

// Teardown function
export function teardown(data) {
  console.log('Load test completed. Finalizing...');
  
  // You could add cleanup logic here if needed
  // For example, cleaning up test data, etc.
}

// Handle test events
export function handleSummary(data) {
  console.log('Test Summary:');
  console.log(`Total requests: ${data.metrics.http_reqs.values.count}`);
  console.log(`Average response time: ${data.metrics.http_req_duration.values.avg}ms`);
  console.log(`95th percentile: ${data.metrics.http_req_duration.values['p(95)']}ms`);
  console.log(`Error rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%`);
  console.log(`Throughput: ${data.metrics.http_req_rate.values.rate.toFixed(2)} req/s`);
  
  return {
    'load-test-summary.json': JSON.stringify(data, null, 2),
  };
}
