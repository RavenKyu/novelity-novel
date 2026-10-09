---
기능: google-login
상태: 완료
마지막 갱신: 2026-10-10
---

# Google 로그인 Context

## 현재 상태 (3~5문장)

완료되었다. Phase 1~4가 모두 끝났다. 2026-10-10에 사용자가 실제 Google 계정으로 로그인되는 것을 확인했다. moai epic `nn-jk9v`의 6개 이슈가 모두 닫혔다. 장기적으로 남길 결정은 ADR-001과 ARCHITECTURE(Mongo 스키마)에 반영했다.

## 핵심 결정 로그 (누적, 최신이 위)

- [2026-10-09] 결정: nginx `/api` 프록시를 `resolver 127.0.0.11` + 변수 방식으로 바꿈 / 이유: core-api 컨테이너를 재생성하면 IP가 바뀌는데, nginx가 시작 시 조회한 옛 IP를 계속 써서 502가 났음
- [2026-10-09] 결정: OIDC discovery는 첫 로그인 때 지연 실행(실패 시 다음에 재시도) / 이유: Google에 닿지 않아도 서버는 떠야 함
- [2026-10-09] 결정: Google 미설정 시 서버는 정상 기동하고 `/login?error=not_configured`로 안내 / 이유: 설정 전에도 개발 가능
- [2026-10-09] 결정: 누구나 가입(허용 목록 없음) / 이유: 사용자 확정 / 재검토 조건: 외부 공개 배포 전 (ADR-001)
- [2026-10-09] 결정: 서버 측 code flow + PKCE, Mongo 세션(토큰 해시 저장, TTL) / 기각: GIS id_token POST, JWT 무상태 (ADR-001)
- [2026-10-09] 결정: Mongo는 127.0.0.1:27017만 호스트 노출 / 이유: 사용자 확정(로컬 도구 접근)

## 시도했으나 실패한 접근 ⚠️

해당 없음

## 발견된 문제 / 열린 질문

- RequestLogger가 콜백 URI(일회용 `code`, `state`)를 로그에 남긴다. code는 PKCE로 보호되고 한 번만 쓸 수 있어 위험은 낮다. 운영 전 로그 마스킹을 검토한다.
- Mongo root 계정을 앱이 그대로 사용한다. 운영 전에 앱 전용 최소 권한 계정으로 분리를 검토한다.

## 다음 세션 시작점

해당 없음(완료). 운영 배포 전에는 "발견된 문제 / 열린 질문"의 두 항목(로그 마스킹, Mongo 앱 전용 계정)을 검토한다.

## 파일 맵

- `apps/core-api/internal/auth/auth.go` — 로그인·콜백·me·logout, RequireSession, Repo 인터페이스
- `apps/core-api/internal/auth/auth_test.go` — 가짜 OIDC 제공자 + 메모리 Repo로 흐름 전체 테스트
- `apps/core-api/internal/store/{store,auth}.go` — Mongo 연결·인덱스·Repo 구현
- `apps/core-api/internal/store/auth_test.go` — MONGO_TEST_URI 통합 테스트
- `apps/core-api/internal/handler/health.go` — healthz(DB ping 반영)
- `apps/web/src/lib/auth.svelte.ts` — 세션 상태, loadSession/logout
- `apps/web/src/routes/+layout.svelte` — 인증 가드, 셸 없는 화면 분기
- `apps/web/src/routes/login/+page.svelte` — 로그인 화면, 오류 코드 → 메시지
- `apps/web/src/lib/components/layout/UserMenu.svelte` — 사용자 칩·로그아웃
- `apps/web/nginx.conf` — 동적 DNS 재조회 프록시
- `docker-compose.yml`, `.env.example` — nn-mongodb, 인증 환경변수

## 검증·승인 상태

- 실행한 검증과 결과:
  - go vet/fmt/test 통과(auth 7, store 2, server 2)
  - web check/lint, Docker 빌드 통과
  - compose: healthz 200, Mongo 중지 시 503
  - 가짜 Client ID로 실제 Google discovery를 거친 authorize 리다이렉트(PKCE·nonce·redirect_uri) 확인
  - 시드 세션으로 Playwright 확인: 미로그인 → `/login`, 로그인 상태에서 `/login` → `/`, 사용자 메뉴, 로그아웃 후 401
- 실제 Google 계정 로그인: 2026-10-10 사용자 확인(2026-10-09 13:49 세션 생성 기록과 일치)
- 이후 e2e 자동화: tests/e2e/auth-guard·session spec(Playwright, Mongo 세션 시드)
- 미실행 검증과 이유: 없음
- 승인된 범위·근거: plan.md 참조
- 남은 승인 대상·재개 조건: 없음
