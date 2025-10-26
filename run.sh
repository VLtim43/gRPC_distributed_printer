#!/bin/bash

if command -v x-terminal-emulator &> /dev/null; then
    TERM_CMD="x-terminal-emulator -e"
elif command -v alacritty &> /dev/null; then
    TERM_CMD="alacritty -e"
elif command -v kitty &> /dev/null; then
    TERM_CMD="kitty -e"
elif command -v gnome-terminal &> /dev/null; then
    TERM_CMD="gnome-terminal --"
elif command -v xterm &> /dev/null; then
    TERM_CMD="xterm -e"
else
    echo "No supported terminal emulator found"
    exit 1
fi

$TERM_CMD bash -c "cd $(pwd) && go run server/main.go; read -p 'Press enter to close...'" &

sleep 2

$TERM_CMD bash -c "cd $(pwd) && go run client/main.go; read -p 'Press enter to close...'" &
