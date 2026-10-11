# Conventions

> 마지막 갱신: 2026-10-11

## 명령어

core-api 명령은 `scripts/go`로 실행한다. 로컬에 `go`가 있으면 그것을 쓰고, 없으면 `golang:1.27-alpine` 컨테이너로 실행한다.

| 목적 | 실제 명령어 |
|------|-------------|
| 전체 테스트 | `cd apps/core-api && MONGO_TEST_URI='mongodb://novelity:novelity-dev@localhost:27017/?authSource=admin' ../../scripts/go test ./...` (compose의 Mongo 필요, 미설정 시 Mongo 테스트 skip) |
| 관련 테스트 | `cd apps/core-api && ../../scripts/go test ./internal/<pkg>/...` |
| 린트·포맷 | `cd apps/core-api && ../../scripts/go vet ./... && ../../scripts/go fmt ./...` · `npm run lint` (수정: `npm run format --workspace @novelity/studio-web`, 공통 패키지는 `npm exec -- prettier --write packages`) |
| 타입 검사 | `npm run check` |
| 빌드 | `npm run build` · `docker compose --profile app build` |
| 개발 실행 | 최초 1회 `brew install go`, `cp .env.example .env` 후 값 입력, 루트에서 `npm ci` → `npm run dev` → http://localhost:8000 (`WEB_PORT`로 변경). 서점은 http://localhost:8001 (`BOOKSTORE_PORT`로 변경), 관리자는 http://localhost:8002 (`ADMIN_PORT`로 변경). 따로 띄우려면 `npm run dev:api` / `npm run dev:studio` / `npm run dev:bookstore` / `npm run dev:admin-api` / `npm run dev:admin` |
| 컨테이너 실행 | `docker compose --profile app up -d --build` (공개 운영 유사 구성. 관리자도 띄우려면 `--profile admin`을 함께 지정한다. 개발 서버와 8000·8001 포트를 같이 쓰므로 동시에 띄우지 않는다) |
| e2e (Playwright) | 최초 1회 `npm ci && npm run test:e2e:install` → 앱 실행(`npm run dev`) 후 `npm run test:e2e` (관련 spec만: `npx playwright test --config tests/e2e/playwright.config.ts tests/e2e/<name>.spec.ts`) |
| 워크플로 린트 | `docker run --rm -v "$PWD":/repo -w /repo rhysd/actionlint:1.7.7` |
| 스모크 | `curl -s localhost:8000/api/healthz` → `{"status":"ok"}` (Mongo 불통 시 503) |
| Mongo 셸 | `docker compose exec nn-mongodb mongosh -u novelity -p novelity-dev` |

## 개발 환경

개발 중에는 MongoDB만 컨테이너로 띄우고 공개·관리자 core-api와 세 웹은 호스트에서 실행한다. 저장하면 바로 반영된다. 진입점은 `scripts/dev`(`npm run dev`)다.

| 구성 | 실행 방식 | 변경 반영 |
|------|-----------|-----------|
| nn-mongodb | `docker compose up -d --wait nn-mongodb` (`dev:api`가 자동 실행) | — |
| core-api | air(`apps/core-api/.air.toml`)가 `:8080`으로 실행 | `*.go` 저장 → 재빌드·재시작. 빌드가 실패하면 마지막으로 성공한 빌드가 계속 돈다 |
| admin core-api | 별도 air(`apps/core-api/.air-admin.toml`), loopback `:${ADMIN_API_PORT:-8081}`, admin 그룹만 제공 | public/admin의 빌드 출력은 tmp/public·tmp/admin으로 분리한다 |
| admin-web | vite가 loopback `:${ADMIN_PORT:-8002}`로 실행, `/api`는 ADMIN_CORE_API_URL로 프록시 | HMR |
| studio-web | vite dev 서버가 `:${WEB_PORT:-8000}`으로 실행, `/api`는 `:8080`으로 프록시 | HMR(새로고침 없이 반영) |
| bookstore-web | vite SSR 서버가 `:${BOOKSTORE_PORT:-8001}`으로 실행, `/api`와 SSR 요청은 CORE_API_URL로 직접 연결 | HMR·서버 모듈 reload |

