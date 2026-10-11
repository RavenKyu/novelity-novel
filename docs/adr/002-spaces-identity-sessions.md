# 002. 세 공간과 identity·역할·세션 경계

- 날짜: 2026-10-11
- 상태: 승인됨 (Raven이 이 작업 창에서 결정)
- 관련 에픽: nn-l4cg
- 이전 결정: [ADR-001](001-google-oidc-mongo-sessions.md)

## Context

기존 apps/web은 작가 기능의 출발점이다. 독자 서점과 시스템 관리자 공간을 더하면서 ADR-001의 역할 필요·출처 분리 재검토 조건에 해당한다. 앱마다 API를 만들면 identity와 저장소 정책을 중복하므로 웹 세 개와 모듈형 core-api 하나를 둔다.

## Decision

### 공간과 이름

| 공간 | 운영 출처 | 앱 | 렌더링 |
|------|-----------|----|--------|
| 독자 서점 | https://novelity.kr | apps/bookstore-web | SvelteKit SSR, adapter-node |
| 작가 | https://studio.novelity.kr | apps/studio-web | SvelteKit SPA, adapter-static |
| 시스템 관리자 | https://admin.novelity.kr | apps/admin-web | SvelteKit SPA, adapter-static |

디렉터리는 kebab-case로 정한다. 현재 앱을 studio-web으로 옮기고 서점·관리자 골격은 각각 nn-aa6i와 nn-w097에서 만든다. 공통 디자인 토큰·기본 컴포넌트는 packages/ui, API 타입·fetch 래퍼는 packages/api-client에 둔다. Sidebar·Header 등의 화면 틀은 작가와 관리자만 함께 쓰고, 서점은 자체 화면 틀을 만든다. npm workspaces와 루트 lockfile 하나를 사용한다.

### API와 도메인

core-api는 identity, studio, catalog, library, moderation 도메인으로 나누는 모듈형 모놀리스다. HTTP 어댑터가 도메인에 의존하며 도메인은 HTTP·Echo를 모른다. Mongo 연결·설정은 internal/platform에 둔다. studio는 catalog에 발행 스냅샷을 전달하며 catalog는 비공개 작가 원고를 읽지 않는다. 발행·작품 기능 자체는 후속 에픽의 일이다.

/api/studio, /api/store, /api/admin 그룹은 각각 인증·권한 정책을 갖는다. /api/auth와 /api/healthz는 공통이다. 실행 설정 --surfaces 또는 API_SURFACES로 켤 그룹을 고른다. 기본 공개 인스턴스는 studio,store만 켜며 /api/admin은 404다. 관리자 인스턴스는 admin만 켜고 내부망에 따로 배치한다. 실제 VPN·IP 허용 목록과 관리자 배포는 nn-w097에서 정한다.

### 역할

users.roles는 복수 역할 배열이다. reader는 기본 역할이며 기존 roles 없는 사용자도 reader로 읽는다. author는 작가 신청과 필명 프로필 기능이 부여한다. 이번 에픽은 신청·필명 기능이나 자동 author 승격을 만들지 않는다. 기존 작가 셸은 세션만 요구하고, 후속 작가 기능의 권한은 RequireSession 위에 RequireRole(author)를 쌓는다.

admin은 ADMIN_EMAILS 환경변수에 나열한 이메일 허용 목록으로 결정한다. Google이 검증한 이메일로 가입한 사용자만 대상이며 요청마다 현재 허용 목록으로 유효 역할을 계산해 목록에서 제거하면 기존 세션에서도 권한을 잃는다. 저장된 admin 문자열만으로 관리자 권한을 부여하지 않는다. reader·author와 admin은 서로 배타적이지 않다.

### 로그인과 세션

