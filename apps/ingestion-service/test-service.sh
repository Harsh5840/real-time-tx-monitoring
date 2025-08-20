#!/bin/bash

# Barclays-Grade Ingestion Service Test Script
# This script tests the key features of the enhanced ingestion service

set -e

BASE_URL=${1:-"http://localhost:8080"}
SERVICE_NAME="Barclays-Grade Ingestion Service"

echo "🧪 Testing $SERVICE_NAME at $BASE_URL"
echo "=========================================="

# Function to check if service is running
check_service() {
    echo "🔍 Checking if service is running..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        echo "✅ Service is running"
        return 0
    else
        echo "❌ Service is not running"
        return 1
    fi
}

# Function to test health endpoint
test_health() {
    echo "🏥 Testing health endpoint..."
    response=$(curl -s "$BASE_URL/health")
    if echo "$response" | grep -q "healthy"; then
        echo "✅ Health check passed"
    else
        echo "❌ Health check failed: $response"
        return 1
    fi
}

# Function to test metrics endpoint
test_metrics() {
    echo "📊 Testing metrics endpoint..."
    if curl -s "$BASE_URL/metrics" | grep -q "http_requests_total"; then
        echo "✅ Metrics endpoint working"
    else
        echo "❌ Metrics endpoint not working"
        return 1
    fi
}

# Function to test JWT token generation
test_auth() {
    echo "🔐 Testing JWT authentication..."
    
    # Generate a token
    token_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/token" \
        -H "Content-Type: application/json" \
        -d '{"user_id":"test_user","account_id":"acc_123","roles":["teller"]}')
    
    if echo "$token_response" | grep -q "token"; then
        echo "✅ JWT token generation working"
        TOKEN=$(echo "$token_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "   Token: ${TOKEN:0:50}..."
    else
        echo "❌ JWT token generation failed: $token_response"
        return 1
    fi
}

# Function to test transaction ingestion
test_transaction_ingestion() {
    echo "💳 Testing transaction ingestion..."
    
    if [ -z "$TOKEN" ]; then
        echo "❌ No JWT token available"
        return 1
    fi
    
    # Test single transaction
    txn_response=$(curl -s -X POST "$BASE_URL/api/v1/transactions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Idempotency-Key: test_key_$(date +%s)" \
        -d '{
            "idempotency_key": "test_key_$(date +%s)",
            "account_id": "acc_123",
            "user_id": "user_456",
            "amount": 99.99,
            "currency": "USD",
            "type": "debit",
            "category": "groceries",
            "merchant": "Walmart"
        }')
    
    if echo "$txn_response" | grep -q "accepted"; then
        echo "✅ Single transaction ingestion working"
    else
        echo "❌ Single transaction ingestion failed: $txn_response"
        return 1
    fi
}

# Function to test idempotency
test_idempotency() {
    echo "🔄 Testing idempotency..."
    
    if [ -z "$TOKEN" ]; then
        echo "❌ No JWT token available"
        return 1
    fi
    
    # Use the same idempotency key
    idempotency_key="idempotency_test_$(date +%s)"
    
    # First request
    first_response=$(curl -s -X POST "$BASE_URL/api/v1/transactions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Idempotency-Key: $idempotency_key" \
        -d "{
            \"idempotency_key\": \"$idempotency_key\",
            \"account_id\": \"acc_123\",
            \"user_id\": \"user_456\",
            \"amount\": 50.00,
            \"currency\": \"USD\",
            \"type\": \"debit\",
            \"category\": \"utilities\"
        }")
    
    # Second request with same key
    second_response=$(curl -s -X POST "$BASE_URL/api/v1/transactions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Idempotency-Key: $idempotency_key" \
        -d "{
            \"idempotency_key\": \"$idempotency_key\",
            \"account_id\": \"acc_123\",
            \"user_id\": \"user_456\",
            \"amount\": 50.00,
            \"currency\": \"USD\",
            \"type\": \"debit\",
            \"category\": \"utilities\"
        }")
    
    # Check if second response has idempotency cache header
    if echo "$second_response" | grep -q "idempotency_cache"; then
        echo "✅ Idempotency working (cached response)"
    else
        echo "❌ Idempotency not working properly"
        return 1
    fi
}

# Function to test batch ingestion
test_batch_ingestion() {
    echo "📦 Testing batch transaction ingestion..."
    
    if [ -z "$TOKEN" ]; then
        echo "❌ No JWT token available"
        return 1
    fi
    
    # Test batch transactions
    batch_response=$(curl -s -X POST "$BASE_URL/api/v1/transactions/batch" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Idempotency-Key: batch_test_$(date +%s)" \
        -d '[
            {
                "idempotency_key": "batch_1_$(date +%s)",
                "account_id": "acc_123",
                "user_id": "user_456",
                "amount": 25.00,
                "currency": "USD",
                "type": "debit",
                "category": "entertainment"
            },
            {
                "idempotency_key": "batch_2_$(date +%s)",
                "account_id": "acc_123",
                "user_id": "user_456",
                "amount": 75.00,
                "currency": "USD",
                "type": "credit",
                "category": "salary"
            }
        ]')
    
    if echo "$batch_response" | grep -q "accepted"; then
        echo "✅ Batch transaction ingestion working"
    else
        echo "❌ Batch transaction ingestion failed: $batch_response"
        return 1
    fi
}

# Function to test unauthorized access
test_unauthorized_access() {
    echo "🚫 Testing unauthorized access..."
    
    # Try to access without token
    unauthorized_response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/api/v1/transactions" \
        -H "Content-Type: application/json" \
        -d '{"test": "data"}' -o /dev/null)
    
    if [ "$unauthorized_response" = "401" ]; then
        echo "✅ Unauthorized access properly blocked"
    else
        echo "❌ Unauthorized access not properly blocked: HTTP $unauthorized_response"
        return 1
    fi
}

# Function to test invalid idempotency key
test_invalid_idempotency() {
    echo "⚠️  Testing invalid idempotency key..."
    
    if [ -z "$TOKEN" ]; then
        echo "❌ No JWT token available"
        return 1
    fi
    
    # Try to access without idempotency key
    invalid_response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/api/v1/transactions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "account_id": "acc_123",
            "user_id": "user_456",
            "amount": 50.00,
            "currency": "USD",
            "type": "debit",
            "category": "utilities"
        }' -o /dev/null)
    
    if [ "$invalid_response" = "400" ]; then
        echo "✅ Invalid idempotency key properly rejected"
    else
        echo "❌ Invalid idempotency key not properly rejected: HTTP $invalid_response"
        return 1
    fi
}

# Main test execution
main() {
    echo "Starting comprehensive test suite..."
    echo ""
    
    # Run all tests
    check_service || exit 1
    echo ""
    
    test_health || exit 1
    echo ""
    
    test_metrics || exit 1
    echo ""
    
    test_auth || exit 1
    echo ""
    
    test_transaction_ingestion || exit 1
    echo ""
    
    test_idempotency || exit 1
    echo ""
    
    test_batch_ingestion || exit 1
    echo ""
    
    test_unauthorized_access || exit 1
    echo ""
    
    test_invalid_idempotency || exit 1
    echo ""
    
    echo "🎉 All tests passed! $SERVICE_NAME is working correctly."
    echo ""
    echo "📊 Check metrics at: $BASE_URL/metrics"
    echo "🏥 Health status at: $BASE_URL/health"
    echo ""
    echo "🚀 Ready for production load testing!"
}

# Run main function
main "$@"
