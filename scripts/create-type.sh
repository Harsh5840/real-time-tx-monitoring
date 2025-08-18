#!/bin/bash
echo "🚀 Starting services..."
docker-compose -f infra/docker/docker-compose.yml up -d