서점과 작가는 SSO로 묶는다. 두 출처가 같은 서버 세션을 쓰며 운영 nn_session 쿠키는 Domain=novelity.kr, HttpOnly, Secure, SameSite=Lax, Path=/다. 이 부모 도메인의 모든 하위 도메인을 신뢰하는 것이 전제다. 부모 도메인 쿠키가 관리자 호스트에도 전송될 수 있으나 관리자 인스턴스는 이를 인증에 사용하지 않는다. 로컬 localhost 개발에서는 Domain을 생략하고 포트별 출처를 명시한다. 공개 세션 기본 TTL은 기존 168시간을 유지한다.

관리자는 별도 Google 로그인과 nn_admin_session 쿠키를 사용한다. Domain을 생략해 admin 호스트에만 전송하며 HttpOnly, Secure, SameSite=Lax, Path=/, 절대 TTL 2시간이다. 관리자 세션은 public과 다른 audience로 저장·조회해 공개 세션 토큰으로 관리자 인증을 할 수 없다. 관리자 로그아웃과 서점·작가 로그아웃은 서로의 세션을 폐기하지 않는다. 관리자가 공개 계정을 쓰더라도 관리자 로그인을 다시 해야 한다.

Google callback URI는 각 공간의 출처 + /api/auth/google/callback이다. 세 공간 URI를 Google OAuth 클라이언트에 등록하며 PUBLIC_URL 하나라는 전제를 없앤다. 요청 Host를 그대로 신뢰하지 않고 구성한 출처에서 OAuth redirect URI를 고른다. state·nonce·PKCE는 유지한다. 로그인 흐름 쿠키는 host-only이고 돌아갈 주소는 흐름과 함께 무결성을 보호한다.

returnTo는 상대 경로 또는 현재 인스턴스가 제공하는 공간의 명시적 허용 출처만 받는다. 공개 로그인에서 admin 출처로 돌아갈 수 없고 관리자 로그인은 admin 출처에만 돌아간다. 스킴 상대 주소, 사용자 정보가 포함된 URL, 제어 문자·역슬래시 등 URL 해석 차이가 생기는 입력은 거절한다. 콜백에서 다시 검증한다. 복귀 주소는 최대 2048바이트, 최종 흐름 쿠키는 최대 4096바이트로 제한한다. 인증 실패 시 해당 공간의 /login으로 돌아가 오류와 유효한 returnTo를 유지한다.

공유 쿠키는 서로 다른 출처에서 쓰이므로 상태 변경 요청은 요청 출처도 확인한다. 브라우저는 각 앱의 /api 프록시로 같은 출처 요청을 한다. 서점 SSR의 쿠키 전달은 nn-aa6i에서 자기 API 출처에 한해 구현한다.

## Alternatives

- 한 도메인의 /studio·/admin: 배포·OAuth 등록은 단순하지만 관리자와 공개 웹의 출처 경계가 약하다.
- 세 공간 모두 별도 로그인: 서점·작가 왕복 때 로그인 상태가 끊긴다.
- 앱마다 API: 공통 identity·세션 정책이 중복되고 데이터 경계가 API 사이의 조율 문제로 바뀐다.
- 중앙 로그인과 일회용 코드 교환: host-only 공개 쿠키를 유지할 수 있으나 별도 인증 프로토콜이 필요하다. 이번에는 신뢰한 부모 도메인 쿠키로 SSO를 구현한다.

## Consequences

웹 배포와 API 권한 경계를 독립적으로 확장한다. 관리자는 라우트 노출, 세션 audience, 역할의 세 경계로 분리한다. 운영에는 세 도메인·TLS·Google callback 등록과 신뢰한 하위 도메인 관리가 필요하다. 내부망 관리자 접근 정책은 관리자 골격 에픽에서 완성한다. 로그인 중 프로세스가 재시작되면 프로세스별 흐름 서명 키가 바뀌어 다시 로그인해야 한다.

재검토 조건은 신뢰하지 않는 하위 도메인 도입, 외부 API 소비자 도입, 다중 API 복제본에서 로그인 흐름 공유 필요, 관리자 MFA·세션 정책 변경, 도메인 간 독립 배포가 모놀리스보다 유리해지는 때다.

Decided in: nn-l4cg
