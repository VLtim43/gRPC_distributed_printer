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
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run server/main.go" &

sleep 2

echo "Starting client..."
$TERM_CMD bash -c "cd $(pwd) && trap 'kill 0' EXIT; go run client/main.go; read -p 'Press enter to close...'" &
