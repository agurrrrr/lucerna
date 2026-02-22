# 앱 스토어 배포 가이드

이 가이드는 macOS App Store와 Microsoft Store에 Lucerna를 배포하는 방법을 설명합니다.

## macOS App Store

### 필수 요구사항

1. **Apple Developer Program 멤버십** (연 $99)
   - https://developer.apple.com/programs/ 에서 가입
   - App Store 배포에 필수

2. **Xcode가 설치된 Mac**
   - Mac App Store에서 최신 Xcode 설치
   - Command Line Tools 설치 필요

3. **인증서 및 프로비저닝**
   - Mac App Distribution Certificate
   - Developer ID Application Certificate
   - App Store Provisioning Profile

### App Store 요구사항

#### 1. App Sandbox
`build/darwin/Info.plist` 엔타이틀먼트 파일 추가:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.files.bookmarks.app-scope</key>
    <true/>
</dict>
</plist>
```

#### 2. wails.json 업데이트

```json
{
  "mac": {
    "icon": "build/appicon.png",
    "universalBuild": true,
    "dmg": {
      "title": "Lucerna Installer"
    }
  },
  "info": {
    "productVersion": "1.0.0",
    "copyright": "© 2024 Lucerna. All rights reserved.",
    "comments": "A modern Git GUI client"
  }
}
```

#### 3. 코드 서명

```bash
# 서명과 함께 빌드
wails build -platform darwin -clean

# 앱 서명
codesign --deep --force --verify --verbose \
  --sign "3rd Party Mac Developer Application: YOUR_NAME" \
  --options runtime \
  --entitlements build/darwin/entitlements.plist \
  build/bin/Lucerna.app

# 서명 확인
codesign --verify --deep --strict --verbose=2 build/bin/Lucerna.app
spctl --assess --verbose=4 --type execute build/bin/Lucerna.app
```

#### 4. 앱 아카이브 생성

```bash
# pkg 인스톨러 생성
productbuild --component build/bin/Lucerna.app /Applications \
  --sign "3rd Party Mac Developer Installer: YOUR_NAME" \
  Lucerna.pkg
```

#### 5. App Store Connect에 업로드

```bash
# 검증
xcrun altool --validate-app -f Lucerna.pkg \
  --type macos \
  --username YOUR_APPLE_ID \
  --password YOUR_APP_SPECIFIC_PASSWORD

# 업로드
xcrun altool --upload-app -f Lucerna.pkg \
  --type macos \
  --username YOUR_APPLE_ID \
  --password YOUR_APP_SPECIFIC_PASSWORD
```

#### 6. App Store Connect 설정

1. App Store Connect에서 새 앱 생성
2. 앱 정보 입력:
   - 앱 이름: Lucerna
   - 카테고리: Developer Tools
   - SKU: com.yourcompany.lucerna
   - Bundle ID: 앱과 일치해야 함

3. 스크린샷 제공 (필수 크기):
   - 1280x800 (13.3" 디스플레이)
   - 1440x900 (15.4" 디스플레이)
   - 2880x1800 (Retina)

4. 앱 설명:
   ```
   Lucerna는 깔끔하고 직관적인 인터페이스를 원하는 개발자를 위해 만들어진
   현대적인 Git GUI 클라이언트입니다.

   기능:
   • 시각적 커밋 히스토리 및 브랜치 그래프
   • 통합 diff 뷰어가 있는 파일 변경사항
   • 쉬운 staging 및 커밋
   • 브랜치 관리 (생성, 머지, 삭제)
   • 원격 작업 (push, pull, fetch)
   • Stash 및 tag 지원
   • 파일 히스토리 및 blame 뷰
   • 다크 테마 인터페이스
   ```

5. 개인정보 처리방침 URL (필수)
6. 지원 URL (필수)
7. 심사 제출

### 공증 (App Store 외부 배포용)

App Store 외부로 배포하는 경우:

```bash
# DMG 생성
create-dmg \
  --volname "Lucerna" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --app-drop-link 600 185 \
  "Lucerna.dmg" \
  "build/bin/Lucerna.app"

# 공증
xcrun notarytool submit Lucerna.dmg \
  --apple-id YOUR_APPLE_ID \
  --password YOUR_APP_SPECIFIC_PASSWORD \
  --team-id YOUR_TEAM_ID \
  --wait

