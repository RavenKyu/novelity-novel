# Architecture

> 마지막 갱신: 2026-10-11

## 시스템 개요

Novelity는 작가 공간, 독자 서점, 시스템 관리자 공간으로 나뉜다. 웹은 독립 앱 세 개를 지향하며 API는 도메인별 모듈을 가진 core-api 하나다. 현재 작가 앱과 공통 기반이 구현되었고 서점·관리자 앱은 nn-aa6i·nn-w097에서 만든다. 공간·도메인·세션 결정은 [ADR-002](adr/002-spaces-identity-sessions.md)에 있다.

| 공간 | 운영 출처 | 웹 | 렌더링 | API |
|------|-----------|----|--------|-----|
| 작가 | https://studio.novelity.kr | apps/studio-web | SPA, adapter-static | /api/studio |
| 독자 서점 | https://novelity.kr | apps/bookstore-web (후속 골격) | SSR, adapter-node | /api/store |
| 관리자 | https://admin.novelity.kr | apps/admin-web (후속 골격) | SPA, adapter-static | /api/admin |

개발 중에는 MongoDB만 컨테이너로 띄우고 core-api와 작가 웹은 호스트에서 live reload로 실행한다. compose `app` 프로필은 공개 인스턴스(core-api + studio-web)를 띄운다.

- `nn-studio-web-server`: nginx가 SPA를 서빙하고 /api 요청을 core-api로 프록시한다. /api/admin 요청은 404로 막는다.
- `nn-core-api-server`: Go·Echo v5 API. compose 내부망에만 노출하며 studio,store만 등록한다. admin 라우트가 없는 것이 기본값이다.
- `nn-mongodb`: MongoDB 8.0. 사용자·세션을 저장한다. 호스트에는 127.0.0.1:27017로만 노출한다.

관리자는 동일 바이너리를 `--surfaces=admin`으로 별도 실행하고 내부망에 배치한다. 공개·관리자 그룹을 함께 선택하면 기동을 거절한다. 관리자 웹·compose 인스턴스와 실제 접근 제한은 nn-w097에서 구현한다. /api/auth와 /api/healthz는 어느 인스턴스에서도 공통으로 등록한다.

브라우저는 자기 웹의 /api 프록시로 같은 출처 요청을 한다. 서점 SSR의 신뢰한 API 대상에 대한 쿠키 전달은 nn-aa6i에서 구현한다.

## 모듈 구조와 경계

| 구성 | 책임 | 위치 |
|------|------|------|
| API 진입점 | 설정·surfaces 플래그 검증, 의존성 조립, graceful shutdown | apps/core-api/cmd/server |
| HTTP 조립 | Echo와 미들웨어, 선택한 그룹 등록 | apps/core-api/internal/server |
| identity | User·Profile·Session·Repo 계약, 유효 역할 계산 | apps/core-api/internal/identity |
| auth HTTP | Google code flow, 출처·returnTo 검증, 세션 쿠키·권한 미들웨어 | apps/core-api/internal/http/auth |
| 공간별 HTTP | studio: 세션 필수, store: 공개, admin: 관리자 세션·역할 필수 | apps/core-api/internal/http/{studio,store,admin} |
| 플랫폼 | 환경변수·선택 검증, Mongo 연결·인덱스·identity.Repo 구현 | apps/core-api/internal/platform/{config,mongo} |
| 공통 health | Mongo ping, 정상 200·불통 503 | apps/core-api/internal/handler |
| 작가 라우트 | 페이지, /login과 인증 가드, 앱별 세션 상태 | apps/studio-web/src/routes, src/lib/auth.svelte.ts |
| 공통 UI | 디자인 토큰, Header·PageContent·Sidebar·SidebarItem·UserMenu | packages/ui |
| API 클라이언트 | User·Role 타입, me·logout·health fetch 래퍼, 로그인 링크 | packages/api-client |
| 작가 서빙 | SPA fallback, /api 프록시 | apps/studio-web/nginx.conf |
| 오케스트레이션 | 공개 인스턴스와 Mongo, opt-in app 프로필 | docker-compose.yml |
| 개발 실행 | Mongo(compose), API(air), 작가 웹(vite) | scripts/dev, apps/core-api/.air.toml |
| 검증 | 경로별 pre-commit, PR 검증·e2e·이미지 빌드 | .githooks, .github/workflows, tests/e2e |
| 릴리스 | 단일 버전·CHANGELOG·릴리스 PR·태그·develop 동기화 | scripts/release, .github/workflows/{prepare,publish}-release.yml |

각 공간에는 연결 확인용 GET /api/<surface>/healthz가 있다. 작품·카탈로그·관리 기능의 실제 엔드포인트는 후속 기능에서 추가한다. store의 선택적 세션과 catalog 패키지는 서점 골격 에픽의 범위다.

