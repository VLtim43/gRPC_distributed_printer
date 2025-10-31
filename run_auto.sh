#!/bin/bash

# Script to run the distributed print system with automatic print request generation

# List of terminals: "command:args"
TERMINALS=(
    "alacritty:-e"
    "kitty:-e"
    "gnome-terminal:--"
    "x-terminal-emulator:-e"
    "xterm:-e"
)

TERM_CMD=""
for term in "${TERMINALS[@]}"; do
    cmd="${term%%:*}"
    args="${term#*:}"
    if command -v "$cmd" &> /dev/null; then
        TERM_CMD="$cmd $args"
        break
    fi
done

if [ -z "$TERM_CMD" ]; then
    echo "No supported terminal emulator found"
    exit 1
fi

# Configuration
INTERVAL=${1:-5}  # Default 5 seconds interval

echo "Starting distributed print system with AUTOMATIC mode"
echo "Print request interval: $INTERVAL seconds"
echo ""

echo "Starting server..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run server/main.go; read -p 'Press enter to close...'" &

sleep 2

echo "Starting client A (port 50052) in AUTOMATIC mode..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=A -port=50052 -peers=localhost:50053,localhost:50054 -auto -interval=$INTERVAL; read -p 'Press enter to close...'" &

sleep 1

echo "Starting client B (port 50053) in AUTOMATIC mode..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=B -port=50053 -peers=localhost:50052,localhost:50054 -auto -interval=$INTERVAL; read -p 'Press enter to close...'" &

sleep 1

echo "Starting client C (port 50054) in AUTOMATIC mode..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=C -port=50054 -peers=localhost:50052,localhost:50053 -auto -interval=$INTERVAL; read -p 'Press enter to close...'" &

echo ""
echo "Started 1 server and 3 clients in AUTOMATIC mode"
echo "Each client will generate print requests every $INTERVAL seconds"
echo "Watch the terminals to see the Ricart-Agrawala mutual exclusion in action!"
echo ""
echo "Press Ctrl+C to stop all processes"