# 공증 스테이플
xcrun stapler staple Lucerna.dmg
```

---

## Microsoft Store

### 필수 요구사항

1. **Microsoft Partner Center 계정**
   - 개인: $19 일회성 비용
   - 회사: $99 일회성 비용
   - https://partner.microsoft.com/dashboard 에서 가입

2. **Windows 개발 환경**
   - Windows 10/11
   - Visual Studio 또는 Windows SDK

### Store 요구사항

#### 1. MSIX 패키지 생성

`build/windows/lucerna.appxmanifest` 생성:

```xml
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10"
         xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10"
         xmlns:rescap="http://schemas.microsoft.com/appx/manifest/foundation/windows10/restrictedcapabilities">
  <Identity Name="YourPublisher.Lucerna"
            Publisher="CN=YourPublisher"
            Version="1.0.0.0" />

  <Properties>
    <DisplayName>Lucerna</DisplayName>
    <PublisherDisplayName>Your Company</PublisherDisplayName>
    <Logo>Assets\StoreLogo.png</Logo>
  </Properties>

  <Dependencies>
    <TargetDeviceFamily Name="Windows.Desktop" MinVersion="10.0.17763.0" MaxVersionTested="10.0.22000.0" />
  </Dependencies>

  <Resources>
    <Resource Language="en-us" />
  </Resources>

  <Applications>
    <Application Id="Lucerna" Executable="lucerna.exe" EntryPoint="Windows.FullTrustApplication">
      <uap:VisualElements DisplayName="Lucerna"
                          Description="A modern Git GUI client"
                          BackgroundColor="transparent"
                          Square150x150Logo="Assets\Square150x150Logo.png"
                          Square44x44Logo="Assets\Square44x44Logo.png">
        <uap:DefaultTile Wide310x150Logo="Assets\Wide310x150Logo.png" />
      </uap:VisualElements>
    </Application>
  </Applications>

  <Capabilities>
    <rescap:Capability Name="runFullTrust" />
  </Capabilities>
</Package>
```

#### 2. 필수 에셋

`build/windows/Assets/`에 다음 로고 크기 생성:
- Square44x44Logo.png (44x44)
- Square150x150Logo.png (150x150)
- Wide310x150Logo.png (310x150)
- StoreLogo.png (50x50)
- SplashScreen.png (620x300)

#### 3. MSIX 패키지 빌드

```powershell
# MakeAppx 도구 사용
MakeAppx.exe pack /d "build\bin" /p "Lucerna.msix" /nv

# 패키지 서명
SignTool.exe sign /fd SHA256 /a /f YourCertificate.pfx /p YourPassword Lucerna.msix
```

또는 Visual Studio 사용:
1. Windows Application Packaging Project 생성
2. Lucerna를 참조로 추가
3. 패키지 매니페스트 구성
4. MSIX 빌드

#### 4. Partner Center 제출

1. **앱 제출 생성**
   - Partner Center Dashboard로 이동
   - 앱 및 게임 → 새 제품 → MSIX 또는 PWA 앱
   - 앱 이름 예약: "Lucerna"

2. **가격 및 가용성**
   - 무료 또는 유료
   - 시장 (국가)
   - 출시 날짜

3. **속성**
   - 카테고리: 개발자 도구
   - 연령 등급: 3+
   - 시스템 요구사항

4. **앱 목록**
   - 설명:
     ```
     Lucerna - 현대적인 Git GUI 클라이언트

     Lucerna는 Git 저장소를 관리하기 위한 깔끔하고 직관적인 인터페이스를 제공합니다.
     버전 관리를 시각적으로 제어하고 싶은 개발자에게 완벽합니다.

     주요 기능:
     • 브랜치 그래프가 있는 시각적 커밋 히스토리
     • 파일 변경사항을 위한 통합 diff 뷰어
     • 쉬운 staging 및 커밋
     • 완전한 브랜치 관리
     • Push, pull, fetch 작업
     • Stash 관리
     • Tag 지원
     • 파일 히스토리 및 blame 뷰
     • 아름다운 다크 테마
     ```

   - 스크린샷 (필수):
     - 최소 1개 스크린샷
     - 권장: 1366x768 이상
     - 최대 10개 스크린샷

   - Store 로고 (모든 필수 크기)

5. **패키지 업로드**
   - .msix 파일 업로드
   - 패키지 검증 통과 확인

6. **인증 제출**
   - 모든 정보 검토
   - 스토어에 제출
   - 인증 대기 (일반적으로 1-3일)

### Windows Store용 wails.json 업데이트

```json
{
  "windows": {
    "icon": "build/appicon.png",
    "webview2": "Embed",
    "wailsWinDllEmbedding": true
  },
  "info": {
    "productVersion": "1.0.0",
    "productName": "Lucerna",
    "companyName": "Your Company",
    "copyright": "© 2024 Your Company. All rights reserved."
  }
}
```

---

## 공통 요구사항 (두 스토어 모두)

### 1. 개인정보 처리방침

두 스토어 모두 필수. `docs/privacy-policy.md` 생성:

```markdown
# Lucerna 개인정보 처리방침

