#!/bin/bash

echo "🚀 Pulling latest Docker images..."

# Pull the latest images
docker pull ghcr.io/antinvestor/service-payment/default:latest
docker pull ghcr.io/antinvestor/service-payment/jenga-api:latest

echo "✅ Images pulled successfully!"

echo "🐳 Starting services with docker-compose..."
docker-compose up -d

echo "📊 Service status:"
docker-compose ps

echo ""
echo "🔗 Service URLs:"
echo "  Payment Service: http://localhost:8081"
echo "  Jenga API: http://localhost:8082"
echo "  gRPC Service: localhost:50051"
echo ""
echo "📝 To view logs: docker-compose logs -f"
echo "🛑 To stop: docker-compose down"