공통 UI는 앱의 `$app`·`$lib`나 세션 상태를 참조하지 않는다. Sidebar는 제목·버전·경로·메뉴·사용자 메뉴 snippet을, UserMenu는 사용자·로그아웃 함수를 받는다. 서점은 토큰과 기본 컴포넌트를 쓰되 Sidebar·Header 화면 틀은 사용하지 않는다. UI는 flowbite-svelte·아이콘과 Tailwind v4를 사용한다([STYLE](STYLE.md)).

## 의존성 방향

- `cmd/server` → `server` → `http/{auth,studio,store,admin}` → `identity`.
- `platform/mongo` → `identity.Repo`. identity는 HTTP·Echo·Mongo 구현을 모른다. config는 진입점에서 읽는다.
- 향후 `studio` → `catalog`: 발행 시 공개 스냅샷을 넘긴다. catalog는 작가의 편집 중 원고를 직접 참조하지 않는다. library·moderation도 도메인 계층에 추가한다.
- 웹 → `packages/{ui,api-client}`. ui는 api-client의 User 타입만 참조하고 API 요청은 앱이 소유한다. 웹 → core-api 통신은 HTTP만 사용한다.

## 인증·역할·세션

Google OIDC 서버 측 code flow는 state·nonce·PKCE를 유지하고 검증된 이메일만 받는다([ADR-001](adr/001-google-oidc-mongo-sessions.md)). OAuth callback URI는 요청 Host가 구성된 공간 출처와 일치할 때 그 출처에서 결정한다. STUDIO_URL·BOOKSTORE_URL·ADMIN_URL 각각의 callback을 Google에 등록한다.

서점·작가는 `nn_session`을 공유한다. 운영 `SESSION_DOMAIN=novelity.kr`로 부모 도메인 쿠키를 발급한다. 관리자 `nn_admin_session`은 host-only, 절대 TTL 2시간이다. 두 쿠키 모두 HttpOnly·SameSite=Lax·Path=/이고 HTTPS 구성에서는 Secure다. 관리자 인스턴스는 공개 쿠키·public audience 토큰을 받지 않는다. 로그아웃은 자기 audience만 지운다.

로그인 흐름은 host-only `nn_oauth`(관리자는 `nn_admin_oauth`)에 담고 프로세스별 HMAC 키로 무결성을 보호한다. 만료는 10분이며 returnTo·시작 출처·state·nonce·PKCE verifier가 함께 묶인다. 복귀 주소는 최대 2048바이트이고 최종 흐름 쿠키도 4096바이트 이하만 허용한다. 시작 출처가 다른 콜백이나 변조·만료된 쿠키는 거절한다. 재시작 또는 복제본 변경 중인 로그인은 다시 시작해야 한다. 이미 로그인된 세션의 공간 복귀도 로그인 시작 API에서 같은 출처·audience·역할 검증을 거쳐 바로 이동한다. returnTo는 명시적 허용 출처 또는 안전한 상대 경로만 허용하고 callback에서도 재검증한다. 공개 로그인은 관리자 출처로 복귀할 수 없다.

역할은 reader(기본)·author(작가 기능이 부여)·admin(ADMIN_EMAILS의 현재 이메일 허용 목록)이다. 복수 역할을 허용한다. roles 없는 기존 사용자는 reader로 읽으며 재로그인이 author를 지우지 않는다. 저장된 admin 역할만으로 권한을 얻을 수 없다. 현재 작가 셸은 세션만 요구한다. author 신청·필명과 author 필수 기능은 후속 작가 에픽의 범위다. `RequireRole`은 `RequireSession` 다음에 적용한다.

쿠키 인증의 상태 변경 요청에는 `SameOrigin`을 적용한다. Origin 또는 Referer가 다른 출처이면 403이며 Fetch Metadata가 cross-site·same-site인 출처 없는 요청도 거절한다. 출처 헤더가 없는 비브라우저 요청은 허용한다.

## 저장소

MongoDB `novelity` DB. 시작 시 EnsureIndexes로 인덱스를 멱등 생성한다.

| 컬렉션 | 필드 | 인덱스 |
|--------|------|--------|
| users | _id, googleSub, email, name, picture, roles, createdAt, lastLoginAt | googleSub unique |
| sessions | _id(토큰 SHA-256), userId, audience(public/admin), createdAt, expiresAt | expiresAt TTL(0초) |

기존 audience 없는 세션은 public으로만 인정한다. 토큰 원문은 브라우저 쿠키에만 둔다. TTL 삭제 지연과 무관하게 조회에서 expiresAt을 검사한다. ADMIN_EMAILS 변경은 프로세스 재시작 후 기존 세션에도 즉시 적용한다.

## 관련 ADR

- [ADR-001](adr/001-google-oidc-mongo-sessions.md) Google OIDC 로그인과 MongoDB 서버 세션
- [ADR-002](adr/002-spaces-identity-sessions.md) 세 공간과 identity·역할·세션 경계

Decided in: nn-l4cg
