#!/bin/bash

# Script to run Wastebin locally with both backend and frontend

echo "Starting Wastebin in local development mode..."

# Set environment variables for local development
export WASTEBIN_LOCAL_DB=true
export WASTEBIN_DEV=true
export WASTEBIN_WEBAPP_PORT=3000
export WASTEBIN_TRACING_ENABLED=false
export WASTEBIN_METRICS_ENABLED=false

# Build the frontend first
echo "Building frontend..."
cd web && npm run build
if [ $? -ne 0 ]; then
    echo "Frontend build failed!"
    exit 1
fi
cd ..

# Function to cleanup on exit
cleanup() {
    echo -e "\n\nShutting down services..."
    kill $BACKEND_PID 2>/dev/null
    exit 0
}

# Set up trap for cleanup
trap cleanup INT TERM EXIT

# Build the backend
echo "Building backend..."
go build -o wastebin-server ./cmd/wastebin
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi

# Start the backend server (now serves both API and frontend)
echo "Starting Wastebin server on port 3000..."
./wastebin-server &
BACKEND_PID=$!

echo -e "\n\n========================================="
echo "Wastebin is running!"
echo "Application URL: http://localhost:3000"
echo "API Endpoint: http://localhost:3000/api/v1"
echo "Press Ctrl+C to stop"
echo "========================================="

# Wait for process
wait