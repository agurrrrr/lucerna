#!/bin/bash
# Lucerna installer for Arch Linux
# Usage: ./install.sh

set -e

echo "=== Lucerna 설치 스크립트 ==="

# Check dependencies
echo "[1/5] 의존성 확인..."
for cmd in go node npm wails; do
    if ! command -v $cmd &> /dev/null; then
        echo "오류: $cmd 가 설치되어 있지 않습니다."
        echo ""
        echo "필수 패키지 설치:"
        echo "  sudo pacman -S go nodejs npm gtk3 webkit2gtk"
        echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        exit 1
    fi
done

# Install frontend dependencies
echo "[2/5] 프론트엔드 의존성 설치..."
cd frontend
npm install --legacy-peer-deps
cd ..

# Build
echo "[3/5] 빌드 중..."
wails build -clean

# Install binary
echo "[4/5] 바이너리 설치..."
sudo install -Dm755 build/bin/lucerna /usr/bin/lucerna

# Install icons
echo "[5/6] 아이콘 설치..."
sudo install -Dm644 assets/lucerna.svg /usr/share/icons/hicolor/scalable/apps/lucerna.svg
# Convert SVG to PNG for better compatibility
if command -v rsvg-convert &> /dev/null; then
    rsvg-convert -w 256 -h 256 assets/lucerna.svg -o /tmp/lucerna-256.png
    sudo install -Dm644 /tmp/lucerna-256.png /usr/share/icons/hicolor/256x256/apps/lucerna.png
    rsvg-convert -w 128 -h 128 assets/lucerna.svg -o /tmp/lucerna-128.png
    sudo install -Dm644 /tmp/lucerna-128.png /usr/share/icons/hicolor/128x128/apps/lucerna.png
    rsvg-convert -w 64 -h 64 assets/lucerna.svg -o /tmp/lucerna-64.png
    sudo install -Dm644 /tmp/lucerna-64.png /usr/share/icons/hicolor/64x64/apps/lucerna.png
    rsvg-convert -w 48 -h 48 assets/lucerna.svg -o /tmp/lucerna-48.png
    sudo install -Dm644 /tmp/lucerna-48.png /usr/share/icons/hicolor/48x48/apps/lucerna.png
    rm -f /tmp/lucerna-*.png
elif command -v convert &> /dev/null; then
    convert -background none -resize 256x256 assets/lucerna.svg /tmp/lucerna-256.png
    sudo install -Dm644 /tmp/lucerna-256.png /usr/share/icons/hicolor/256x256/apps/lucerna.png
    convert -background none -resize 128x128 assets/lucerna.svg /tmp/lucerna-128.png
    sudo install -Dm644 /tmp/lucerna-128.png /usr/share/icons/hicolor/128x128/apps/lucerna.png
    rm -f /tmp/lucerna-*.png
fi

# Install desktop file
echo "[6/6] 데스크톱 파일 설치..."
sudo tee /usr/share/applications/lucerna.desktop > /dev/null << EOF
[Desktop Entry]
Name=Lucerna
Comment=A modern Git GUI application
Exec=lucerna
Icon=lucerna
Terminal=false
Type=Application
Categories=Development;RevisionControl;
Keywords=git;version control;vcs;
StartupWMClass=lucerna
EOF

# Update icon cache
gtk-update-icon-cache -f /usr/share/icons/hicolor 2>/dev/null || true

echo ""
echo "=== 설치 완료! ==="
echo "실행: lucerna"
