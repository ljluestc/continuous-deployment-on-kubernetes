#!/bin/bash
# Script to improve test coverage for all Go services

set -e

SERVICES=(
    "dns"
    "webcrawler"
    "newsfeed"
    "loadbalancer"
    "tinyurl"
    "typeahead"
    "sample-app"
)

echo "=========================================="
echo "Improving coverage for all services"
echo "=========================================="
echo

for service in "${SERVICES[@]}"; do
    echo "Processing $service..."
    if [ "$service" = "sample-app" ]; then
        cd "/home/calelin/dev/continuous-deployment-on-kubernetes/sample-app"
    else
        cd "/home/calelin/dev/continuous-deployment-on-kubernetes/services/$service"
    fi
    
    echo "Running tests for $service..."
    if go test -tags=unit -v -coverprofile=coverage.out ./... 2>&1; then
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo "✅ $service: $coverage coverage"
    else
        echo "❌ $service: Tests failed"
    fi
    echo
done

echo "=========================================="
echo "Coverage improvement complete!"
echo "=========================================="

