#!/bin/bash

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

echo "Starting server..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run server/main.go; read -p 'Press enter to close...'" &

sleep 2

echo "Starting client A (port 50052)..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=A -port=50052 -peers=localhost:50053,localhost:50054; read -p 'Press enter to close...'" &

sleep 1

echo "Starting client B (port 50053)..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=B -port=50053 -peers=localhost:50052,localhost:50054; read -p 'Press enter to close...'" &

sleep 1

echo "Starting client C (port 50054)..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/*.go -id=C -port=50054 -peers=localhost:50052,localhost:50053; read -p 'Press enter to close...'" &

echo ""
echo "Started 1 server and 3 clients"
echo "Press Ctrl+C to stop all processes"
