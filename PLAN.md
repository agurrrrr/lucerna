# Lucerna - Git GUI Client 개발 계획

## 기술 스택
- **Backend**: Go (Wails framework)
- **Frontend**: Svelte + JavaScript (No TypeScript)
- **Styling**: CSS / TailwindCSS
- **Git Operations**: go-git 또는 git CLI wrapper

## 프로젝트 구조
```
lucerna/
├── frontend/               # Svelte 프론트엔드
│   ├── src/
│   │   ├── lib/           # 공통 컴포넌트
│   │   │   ├── components/
│   │   │   │   ├── RepositoryList.svelte
│   │   │   │   ├── CommitGraph.svelte
│   │   │   │   ├── CommitHistory.svelte
│   │   │   │   ├── FileTree.svelte
│   │   │   │   ├── DiffViewer.svelte
│   │   │   │   ├── BranchList.svelte
│   │   │   │   ├── Sidebar.svelte
│   │   │   │   └── Toolbar.svelte
│   │   │   └── stores/    # Svelte stores (상태 관리)
│   │   │       ├── repository.js
│   │   │       ├── commits.js
│   │   │       └── branches.js
│   │   ├── routes/        # 라우팅 (SvelteKit 사용 시)
│   │   ├── App.svelte     # 메인 앱
│   │   └── main.js        # 엔트리 포인트
│   ├── public/
│   └── package.json
│
├── backend/               # Go 백엔드
│   ├── cmd/
│   │   └── main.go       # Wails 앱 엔트리
│   ├── internal/
│   │   ├── git/          # Git 작업 핸들러
│   │   │   ├── repository.go
│   │   │   ├── commit.go
│   │   │   ├── branch.go
│   │   │   ├── diff.go
│   │   │   ├── stage.go
│   │   │   └── remote.go
│   │   └── models/       # 데이터 모델
│   │       ├── repository.go
│   │       ├── commit.go
│   │       └── branch.go
│   └── go.mod
│
├── wails.json            # Wails 설정
└── README.md
```

## 개발 단계

### Phase 1: 프로젝트 초기 설정 (Week 1)
1. **Wails 프로젝트 초기화**
   - `wails init -n lucerna -t svelte`
   - JavaScript 설정 (TypeScript 제거)
   - 기본 디렉토리 구조 설정

2. **개발 환경 구성**
   - Go 모듈 설정
   - Svelte 개발 환경 설정
   - ESLint, Prettier 설정 (JavaScript용)

3. **기본 UI 레이아웃**
   - 메인 윈도우 레이아웃
   - 사이드바, 툴바, 메인 콘텐츠 영역 구분

### Phase 2: 저장소 관리 (Week 2-3)
1. **저장소 열기/관리**
   - 로컬 저장소 선택 및 열기
   - 저장소 목록 관리
   - 최근 저장소 목록

2. **Go 백엔드 - 기본 Git 작업**
   - Repository 정보 읽기
   - 저장소 상태 확인
   - 브랜치 목록 조회

3. **UI 컴포넌트**
   - RepositoryList.svelte
   - Sidebar.svelte (저장소 선택)

### Phase 3: 커밋 히스토리 & 그래프 (Week 4-5)
1. **커밋 히스토리**
   - 커밋 로그 가져오기
   - 커밋 상세 정보 표시
   - 커밋 검색/필터링

2. **커밋 그래프 시각화**
   - Git 브랜치 그래프 렌더링
   - Canvas 또는 SVG 기반 그래프
   - 머지, 브랜치 시각화

3. **UI 컴포넌트**
   - CommitHistory.svelte
   - CommitGraph.svelte
   - CommitDetail.svelte

### Phase 4: 파일 변경사항 & Diff (Week 6-7)
1. **파일 트리**
   - Working Directory 파일 트리
   - 변경된 파일 표시
   - Staged/Unstaged 구분

2. **Diff 뷰어**
   - 파일 diff 표시
   - Syntax 하이라이팅
   - Side-by-side / Unified diff