- `scripts/dev`는 루트 `.env`를 읽고 `MONGO_ROOT_*`·`MONGO_PORT`로 `MONGO_URI`(localhost)를 만든다. `.env`에 `MONGO_URI`를 직접 적으면 그 값을 쓴다.
- 작가 개발 서버와 컨테이너의 기본 출처는 `STUDIO_URL=http://localhost:8000`이다. `WEB_PORT`를 바꾸면 STUDIO_URL도 함께 바꾸고 Google callback URI를 다시 등록한다. 프록시는 포트를 포함한 Host를 전달한다.
- 서점 기본값은 `BOOKSTORE_URL=http://localhost:8001`, `BOOKSTORE_PORT=8001`이다. 포트를 바꾸면 출처도 함께 바꾸고 Google callback URI를 다시 등록한다. 관리자는 `ADMIN_URL=http://localhost:8002`, `ADMIN_PORT=8002`, `ADMIN_API_PORT=8081`, `ADMIN_CORE_API_URL=http://localhost:8081`을 사용한다. 포트 변경 시 출처·proxy target·Google callback을 함께 맞춘다.
- `CORE_API_URL`은 서점 서버만 사용하는 신뢰한 내부 API 출처다(호스트 기본 http://localhost:8080, compose에서는 http://nn-core-api-server:8080). 경로·사용자 정보·쿼리가 없는 HTTP(S) 출처만 설정한다. 요청마다 공개 쿠키만 전달하고 관리자·작가 API는 프록시하지 않는다.
- 서점은 adapter-node의 Node 서버다. `npm run build --workspace @novelity/bookstore-web` 후 앱 디렉터리에서 `ORIGIN=http://localhost:8001 node build`로 실행한다. 운영에서는 ORIGIN과 BOOKSTORE_URL을 서점 HTTPS 출처로 맞추고 CORE_API_URL·STUDIO_URL도 함께 구성한다. compose는 이 환경변수를 설정한다. 로그인 상태 확인이 실패해도 공개 홈은 표시된다.
- `.env`에서 STUDIO_URL·BOOKSTORE_URL·ADMIN_URL, SESSION_DOMAIN, ADMIN_EMAILS를 구성한다. PUBLIC_URL은 제거했다. 각 출처 + `/api/auth/google/callback`을 Google에 등록한다. 운영 SSO는 SESSION_DOMAIN=novelity.kr, 로컬은 빈 값이다.
- API_SURFACES 기본값은 studio,store다. 직접 실행할 때 `go run ./cmd/server --surfaces=admin`으로 관리자만 켤 수 있다. admin과 공개 그룹을 함께 고르면 기동에 실패한다. 공개 compose와 dev:api는 studio,store로 고정한다. dev:admin-api는 admin으로 고정하고 BIND_HOST=127.0.0.1을 설정한다. 직접 API를 실행할 때 BIND_HOST를 생략하면 모든 인터페이스에서 받는다.
- 관리자 세션은 설정 가능한 공개 SESSION_TTL(기본 168h)과 별개로 2시간이다. ADMIN_EMAILS는 쉼표로 구분한 검증된 Google 이메일이며 변경 후 API를 재시작한다.
- npm은 루트 lockfile 하나를 쓴다. 앱에서 별도 npm ci를 실행하지 않는다. `npm run check`, `npm run lint`, `npm run build`는 모든 워크스페이스의 해당 명령을 실행한다. 앱만 검사하려면 `npm run check --workspace @novelity/studio-web`을 쓴다. `npm test --workspace @novelity/api-client`는 공통 클라이언트 테스트다. `npm test --workspace @novelity/bookstore-web`는 Host·쿠키·리디렉션·CSRF 전달 회귀 테스트이며 로컬 HTTP 테스트 서버를 사용한다.
- air는 `scripts/dev`에 버전을 고정했고, 처음 실행할 때 `~/.cache/nn-dev/`에 설치한다(go.mod에 넣지 않는다).
- 컨테이너는 웹의 루트 컨텍스트 Dockerfile을 쓴다. 작가·관리자는 nginx SPA, 서점은 Node SSR이다. 공개는 `app`, 관리자는 `admin` 프로필이다. 관리자 API는 host port가 없고 관리자 웹은 127.0.0.1:8002로만 공개한다. public 웹은 admin 네트워크에 속하지 않는다. admin 네트워크는 Google OIDC outbound를 허용한다. 운영 VPN·IP 허용 목록·TLS 내부 게이트웨이는 배포 환경 확정 후 설정하며 공개 프록시에 관리자 upstream을 넣지 않는다.

## 검증 흐름

- **개발 중:** 바꾼 부분의 관련 테스트만 돌린다. Go는 해당 패키지, web은 check와 lint, UI 동작이 바뀌었으면 관련 e2e spec을 실행한다. 전체 e2e는 CI나 명시 요청이 있을 때 돌린다.
- **pre-commit 훅** (`.githooks/pre-commit`): 클론이나 worktree마다 `npm run hooks:install`을 한 번 실행해야 켜진다. 커밋에 포함된 경로를 보고 관련 검사만 실행한다.
  - `apps/core-api/`가 포함되면 go vet과 go test를 돌린다.
  - `apps/*-web/`, `packages/`, 루트 package.json·package-lock.json이 포함되면 모든 웹 워크스페이스의 타입 검사와 prettier를 돌린다. 서점·공통 패키지·루트 manifest 변경에는 서점 전송 테스트도 실행한다. 관리자·공통 패키지·루트 manifest 변경에는 자체 Vite 서버와 Chromium으로 관리자 guard 테스트도 실행한다(최초 `npm run test:e2e:install`).
  - e2e는 커밋에 포함된 `tests/e2e/*.spec.ts`와 `E2E_SPEC_FILES`로 지정한 spec만 실행한다. 전체 suite로 대신하지 않는다. 예: `E2E_SPEC_FILES="tests/e2e/session.spec.ts" git commit …`
  - e2e는 실행 중인 앱(`npm run dev`)을 대상으로 한다. 개발 서버는 변경을 바로 반영하므로 다시 빌드할 필요가 없다.
  - 우회: `SKIP_HOOKS=1`(전체) 또는 `SKIP_E2E=1`(e2e만). 실패를 남긴 채 넘길 때는 이유를 커밋 메시지에 적는다.
  - 훅은 커밋 대상 스냅샷이 아니라 현재 작업 트리를 검사한다. 이를 피하려고 `git stash`를 쓰지 않는다(stash는 worktree끼리 공유된다).
- **서점 e2e:** `tests/e2e/bookstore.spec.ts`는 JavaScript 없는 SSR·OG, 공개/유효/만료 세션, CSRF 거절, 로그아웃과 로그인 실패 복귀를 확인한다. `E2E_BOOKSTORE_URL`의 기본값은 http://localhost:${BOOKSTORE_PORT:-8001}이며, studio는 기존 E2E_BASE_URL을 쓴다. 로컬 쿠키는 포트를 구분하지 않으므로 두 앱을 같은 localhost로 실행한다. 기본 CI는 Google 자격증명이 비어 있는 스택을 사용한다.
- **관리자 검증:** `npm test --workspace @novelity/admin-web`는 Mongo·Google 없이 API mock으로 역할·로그인·로그아웃·만료 가드를 검사한다. 실제 배포 경계는 `tests/e2e/admin.spec.ts`가 공개 admin 404, public 토큰 거절, reader 403, admin 200, 관리자 로그아웃의 공개 세션 보존을 검사한다. 관리자 스택과 ADMIN_EMAILS=e2e-admin@example.com이 필요하다. 다른 테스트 이메일을 쓰면 E2E_ADMIN_EMAIL도 맞춘다. E2E_ADMIN_URL 기본값은 http://localhost:${ADMIN_PORT:-8002}다.
- **e2e 세션:** 실제 Google 로그인은 자동화할 수 없다. 그래서 `tests/e2e/fixtures.ts`의 `user` fixture가 Mongo에 사용자와 세션을 직접 넣고 쿠키를 설정한다. googleSub는 `e2e-…` 형식이고 테스트가 끝나면 삭제한다.

## CI (GitHub Actions)

| 워크플로 | 트리거 | 내용 | 필수 여부 |
|----------|--------|------|-----------|
| `verify-pull-request.yml` | main·develop 대상 PR, main·develop push, 수동 | core-api: gofmt, vet, `go test -race`(Mongo 서비스 컨테이너로 Mongo 통합 테스트 포함) · web(워크스페이스): 루트 npm ci, 전체 check·lint·build와 API 클라이언트·서점 전송·관리자 mock guard 테스트 | 필수 게이트로 지정한다 |
| `e2e.yml` | main 대상 PR, 수동(릴리스 PR은 prepare-release가 실행) | `COMPOSE_PROFILES=app,admin`으로 `docker compose up --build --wait` → Playwright 전체 실행. 실패 시 trace와 리포트를 업로드한다 | 필수 아님 |
| `build-images.yml` | main push, `v*` 태그, 수동(릴리스 태그는 publish-release가 실행) | core-api·studio-web·bookstore-web·admin-web 이미지를 빌드만 한다. 레지스트리가 없어 로그인·푸시는 주석 처리했다. 활성화 방법은 파일 머리말에 있다 | — |
| `prepare-release.yml` | 수동(Actions → Prepare release) | 버전 계산 → `release/vX.Y.Z` 브랜치에 bump·CHANGELOG 커밋 → main 대상 릴리스 PR 생성, 검증·e2e 실행 | — |
| `publish-release.yml` | 릴리스 PR 병합 | 태그·GitHub Release 생성, 이미지 빌드 실행, main → develop 동기화 PR 자동 병합 | — |

- PR 게이트는 PR head가 아니라 base와 병합한 결과를 검사한다. 각자는 통과하지만 합치면 깨지는 경우를 이 단계에서 잡는다.
- e2e는 필수 게이트에서 뺐다. 스택 전체와 브라우저가 필요해 느리고 불안정할 수 있는데, 불안정한 필수 게이트는 우회를 습관으로 만들기 때문이다.
- 브랜치 규칙(Settings → Rules → Rulesets, `main·develop: PR only …`): main과 develop은 PR로만 바꿀 수 있고, 삭제와 force push를 막는다. 필수 검사는 `Verify pull request`의 `core-api`와 `web`이다. 승인 리뷰 수는 0이고 우회 계정은 없다.
- 작업 흐름: 기능 브랜치 → develop 대상 PR → (릴리스) `Prepare release`가 만든 release PR → main. e2e는 main 대상 PR에서만 돌린다.

## 릴리스

저장소 전체가 버전 하나를 쓴다. 태그는 `vX.Y.Z`이고, 세 웹 앱 manifest의 version이 같은 값을 따른다(작가 화면의 `__APP_VERSION__`은 studio manifest 기준). 변경 내역은 `CHANGELOG.md`와 GitHub Release에 남는다.

1. develop 대상 PR에는 라벨을 하나 붙인다. 라벨이 CHANGELOG 분류와 자동 bump를 정한다(`.github/release.yml`, `scripts/release`). PR 제목이 그대로 CHANGELOG 항목이 되므로 변경 내용을 알 수 있게 쓴다.

   | 라벨 | auto bump | CHANGELOG 분류 |
   |------|-----------|----------------|
   | `breaking` | major (1.0 전에는 minor) | 호환성 변경 |
   | `feature` | minor | 새 기능 |
   | `fix` | patch | 버그 수정 |
   | `chore` 또는 없음 | patch | 기타 |
   | `release` | — | 제외(릴리스·동기화 PR 전용) |

2. Actions → **Prepare release** → Run workflow(`auto` 또는 patch·minor·major). 지난 태그 이후 develop을 기준으로 `release/vX.Y.Z` 브랜치와 main 대상 릴리스 PR이 만들어진다. 태그가 없으면 첫 릴리스는 `apps/studio-web/package.json`의 버전이다.
3. 릴리스 PR에서 버전과 CHANGELOG를 확인한다. 릴리스 사이에 손으로 쓴 항목은 `CHANGELOG.md`의 `## Unreleased`(뒤에 ` — 부제`를 붙여도 된다) 절에 둔다. Prepare release가 그 본문을 새 버전 절 맨 위로 옮기고 Unreleased 제목은 지우므로 GitHub Release 본문에도 들어간다(`scripts/test-release`로 확인). 고칠 내용은 그 release 브랜치에 push한다.
4. 릴리스 PR을 **merge commit**으로 병합한다(squash·rebase 금지). 그러면 publish-release가 태그와 GitHub Release를 만들고, 이미지 빌드를 실행하고, main → develop 동기화 PR을 연다. 동기화 PR은 검사가 통과하면 자동 병합된다.

- Action이 `GITHUB_TOKEN`으로 만든 PR·push·태그는 다른 워크플로를 실행시키지 않는다. 그래서 검증·e2e·이미지 빌드는 workflow_dispatch로 직접 실행한다.
- 저장소 설정 "Allow GitHub Actions to create and approve pull requests"와 "Allow auto-merge"가 켜져 있어야 한다.

## 명명·문서 컨벤션

- 기능 폴더는 kebab-case. ADR은 docs/adr/NNN-title.md.
- 새 기능 문서·ADR은 /init-project:feature가 플러그인 양식에서 만든다.
- 상태·체크박스·검증 결과는 실제 작업과 일치시킨다.
- 서비스 디렉토리는 `apps/<name>`이고 compose 서비스명은 `nn-<name>-server`다.
- API 경로는 모두 `/api` 접두사 아래에 둔다(nginx 프록시 기준). 로그인이 필요한 API에는 `internal/http/auth.Service.RequireSession`을 붙이고, 역할이 필요하면 `auth.RequireRole(identity.Author)` 등을 그 다음에 쌓는다. 상태 변경 그룹에는 `SameOrigin`을 적용한다. 핸들러는 `auth.CurrentUser(c)`로 사용자를 읽는다. 도메인 identity는 HTTP를 참조하지 않는다.
- 비밀값은 `.env`에만 둔다(git 제외). 새 설정을 추가하면 `.env.example`과 compose 환경변수를 함께 갱신한다.
- git 제외 파일 중 새 worktree에도 있어야 하는 것(`.env` 등)은 `.worktreeinclude`에 적는다. Claude Code가 worktree를 만들 때 복사한다.
- web UI는 docs/STYLE.md를 따른다. 색은 공통 CSS 토큰만 쓰고, 새 페이지는 `@novelity/ui`의 `Header` + `PageContent`로 만든다. 디자인 토큰 원본은 `packages/ui/src/tokens.css`다. 서점은 자체 화면 틀을 쓴다.

- 웹 이미지 빌드 컨텍스트는 저장소 루트다: `docker build -f apps/studio-web/Dockerfile .`. core-api 컨텍스트는 apps/core-api다. CI 이미지 매트릭스는 context와 dockerfile을 함께 지정한다.
- 릴리스 버전 변경은 `npm version <version> --workspace @novelity/studio-web --workspace @novelity/bookstore-web --workspace @novelity/admin-web --no-git-tag-version`으로 앱 manifest와 루트 lockfile을 함께 갱신한다.

Decided in: nn-l4cg, nn-aa6i, nn-w097
