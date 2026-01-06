#!/bin/bash
# Example health check script for TaskFlow

TASKFLOW_URL="${TASKFLOW_URL:-http://localhost:8080}"

echo "Checking TaskFlow health..."

response=$(curl -s -w "\n%{http_code}" "$TASKFLOW_URL/health")
http_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" -eq 200 ]; then
    echo "✓ TaskFlow is healthy"
    echo "$body" | jq . || echo "$body"
    exit 0
else
    echo "✗ TaskFlow is unhealthy (HTTP $http_code)"
    echo "$body"
    exit 1
fi
