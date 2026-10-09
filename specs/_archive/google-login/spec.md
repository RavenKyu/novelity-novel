# Google 로그인 Spec

> 작성일: 2026-10-09 / 상태: 승인됨 (2026-10-09 대화에서 사용자가 범위 확정)
> ⚠️ 승인 후에는 사용자 지시 없이 수정 금지

## 목표

admin 사용자가 Google 계정으로 로그인해 세션을 유지하고, 로그인하지 않은 사용자는 admin 화면에 들어올 수 없게 한다. 사용자·세션 저장소로 MongoDB를 도입한다.

## 요구사항

- R1. Given compose 실행, When `docker compose up`, Then `nn-mongodb`가 함께 뜨고 core-api가 접속한다. `/api/healthz`는 Mongo ping 실패 시 503을 반환한다.
- R2. Given 로그인 안 된 브라우저, When admin 페이지에 접근, Then `/login`으로 이동한다.
- R3. Given `/login`, When "Google로 로그인"을 누르고 Google에서 동의, Then 사용자가 `users`에 생성(최초) 또는 갱신되고 세션 쿠키가 발급되며 `/`로 돌아온다.
- R4. Given 어떤 Google 계정이든(이메일 검증됨), When 로그인, Then 가입이 허용된다 (허용 목록 없음).
- R5. Given 유효한 세션, When `GET /api/auth/me`, Then 사용자(이름·이메일·사진)를 반환한다. 세션이 없거나 만료되면 401을 반환한다.
- R6. Given 로그인 상태, When 로그아웃, Then 서버 세션이 삭제되고 쿠키가 만료되며 `/login`으로 이동한다.
- R7. Given 위조·재사용된 OAuth state 또는 검증 실패한 id_token, When 콜백, Then 세션을 만들지 않고 `/login?error=…`로 보낸다.
- R8. Given 보호 대상 API, When 세션 없이 호출, Then 401을 반환한다 (인증 미들웨어).

## 비목표 (Non-Goals)

- 역할·권한(RBAC), 사용자 승인/차단 UI — 요청 범위 밖이며, 누구나 가입 가능으로 확정.
- Google 외 로그인 수단(이메일/비밀번호, 다른 IdP).
- 운영 배포용 TLS 종단·도메인 설정 — 로컬 compose 기준. 쿠키 `Secure`는 설정으로 켤 수 있게만 한다.
- Mongo 복제셋·백업.

## 제약

- 보안: 서버 측 Authorization Code + PKCE + state를 쓴다. 세션 ID는 32바이트 난수이고, 쿠키는 `HttpOnly; SameSite=Lax; Path=/`로 둔다. Client Secret은 `.env`로만 주입하고 커밋하지 않는다.
- 호환성: 기존 `/api/healthz` 경로를 유지한다. web은 같은 출처(nginx 프록시)를 유지한다.
- 스타일: 로그인 화면은 docs/STYLE.md를 따른다.

## 완료 기준 (Definition of Done)

- [x] R1~R8 검증. R3·R6은 실제 Google 계정 E2E가 필요하므로 사용자 Client ID 발급 후 수행하고, 그 전에는 가짜 OIDC 제공자로 통합 테스트한다.
- [x] go vet/test, web check/lint/build 통과
- [x] ARCHITECTURE·CONVENTIONS·.env.example 갱신, 설정 안내 제공
