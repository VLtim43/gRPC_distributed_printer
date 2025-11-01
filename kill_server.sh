#!/bin/bash

# Kill any process using port 50051 i.e the main server
echo "Checking for processes on port 50051..."
PID=$(lsof -ti :50051)

if [ -z "$PID" ]; then
    echo "No process found on port 50051"
else
    echo "Killing process $PID on port 50051..."
    kill $PID
    sleep 1

    # Force kill if still running
    if lsof -ti :50051 > /dev/null 2>&1; then
        echo "Process still running, force killing..."
        kill -9 $PID
    fi

    echo "Port 50051 is now free"
fi
