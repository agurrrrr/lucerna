# Lucerna 빌드하기

이 가이드는 여러 플랫폼에서 Lucerna를 빌드하는 방법을 설명합니다.

## 필수 요구사항

- Go 1.21 이상
- Node.js 18 이상
- Wails CLI v2

### 플랫폼별 요구사항

#### Linux
```bash
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
```

#### macOS
Xcode Command Line Tools 외에 추가 의존성이 필요하지 않습니다.

#### Windows
표준 빌드 도구 외에 추가 의존성이 필요하지 않습니다.

## Wails CLI 설치

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 빌드하기

### 개발 빌드

```bash
# 프론트엔드 의존성 설치
cd frontend
npm install --legacy-peer-deps
cd ..

# 개발 모드로 실행
wails dev
```

### 프로덕션 빌드

#### 현재 플랫폼용 빌드

```bash
wails build
```

바이너리는 `build/bin/`에 생성됩니다.

#### 특정 플랫폼용 빌드

```bash
# Linux
wails build -platform linux

# macOS (Universal Binary)
wails build -platform darwin

# Windows
wails build -platform windows
```

#### 클린 빌드

```bash
wails build -clean
```

## 빌드 결과물

- **Linux**: `build/bin/lucerna`
- **macOS**: `build/bin/Lucerna.app`
- **Windows**: `build/bin/lucerna.exe`

## GitHub Actions

이 프로젝트는 GitHub Actions를 통해 모든 플랫폼에 대한 자동 빌드를 지원합니다:

### 지속적 통합 (`build.yml`)
- main/master 브랜치로 push/PR 시 트리거
- Linux, Windows, macOS용 빌드
- 빌드 아티팩트 업로드

### 릴리스 (`release.yml`)
- 버전 태그 시 트리거 (예: `v1.0.0`)
- 모든 플랫폼용 빌드
- 바이너리와 함께 GitHub 릴리스 생성
- 릴리스 노트 자동 생성

## 릴리스 만들기

1. `wails.json`에서 버전 업데이트
2. 변경사항 커밋
3. 태그 생성 및 푸시:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. GitHub Actions가 자동으로 빌드하고 릴리스를 생성합니다

## 문제 해결

### Linux

GTK 오류가 발생하는 경우:
```bash
sudo apt-get install -y libgtk-3-0 libwebkit2gtk-4.0-37
```

### macOS

서명 오류가 발생하는 경우, 코드 서명을 비활성화할 수 있습니다:
```bash
wails build -skipbindings -s
```

### Windows

NSIS 인스톨러가 실패하는 경우, NSIS가 설치되어 있는지 확인하거나 인스톨러 없이 빌드하세요:
```bash
wails build -nsis=false
```

## 빌드 구성

빌드 구성은 `wails.json`에서 관리됩니다:
- 애플리케이션 메타데이터
- 아이콘 경로
- 플랫폼별 설정
- NSIS 인스톨러 구성 (Windows)
- Debian 의존성 (Linux)
- Universal binary 지원 (macOS)

## 배포

### Linux
- 바이너리: `lucerna`
- `wails build -deb`를 통한 Debian 패키지 지원

### macOS
- Universal 앱 번들: `Lucerna.app`
- Intel 및 Apple Silicon 모두 지원

### Windows
- 실행 파일: `lucerna.exe`
- `wails build -nsis`를 통한 NSIS 인스톨러 지원
