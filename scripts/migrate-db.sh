#!/bin/bash
set -e

echo "Running DB migrations on rtMonitor (tx_db)..."

docker exec -i tx_postgres psql -U tx_user -d tx_db <<EOF
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(id),
    alert_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
EOF

echo "✅ Migrations applied (transactions, alerts tables)"
