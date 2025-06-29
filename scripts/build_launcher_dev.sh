#!/bin/bash

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${PROJECT_ROOT}/launcher"

echo "Building DreamHub Launcher..."
echo ""

# Check if nosystray flag is passed
if [[ "$1" == "nosystray" ]]; then
    echo "Building without systray support (CLI only)..."
    go build -tags=nosystray -o "${PROJECT_ROOT}/dreamhub" ./cmd/dreamhub
else
    echo "Building with systray support..."
    echo ""
    echo "Note: On Linux, systray requires the following packages to be installed:"
    echo "  sudo apt-get install gcc libgtk-3-dev libayatana-appindicator3-dev"
    echo ""
    echo "Attempting to build..."
    go build -o "${PROJECT_ROOT}/dreamhub" ./cmd/dreamhub
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "Build successful! Binary created at ${PROJECT_ROOT}/dreamhub"
    echo ""
    echo "Usage:"
    echo "  ./dreamhub         - Start with system tray (requires dependencies)"
    echo "  ./dreamhub status  - Check PCAS status (works without dependencies)"
else
    echo ""
    echo "Build failed. This is likely due to missing system dependencies."
    echo "Please install the required packages mentioned above."
    echo ""
    echo "Alternatively, build without systray support:"
    echo "  ./build.sh nosystray"
fi