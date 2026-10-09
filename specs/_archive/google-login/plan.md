# Google 로그인 Plan

> 작성일: 2026-10-09 / 상태: 완료
> 근거 문서: [spec.md](./spec.md)

## 아키텍처 영향

| 항목 | 내용 |
|------|------|
| 관련 모듈/레이어 | core-api: config, server, handler, 신규 `internal/auth`, `internal/store`. web: 루트 레이아웃, 신규 `/login`, Sidebar |
| 새 외부 의존성 | `go.mongodb.org/mongo-driver/v2` v2.9.1, `github.com/coreos/go-oidc/v3` v3.21.0(id_token 검증), `golang.org/x/oauth2` v0.37.0. 컨테이너 `mongo:8.0`. 외부 서비스: Google OAuth |
| 모듈 경계/공개 API 변경 | 신규 `/api/auth/{google/login,google/callback,me,logout}`. `/api/healthz`가 DB 상태 반영 |
| 데이터 스키마 변경 | 신규 DB `novelity`: `users`(unique `googleSub`, `email`), `sessions`(`_id`=세션ID 해시, TTL `expiresAt`) |

## 접근 방식

core-api가 OIDC 클라이언트 역할을 한다. 로그인 시작 시 state·PKCE verifier·nonce를 짧은 수명의 HttpOnly 쿠키에 넣고 Google로 리다이렉트한다. 콜백에서 이 값들을 대조하고, 코드를 교환한 뒤 id_token을 검증하고, `email_verified`를 확인한다. 그 다음 user를 upsert하고 세션을 만든다. 세션 토큰은 쿠키에 원문으로 두고, DB에는 SHA-256 해시만 저장한다(DB가 유출돼도 세션을 탈취할 수 없음).

기각한 대안:
- 프론트 GIS 버튼으로 id_token을 POST하는 방식: SPA에서 서드파티 스크립트를 써야 하고 서버 측 흐름보다 이점이 없다.
- JWT 무상태 세션: 로그아웃·강제 만료가 어렵다.

## 단계 (Phases)

- [x] **Phase 1: MongoDB** → 검증: compose up 후 `/api/healthz` 200, mongo 중지 시 503
- [x] **Phase 2: 인증 API** → 검증: go test — 가짜 OIDC 서버로 login→callback→me→logout 통합 테스트, state 위조 거부
- [x] **Phase 3: web** → 검증: check/lint/build, 미로그인 시 `/login` 리다이렉트 화면 확인
- [x] **Phase 4: 문서·E2E** → 검증: 문서 갱신, 사용자 Client ID로 실제 Google 로그인

## 리스크와 대응

- Google Client 미발급 시 E2E 불가 → 가짜 OIDC 통합 테스트로 대체하고, 실제 E2E는 미검증으로 표기
- 리다이렉트 URI 불일치(포트 변경) → `PUBLIC_URL`을 설정 하나로 두고 안내
- Mongo 데이터 손실 → 이름 있는 볼륨 사용. `docker compose down -v`는 데이터를 삭제한다고 안내. 롤백은 compose에서 서비스를 제거하고 의존성을 되돌리면 된다(기존 데이터 없음)

## 실행 근거와 승인 상태

- 사용자 요청 범위·실행 근거: "google login 을 되게하자. mongodb를 docker에 추가해주고 필요한 내용을 내게 알려줘."(2026-10-09)
- 승인 필요 여부·이유: DB·외부 서비스 도입(§1.2). 사용자가 직접 요청했고, 설계안·작업 계획을 제시한 뒤 질문으로 정책을 확정해 실행 근거로 본다
- 승인받은 사항: 누구나 가입, 마일스톤 v0.1.0에 포함, Mongo 포트 127.0.0.1:27017만 노출
- 사용자 지정 검토 지점: 없음
- 보류 범위·재개 조건: 없음 (실제 Google 로그인 2026-10-10 사용자 확인)
