# Lucerna

[Wails](https://wails.io/)와 [Svelte](https://svelte.dev/)로 제작된 현대적인 크로스 플랫폼 Git GUI 클라이언트입니다. [Fork](https://git-fork.com/)에서 영감을 받았습니다.

[English](./README.md)

## 스크린샷

![Lucerna](./assets/screenshot.png)

## 기능

- **저장소 관리** — 로컬 저장소 열기, 실시간 상태 모니터링, 최근 저장소 목록
- **커밋 히스토리 & 시각화** — 페이지네이션이 있는 타임라인, SVG 기반 커밋 그래프, 파일별 커밋 히스토리, git blame 뷰
- **Diff 뷰어** — 통합 diff 뷰, 파일 통계 (추가/삭제), 구문 인식 표시
- **Staging & 커밋** — 개별 파일 또는 전체 변경사항 Stage/Unstage, 검증이 포함된 커밋 생성
- **브랜치 관리** — 브랜치 생성, 삭제, 이름 변경, 체크아웃, 머지; 임의의 커밋에서 브랜치 생성
- **원격 작업** — Fetch, Pull, Push (강제 푸시 옵션), 다중 원격 저장소 관리, 자격 증명 저장
- **Stash** — 생성 (추적되지 않은 파일 포함 옵션), 적용, Pop, 삭제, 전체 제거
- **Tag** — Annotated/Lightweight 태그 생성, 목록, 삭제
- **머지 충돌 해결** — 충돌 해결을 위한 3-way diff 에디터
- **Cherry-pick & Revert** — 개별 커밋 Cherry-pick 및 Revert

## 기술 스택

| 레이어 | 기술 |
|--------|------|
| 백엔드 | Go 1.24, Wails v2 |
| 프론트엔드 | Svelte 4, Vite 5 |
| Git | go-git v5 + 네이티브 git CLI |
| 플랫폼 | Linux, macOS, Windows |

## 시작하기

### 필수 요구사항

- [Go](https://go.dev/) 1.24+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2

**Linux만 해당:**
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### 개발

```bash
# 저장소 클론
git clone https://github.com/yourusername/lucerna.git
cd lucerna

# 프론트엔드 의존성 설치
cd frontend && npm install --legacy-peer-deps && cd ..

# 개발 모드로 실행
wails dev
```

### 빌드

```bash
# 현재 플랫폼용 빌드
wails build

# Linux (webkit2gtk-4.1 지원)
wails build -tags webkit2_41

# macOS (Universal Binary)
wails build -platform darwin/universal

# Windows
wails build -platform windows/amd64
```

자세한 빌드 지침은 [BUILD.md](./BUILD.md)를 참조하세요.

## 다운로드

사전 빌드된 바이너리는 [Releases](https://github.com/yourusername/lucerna/releases) 페이지에서 다운로드할 수 있습니다:

| 플랫폼 | 파일 |
|--------|------|
| Linux | `lucerna-{version}-linux-amd64.tar.gz` |
| macOS | `lucerna-{version}-darwin-amd64.tar.gz` (Universal Binary) |
| Windows | `lucerna-{version}-windows-amd64.zip` |

## 프로젝트 구조

```
lucerna/
├── main.go                          # Wails 애플리케이션 진입점
├── app.go                           # App 메서드 (Wails 바인딩 / API)
├── internal/
│   ├── git/                         # Git 작업
│   │   ├── repository.go            # 저장소 관리
│   │   ├── commit.go                # 커밋 작업
│   │   ├── branch.go                # 브랜치 관리
│   │   ├── diff.go                  # Diff 생성
│   │   ├── stage.go                 # Staging & 커밋 생성
│   │   ├── remote.go                # 원격 작업
│   │   ├── merge.go                 # 머지 & 충돌 해결
│   │   ├── stash.go                 # Stash 작업
│   │   ├── tag.go                   # Tag 작업
│   │   ├── history.go               # 파일 히스토리 & blame
│   │   └── credentials.go           # 자격 증명 저장 (AES-GCM 암호화)
│   └── models/                      # 데이터 모델
├── frontend/                        # Svelte 프론트엔드
│   └── src/
│       ├── App.svelte               # 메인 애플리케이션
│       ├── lib/
│       │   ├── components/          # UI 컴포넌트
│       │   └── stores/              # Svelte 스토어 (상태 관리)
│       └── main.js
├── assets/                          # 앱 아이콘
├── .github/workflows/               # CI/CD (빌드 + 릴리스)
├── wails.json                       # Wails 설정
└── install.sh                       # Linux 설치 스크립트
```

## CI/CD

이 프로젝트는 자동 빌드를 위해 GitHub Actions를 사용합니다:

- **빌드**: `main` 브랜치로 push/PR 시 트리거 — Linux, macOS, Windows용 빌드
- **릴리스**: `v*.*.*` 태그 시 트리거 — 모든 플랫폼 바이너리가 포함된 GitHub Release 생성

## 기여하기

기여를 환영합니다! Pull Request를 자유롭게 제출해주세요.

1. 저장소를 Fork합니다
2. 기능 브랜치를 생성합니다 (`git checkout -b feature/amazing-feature`)
3. 변경사항을 커밋합니다 (`git commit -m 'Add some amazing feature'`)
4. 브랜치에 Push합니다 (`git push origin feature/amazing-feature`)
5. Pull Request를 생성합니다

## 라이선스

이 프로젝트는 MIT 라이선스에 따라 배포됩니다. 자세한 내용은 [LICENSE](./LICENSE) 파일을 참조하세요.

---

> **"주의 말씀은 내 발에 등이요 내 길에 빛이니이다"**
>
> *— 시편 119:105 (개역개정)*
