#!/bin/bash

# Build script for Event Storage System
# This script builds all Docker images with consistent naming

set -e

echo "Building Event Storage System Docker images..."

# Build log-push
echo "Building log-push..."
cd logPush
docker build -t ess/log-push:latest .
cd ..

# Build log-persist
echo "Building log-persist..."
cd log-persist-v2
docker build -t ess/log-persist:latest .
cd ..

# Build log-search
echo "Building log-search..."
cd logSearch
docker build -t ess/log-search:latest .
cd ..

# Build lambda (LRE)
echo "Building lambda (LRE)..."
cd lambda
docker build -t ess/log-monitor-lambda:latest .
cd ..

echo "All images built successfully!"
echo ""
echo "Built images:"
echo "- ess/log-push:latest"
echo "- ess/log-persist:latest" 
echo "- ess/log-search:latest"
echo "- ess/log-monitor-lambda:latest"
echo ""
echo "To start the system, run:"
echo "cd infra && docker-compose up -d"
