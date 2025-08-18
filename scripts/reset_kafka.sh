#!/bin/bash

echo "🔄 Resetting Kafka topic: transactions"

# Delete topic if exists
docker exec -it tx_kafka kafka-topics \
  --delete \
  --topic transactions \
  --bootstrap-server localhost:9092 || true

# Recreate topic
docker exec -it tx_kafka kafka-topics \
  --create \
  --topic transactions \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1

echo "✅ Kafka topic 'transactions' reset."
