#!/bin/bash
echo "🛑 Stopping services..."
docker-compose -f infra/docker/docker-compose.yml down
