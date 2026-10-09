# 001. Google OIDC 로그인과 MongoDB 서버 세션

- 날짜: 2026-10-09
- 상태: 승인됨
- 관련 기능: specs/_archive/google-login/

## Context

admin에 로그인이 필요하고, 수단은 Google 계정으로 정해졌다. web은 nginx가 같은 출처에서 서빙하는 SPA이고, `/api`는 core-api로 프록시된다. 사용자와 세션을 저장할 DB로 MongoDB를 도입한다.

## Decision

- core-api가 OIDC 클라이언트가 되어 서버 측 Authorization Code 흐름을 수행한다. state, nonce, PKCE(S256)를 함께 쓴다. id_token은 go-oidc로 검증하고, `email_verified`가 참인 계정만 받는다.
- 로그인 흐름 값(state, nonce, verifier)은 10분짜리 HttpOnly 쿠키(`nn_oauth`, Path `/api/auth/google`)에 담는다. 이 쿠키는 한 번 쓰면 삭제한다.
- 세션은 32바이트 난수 토큰이다. 브라우저에는 `nn_session` 쿠키(HttpOnly, SameSite=Lax, Path=/)로 원문을 주고, MongoDB `sessions`에는 SHA-256 해시만 `_id`로 저장한다. 만료는 TTL 인덱스와 조회 조건 두 곳에서 모두 확인한다.
- 사용자는 Google `sub` 기준으로 upsert한다(`users.googleSub`은 unique). 가입 제한은 두지 않는다.

## Alternatives

- 프론트에서 Google Identity Services로 id_token을 받아 서버에 POST — 기각 이유: SPA에 서드파티 스크립트가 들어오고, 서버 측 흐름에 비해 얻는 것이 없다.
- 서명된 JWT로 세션을 무상태 관리 — 기각 이유: 로그아웃이나 강제 만료를 하려면 결국 차단 목록 저장소가 필요하다.
- 세션 저장소로 Redis — 기각 이유: MongoDB를 이미 도입하므로 저장소를 하나 더 늘리지 않는다.

## Consequences

- 좋아지는 것: 토큰이 JS에 노출되지 않는다. 세션을 서버에서 즉시 폐기할 수 있다. DB가 유출돼도 세션을 탈취할 수 없다.
- 감수하는 것: 인증이 필요한 요청마다 Mongo 조회가 2회 일어난다(세션 → 사용자). 같은 출처를 전제로 하므로, API를 다른 도메인으로 분리하면 쿠키·CSRF 정책을 다시 설계해야 한다.
- 되돌리려면: `internal/auth.Repo` 인터페이스 뒤의 구현만 바꾸면 된다. 기존 세션은 버려도 된다(재로그인).
- **재검토 조건**: 외부 공개 배포 전(가입 제한·역할 필요), API가 다른 출처로 분리될 때, 세션 조회가 지연 병목이 될 때.