3. **UI 컴포넌트**
   - FileTree.svelte
   - DiffViewer.svelte
   - ChangesList.svelte

### Phase 5: Staging & Commit (Week 8)
1. **스테이징 작업**
   - 파일 stage/unstage
   - Partial staging (hunk 단위)
   - Stage all / Unstage all

2. **커밋 생성**
   - 커밋 메시지 입력
   - 커밋 작성
   - Amend commit

3. **UI 컴포넌트**
   - StagingArea.svelte
   - CommitDialog.svelte

### Phase 6: 브랜치 관리 (Week 9)
1. **브랜치 작업**
   - 브랜치 생성/삭제
   - 브랜치 체크아웃
   - 브랜치 머지
   - 브랜치 리베이스

2. **UI 컴포넌트**
   - BranchList.svelte
   - BranchDialog.svelte
   - MergeDialog.svelte

### Phase 7: 원격 저장소 (Week 10-11)
1. **리모트 작업**
   - Remote 목록
   - Fetch
   - Pull
   - Push
   - Remote 브랜치 관리

2. **UI 컴포넌트**
   - RemoteList.svelte
   - PushPullDialog.svelte

### Phase 8: 추가 기능 (Week 12-13)
1. **Stash**
   - Stash 생성/적용/삭제
   - Stash 목록

2. **Tag**
   - Tag 생성/삭제
   - Tag 목록

3. **History & Blame**
   - 파일 히스토리
   - Git blame

4. **UI 개선**
   - 다크 모드
   - 테마 설정
   - 단축키

### Phase 9: 성능 최적화 & 테스트 (Week 14)
1. **성능 최적화**
   - 대용량 저장소 처리
   - 가상 스크롤링
   - 백그라운드 작업

2. **테스트**
   - 유닛 테스트
   - 통합 테스트
   - E2E 테스트

### Phase 10: 배포 준비 (Week 15-16)
1. **빌드 & 패키징**
   - Windows / macOS / Linux 빌드
   - 인스톨러 생성
   - Auto-update 구현

2. **문서화**
   - 사용자 가이드
   - 개발자 문서
   - README 업데이트

## 핵심 기능 목록

### 필수 기능 (MVP)
- ✅ 로컬 저장소 열기
- ✅ 커밋 히스토리 조회
- ✅ 브랜치 그래프 시각화
- ✅ 파일 변경사항 확인
- ✅ Diff 뷰어
- ✅ Stage/Unstage
- ✅ 커밋 생성
- ✅ 브랜치 생성/삭제/체크아웃
- ✅ Pull/Push

### 추가 기능
- Merge/Rebase
- Cherry-pick
- Stash
- Tag
- Submodules
- Git LFS 지원
- Conflict 해결 도구
- Interactive Rebase

## 기술적 고려사항

### Go 백엔드
1. **Git 라이브러리 선택**
   - `go-git`: Pure Go 구현 (추천)
   - `git2go`: libgit2 바인딩
   - `os/exec`: Git CLI 직접 실행

2. **Wails 바인딩**
   - Go 함수를 JavaScript에서 호출
   - 이벤트 시스템 활용
   - 에러 핸들링

### Svelte 프론트엔드
1. **상태 관리**
   - Svelte Stores 사용
   - Reactive statements
   - 컴포넌트 간 데이터 공유

2. **성능**
   - Virtual scrolling (대량 커밋)
   - Lazy loading
   - Debouncing/Throttling

3. **UI/UX**
   - 반응형 레이아웃
   - 키보드 단축키
   - 컨텍스트 메뉴
   - 드래그 앤 드롭

## 참고할 오픈소스
- **Fork**: UI/UX 참고
- **GitKraken**: 기능 참고
- **Sourcetree**: 워크플로우 참고
- **lazygit**: TUI 인터페이스 아이디어

## 다음 단계
1. Wails 프로젝트 초기화
2. 기본 프로젝트 구조 생성
3. 첫 번째 기능 구현 (저장소 열기)
