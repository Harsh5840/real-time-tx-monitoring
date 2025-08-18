#!/bin/bash
set -e

echo "Creating Kafka topics for rtMonitor..."

docker exec -it tx_kafka kafka-topics \
  --create \
  --topic transactions \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1 || echo "Topic 'transactions' already exists."

docker exec -it tx_kafka kafka-topics \
  --create \
  --topic alerts \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1 || echo "Topic 'alerts' already exists."

echo "✅ Topics created (transactions, alerts)"
