#!/bin/bash
# Lucerna installer for Arch Linux
# Usage: ./install.sh [--yes]

set -e

AUTO_YES=0
if [[ "$1" == "--yes" || "$1" == "-y" ]]; then
    AUTO_YES=1
fi

echo "=== Lucerna 설치 스크립트 ==="

# pacman 패키지 존재 여부 확인
pkg_installed() {
    pacman -Qi "$1" &> /dev/null
}

confirm() {
    local prompt="$1"
    if [[ $AUTO_YES -eq 1 ]]; then
        return 0
    fi
    read -r -p "$prompt [Y/n] " response
    case "$response" in
        ""|y|Y|yes|YES) return 0 ;;
        *) return 1 ;;
    esac
}

# [1/6] 의존성 확인 및 설치
echo "[1/6] 의존성 확인..."

# pacman 패키지: command 이름 → 패키지 이름 매핑
declare -A CMD_TO_PKG=(
    [go]=go
    [node]=nodejs
    [npm]=npm
    [pkg-config]=pkgconf
)

# 빌드용 라이브러리 패키지 (command로 확인 불가)
LIB_PKGS=(gtk3 webkit2gtk-4.1)

MISSING_PKGS=()

for cmd in "${!CMD_TO_PKG[@]}"; do
    if ! command -v "$cmd" &> /dev/null; then
        MISSING_PKGS+=("${CMD_TO_PKG[$cmd]}")
    fi
done

for pkg in "${LIB_PKGS[@]}"; do
    if ! pkg_installed "$pkg"; then
        # webkit2gtk-4.1이 없으면 webkit2gtk(구버전)도 체크
        if [[ "$pkg" == "webkit2gtk-4.1" ]] && pkg_installed "webkit2gtk"; then
            continue
        fi
        MISSING_PKGS+=("$pkg")
    fi
done

# 아이콘 변환용 (선택적이지만 권장)
if ! command -v rsvg-convert &> /dev/null && ! command -v convert &> /dev/null; then
    MISSING_PKGS+=(librsvg)
fi

if [[ ${#MISSING_PKGS[@]} -gt 0 ]]; then
    echo "다음 패키지가 누락되어 있습니다: ${MISSING_PKGS[*]}"
    if confirm "지금 sudo pacman으로 설치할까요?"; then
        sudo pacman -S --needed "${MISSING_PKGS[@]}"
    else
        echo "설치를 중단합니다. 수동으로 설치 후 다시 실행해 주세요:"
        echo "  sudo pacman -S --needed ${MISSING_PKGS[*]}"
        exit 1
    fi
fi

# wails 확인 및 설치
if ! command -v wails &> /dev/null; then
    # GOPATH/bin 도 확인
    GOBIN_PATH="$(go env GOPATH 2>/dev/null)/bin"
    if [[ -x "$GOBIN_PATH/wails" ]]; then
        export PATH="$GOBIN_PATH:$PATH"
    else
        echo "wails 가 설치되어 있지 않습니다."
        if confirm "go install 로 wails 를 설치할까요?"; then
            go install github.com/wailsapp/wails/v2/cmd/wails@latest
            export PATH="$GOBIN_PATH:$PATH"
            if ! command -v wails &> /dev/null; then
                echo "오류: wails 설치 후에도 PATH 에서 찾을 수 없습니다."
                echo "  $GOBIN_PATH 를 PATH 에 추가해 주세요."
                exit 1
            fi
        else
            echo "wails 없이는 빌드할 수 없습니다. 종료합니다."
            exit 1
        fi
    fi
fi

# webkit2gtk 버전 감지 (Arch는 4.0을 4.1로 대체함)
# 4.1을 쓰면 wails 빌드에 webkit2_41 태그가 필요하다.
WAILS_TAGS=""
if pkg-config --exists webkit2gtk-4.0; then
    echo "  webkit2gtk-4.0 감지됨"
elif pkg-config --exists webkit2gtk-4.1; then
    echo "  webkit2gtk-4.1 감지됨 (webkit2_41 태그 사용)"
    WAILS_TAGS="-tags webkit2_41"
else
    echo "오류: webkit2gtk-4.0 또는 4.1 이 설치되어 있지 않습니다."
    echo "  sudo pacman -S webkit2gtk-4.1"
    exit 1
fi

# [2/6] 프론트엔드 의존성 설치
echo "[2/6] 프론트엔드 의존성 설치..."
cd frontend
npm install --legacy-peer-deps
cd ..

# [3/6] 빌드
echo "[3/6] 빌드 중..."
wails build -clean $WAILS_TAGS

# [4/6] 바이너리 설치
echo "[4/6] 바이너리 설치..."
sudo install -Dm755 build/bin/lucerna /usr/bin/lucerna

# [5/6] 아이콘 설치
echo "[5/6] 아이콘 설치..."
sudo install -Dm644 assets/lucerna.svg /usr/share/icons/hicolor/scalable/apps/lucerna.svg
if command -v rsvg-convert &> /dev/null; then
    for size in 256 128 64 48; do
        rsvg-convert -w $size -h $size assets/lucerna.svg -o /tmp/lucerna-$size.png
        sudo install -Dm644 /tmp/lucerna-$size.png /usr/share/icons/hicolor/${size}x${size}/apps/lucerna.png
    done
    rm -f /tmp/lucerna-*.png
elif command -v convert &> /dev/null; then
    for size in 256 128; do
        convert -background none -resize ${size}x${size} assets/lucerna.svg /tmp/lucerna-$size.png
        sudo install -Dm644 /tmp/lucerna-$size.png /usr/share/icons/hicolor/${size}x${size}/apps/lucerna.png
    done
    rm -f /tmp/lucerna-*.png
fi

# [6/6] 데스크톱 파일 설치
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

# 아이콘 캐시 갱신
gtk-update-icon-cache -f /usr/share/icons/hicolor 2>/dev/null || true

echo ""
echo "=== 설치 완료! ==="
echo "실행: lucerna"
