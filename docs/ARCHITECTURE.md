# Architecture

> 마지막 갱신: 2026-10-11

## 시스템 개요

Novelity 관리자(admin) 서비스다. 컨테이너 세 개를 docker-compose로 묶는다. core-api와 web은 compose `app` 프로필에 속한다. 그래서 개발 중에는 MongoDB만 컨테이너로 띄우고, 두 서비스는 호스트에서 live reload로 실행한다(`scripts/dev`, CONVENTIONS.md의 개발 환경).

- `nn-web-server`: nginx가 SvelteKit SPA 빌드 결과를 서빙하고, `/api/*` 요청은 core-api로 프록시한다.
- `nn-core-api-server`: Go 언어와 Echo v5로 만든 HTTP API다. compose 내부망에만 노출한다.
- `nn-mongodb`: MongoDB 8.0이다. 사용자·세션을 저장한다. 호스트에는 127.0.0.1:27017로만 노출한다.

로그인은 Google OIDC(서버 측 code flow)와 Mongo 세션 쿠키로 한다(ADR-001).

브라우저는 web-server 한 곳(동일 출처)만 호출하므로 CORS 설정이 필요 없다.

## 모듈 구조와 경계

| 구성 | 책임 | 위치 |
|------|------|------|
| core-api 진입점 | 설정 로딩, 시그널 처리, graceful shutdown | apps/core-api/cmd/server |
| core-api server | Echo 생성, 미들웨어, 라우트 등록(`/api` 그룹) | apps/core-api/internal/server |
| core-api handler | 요청 처리(healthz) | apps/core-api/internal/handler |
| core-api auth | Google 로그인·콜백, 세션 쿠키, `RequireSession` 미들웨어, `Repo` 인터페이스 | apps/core-api/internal/auth |
| core-api store | Mongo 연결·인덱스, `auth.Repo` 구현 | apps/core-api/internal/store |
| core-api config | 환경변수 → Config (`.env.example` 참조) | apps/core-api/internal/config |
| web 라우트 | 페이지(각 페이지가 Header·PageContent 렌더) | apps/web/src/routes |
| web 셸 | Sidebar·Header·PageContent (argos 스타일, docs/STYLE.md) | apps/web/src/lib/components/layout |
| web 공용 | 컴포넌트·자산, 디자인 토큰(`routes/layout.css`), 세션 상태(`auth.svelte.ts`) | apps/web/src/lib |
| web 인증 가드 | `/api/auth/me` 확인 후 미로그인 시 `/login`(셸 없는 화면)으로 보냄 | apps/web/src/routes/+layout.svelte |
| web 서빙 | 정적 서빙, SPA fallback, `/api` 프록시 | apps/web/nginx.conf |
| 오케스트레이션 | 서비스 정의·포트, core-api·web은 `app` 프로필 | docker-compose.yml |
| 개발 실행 | MongoDB(compose) + core-api(air) + web(vite dev) 실행, `.env` 로딩 | scripts/dev, apps/core-api/.air.toml |
| 개발 도구 | Go 툴체인 래퍼(로컬 go 또는 Docker, host 네트워크) | scripts/go |
| e2e | 실행 중인 앱(개발 서버 또는 `app` 프로필) 대상 Playwright, Mongo 세션 시드 fixture | tests/e2e, package.json(루트 도구 전용) |
| 검증 자동화 | pre-commit 훅, GitHub Actions(PR 검증·e2e·이미지 빌드) | .githooks, .github/workflows |
| 릴리스 자동화 | 버전 계산·CHANGELOG(라벨 기반), 릴리스 PR 생성, 태그·GitHub Release·develop 동기화 | scripts/release, .github/release.yml, .github/workflows/{prepare,publish}-release.yml, CHANGELOG.md |

- web은 SvelteKit `adapter-static`(fallback `index.html`)과 `ssr = false` 설정으로 클라이언트 전용 SPA로 빌드한다.
- UI 컴포넌트는 flowbite-svelte와 flowbite-svelte-icons를 쓰고, 스타일은 Tailwind v4로 처리한다. 겉모양은 argos에서 차용한 디자인 토큰을 따른다(docs/STYLE.md).

## 의존성 방향

- core-api: `cmd/server` → `internal/server` → `internal/{handler,auth}`. `internal/store` → `internal/auth`(Repo 구현). `auth`는 저장소를 모르고 `Repo` 인터페이스에만 의존한다. `config`는 `cmd`만 사용한다.
- web → core-api: HTTP(`/api/*`)로만 통신한다. 코드 의존은 없다.

## 데이터 저장소와 스키마

MongoDB `novelity` DB. 인덱스는 core-api가 시작할 때 `EnsureIndexes`로 만든다(멱등). 마이그레이션 도구는 아직 없다.

| 컬렉션 | 필드 | 인덱스 |
|--------|------|--------|
| users | `_id`, `googleSub`, `email`, `name`, `picture`, `createdAt`, `lastLoginAt` | `googleSub` unique |
| sessions | `_id`(세션 토큰 SHA-256), `userId`, `createdAt`, `expiresAt` | `expiresAt` TTL(0초) |

## 관련 ADR

- [ADR-001](adr/001-google-oidc-mongo-sessions.md) Google OIDC 로그인과 MongoDB 서버 세션