최종 업데이트: [날짜]

## 데이터 수집
Lucerna는 개인 데이터를 수집, 저장 또는 전송하지 않습니다.
모든 Git 작업은 장치에서 로컬로 수행됩니다.

## 로컬 저장소
- 저장소 경로는 애플리케이션 환경설정에 로컬로 저장됩니다
- 클라우드 동기화나 외부 서버는 사용되지 않습니다
- 모든 데이터는 장치에 유지됩니다

## 타사 서비스
Lucerna는 타사 분석 또는 추적 서비스와 통합되지 않습니다.

## 문의
개인정보 보호 관련 문의: privacy@yourcompany.com
```

호스팅: `https://yourwebsite.com/privacy`

### 2. 지원 웹사이트

다음 내용이 포함된 필수 지원 URL:
- 연락처 정보
- FAQ
- 알려진 문제
- 버그 신고 방법

### 3. 앱 아이콘

고품질 아이콘 생성:
- macOS: .icns 파일 (512x512, 1024x1024)
- Windows: .ico 파일 (여러 크기)
- `iconutil` (macOS) 또는 온라인 변환기 사용

### 4. 스크린샷

모범 사례:
- 주요 기능 표시
- 실제 앱 스크린샷 사용 (목업 금지)
- 기능을 설명하는 캡션 포함
- 적절한 OS 배경에 표시
- 3-5개 스크린샷 권장

### 5. 연령 등급

두 스토어 모두 연령 등급 필요:
- Lucerna: 전체 이용가 (성인 콘텐츠 없음)
- ESRB/PEGI: E for Everyone

### 6. 현지화 (선택사항)

국제 시장용:
- 앱 설명 번역
- 스크린샷 현지화
- 앱 내 다국어 지원

---

## Store 배포를 위한 GitHub Actions

### 자동화된 MSIX 빌드

`.github/workflows/store-release.yml`에 추가:

```yaml
name: Store Release

on:
  push:
    tags:
      - 'store-v*'

jobs:
  windows-store:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build MSIX
        run: |
          # MSIX 빌드 및 패키징
          wails build -platform windows -clean
          # MSIX 서명 및 생성 (인증서 필요)
```

---

## 심사 프로세스

### macOS App Store
- **초기 심사**: 1-3일
- **업데이트**: 1-2일
- **일반적인 거부 사유**:
  - 누락된 엔타이틀먼트
  - Sandbox 위반
  - 실행 시 충돌
  - 누락된 개인정보 처리방침

### Microsoft Store
- **초기 심사**: 1-3일
- **업데이트**: 1-2일
- **일반적인 거부 사유**:
  - 패키지 검증 실패
  - 필수 에셋 누락
  - 앱 충돌
  - 정책 위반

---

## 유지보수

### 업데이트

두 스토어 모두:
1. wails.json에서 버전 업데이트
2. 새 패키지 빌드
3. 스토어에 업로드
4. 심사 제출
5. 릴리스 노트 업데이트

### 리뷰 응답

- 사용자 리뷰 모니터링
- 문제에 신속하게 대응
- 업데이트에서 버그 수정
- 릴리스 노트에 변경사항 전달

---

## 리소스

### macOS
- [App Store 심사 가이드라인](https://developer.apple.com/app-store/review/guidelines/)
- [Xcode 도움말](https://developer.apple.com/documentation/xcode)
- [공증 가이드](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)

### Windows
- [Microsoft Store 정책](https://docs.microsoft.com/en-us/windows/uwp/publish/store-policies)
- [MSIX 패키징](https://docs.microsoft.com/en-us/windows/msix/)
- [Partner Center 가이드](https://docs.microsoft.com/en-us/windows/uwp/publish/)

### Wails
- [Wails 문서](https://wails.io/docs/)
- [프로덕션 빌드](https://wails.io/docs/guides/building/)
