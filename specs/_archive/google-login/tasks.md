# Google 로그인 Tasks

> moai epic `nn-jk9v`와 대응.

## 실행 순서 근거 (한 줄)

Mongo가 세션·사용자 저장의 전제 → 인증 API → 이를 소비하는 web → 문서·실검증.

## Phase 1: MongoDB (nn-jk9v.lck)

- [x] T1. compose에 `nn-mongodb`(mongo:8.0, 볼륨, healthcheck, 127.0.0.1:27017) 추가 → 검증: `docker compose ps` healthy
- [x] T2. core-api Mongo 연결 + healthz에 ping 반영 → 검증: healthz 테스트(ping 성공/실패), compose 스모크

## Phase 2: 인증 API (nn-jk9v.z28, nn-jk9v.ovg)

- [x] T3. store: users upsert, sessions create/get/delete + 인덱스 → 검증: 통합 테스트(Mongo 컨테이너)
- [x] T4. auth: Google login/callback(state·PKCE·nonce, id_token 검증) → 검증: 가짜 OIDC 서버 테스트, state 위조 거부
- [x] T5. me/logout + RequireSession 미들웨어 → 검증: 테스트(401/200, 로그아웃 후 401)

## Phase 3: web (nn-jk9v.mze)

- [x] T6. `/login` 페이지(크롬리스), 레이아웃 인증 가드, 사이드바 사용자 칩·로그아웃 → 검증: check/lint/build, 스크린샷

## Phase 4: 문서·검증 (nn-jk9v.ztp, nn-jk9v.4qr)

- [x] T7. ARCHITECTURE·CONVENTIONS·.env.example, 설정 안내 → 검증: 문서 명령 실행
- [x] T8. 실제 Google 로그인 E2E → 검증: 사용자 Client ID 필요
