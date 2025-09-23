#!/bin/bash
echo "Starting Corridor OS on Ubuntu..."
echo "Opening web interface..."

# Try different browsers in order of preference
if command -v google-chrome > /dev/null; then
    google-chrome --new-window "file://$(pwd)/LAUNCH-CORRIDOR-OS.html"
elif command -v firefox > /dev/null; then
    firefox --new-window "file://$(pwd)/LAUNCH-CORRIDOR-OS.html"
elif command -v chromium-browser > /dev/null; then
    chromium-browser --new-window "file://$(pwd)/LAUNCH-CORRIDOR-OS.html"
elif command -v xdg-open > /dev/null; then
    xdg-open "LAUNCH-CORRIDOR-OS.html"
else
    echo "Please open LAUNCH-CORRIDOR-OS.html in your web browser"
fi
