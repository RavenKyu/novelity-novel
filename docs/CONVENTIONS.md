# Conventions

> 마지막 갱신: 2026-10-10

## 명령어

core-api 명령은 `scripts/go`로 실행한다. 로컬에 `go`가 있으면 그것을 쓰고, 없으면 `golang:1.27-alpine` 컨테이너로 실행한다.

| 목적 | 실제 명령어 |
|------|-------------|
| 전체 테스트 | `cd apps/core-api && MONGO_TEST_URI='mongodb://novelity:novelity-dev@localhost:27017/?authSource=admin' ../../scripts/go test ./...` (compose의 Mongo 필요, 미설정 시 store 테스트 skip) |
| 관련 테스트 | `cd apps/core-api && ../../scripts/go test ./internal/<pkg>/...` |
| 린트·포맷 | `cd apps/core-api && ../../scripts/go vet ./... && ../../scripts/go fmt ./...` · `cd apps/web && npm run lint` (수정: `npm run format`) |
| 타입 검사 | `cd apps/web && npm run check` |
| 빌드 | `cd apps/web && npm run build` · `docker compose --profile app build` |
| 개발 실행 | 최초 1회 `brew install go`, `cp .env.example .env` 후 값 입력, `npm ci`, `(cd apps/web && npm ci)` → `npm run dev` → http://localhost:8000 (`WEB_PORT`로 변경). 따로 띄우려면 `npm run dev:api` / `npm run dev:web` |
| 컨테이너 실행 | `docker compose --profile app up -d --build` (CI e2e와 같은 운영 유사 구성. 개발 서버와 8000 포트를 같이 쓰므로 동시에 띄우지 않는다) |
| e2e (Playwright) | 최초 1회 `npm ci && npm run test:e2e:install` → 앱 실행(`npm run dev`) 후 `npm run test:e2e` (관련 spec만: `npx playwright test --config tests/e2e/playwright.config.ts tests/e2e/<name>.spec.ts`) |
| 워크플로 린트 | `docker run --rm -v "$PWD":/repo -w /repo rhysd/actionlint:1.7.7` |
| 스모크 | `curl -s localhost:8000/api/healthz` → `{"status":"ok"}` (Mongo 불통 시 503) |
| Mongo 셸 | `docker compose exec nn-mongodb mongosh -u novelity -p novelity-dev` |

## 개발 환경

개발 중에는 MongoDB만 컨테이너로 띄우고 core-api와 web은 호스트에서 실행한다. 저장하면 바로 반영된다. 진입점은 `scripts/dev`(`npm run dev`)다.

| 구성 | 실행 방식 | 변경 반영 |
|------|-----------|-----------|
| nn-mongodb | `docker compose up -d --wait nn-mongodb` (`dev:api`가 자동 실행) | — |
| core-api | air(`apps/core-api/.air.toml`)가 `:8080`으로 실행 | `*.go` 저장 → 재빌드·재시작. 빌드가 실패하면 마지막으로 성공한 빌드가 계속 돈다 |
| web | vite dev 서버가 `:${WEB_PORT:-8000}`으로 실행, `/api`는 `:8080`으로 프록시 | HMR(새로고침 없이 반영) |

- `scripts/dev`는 루트 `.env`를 읽고 `MONGO_ROOT_*`·`MONGO_PORT`로 `MONGO_URI`(localhost)를 만든다. `.env`에 `MONGO_URI`를 직접 적으면 그 값을 쓴다.
- web 개발 서버는 컨테이너와 같은 출처(`localhost:8000`)를 쓴다. 그래서 `PUBLIC_URL`과 Google에 등록한 redirect URI를 그대로 쓸 수 있다.
- air는 `scripts/dev`에 버전을 고정했고, 처음 실행할 때 `~/.cache/nn-dev/`에 설치한다(go.mod에 넣지 않는다).
- 컨테이너 이미지(`Dockerfile`, `nginx.conf`)는 그대로 유지한다. compose에서는 `app` 프로필로만 띄우고, CI e2e가 이 방식을 쓴다.

## 검증 흐름

- **개발 중:** 바꾼 부분의 관련 테스트만 돌린다. Go는 해당 패키지, web은 check와 lint, UI 동작이 바뀌었으면 관련 e2e spec을 실행한다. 전체 e2e는 CI나 명시 요청이 있을 때 돌린다.
- **pre-commit 훅** (`.githooks/pre-commit`): 클론이나 worktree마다 `npm run hooks:install`을 한 번 실행해야 켜진다. 커밋에 포함된 경로를 보고 관련 검사만 실행한다.
  - `apps/core-api/`가 포함되면 go vet과 go test를 돌린다.
  - `apps/web/`이 포함되면 svelte-check와 prettier를 돌린다.
  - e2e는 커밋에 포함된 `tests/e2e/*.spec.ts`와 `E2E_SPEC_FILES`로 지정한 spec만 실행한다. 전체 suite로 대신하지 않는다. 예: `E2E_SPEC_FILES="tests/e2e/session.spec.ts" git commit …`
  - e2e는 실행 중인 앱(`npm run dev`)을 대상으로 한다. 개발 서버는 변경을 바로 반영하므로 다시 빌드할 필요가 없다.
  - 우회: `SKIP_HOOKS=1`(전체) 또는 `SKIP_E2E=1`(e2e만). 실패를 남긴 채 넘길 때는 이유를 커밋 메시지에 적는다.
  - 훅은 커밋 대상 스냅샷이 아니라 현재 작업 트리를 검사한다. 이를 피하려고 `git stash`를 쓰지 않는다(stash는 worktree끼리 공유된다).
- **e2e 세션:** 실제 Google 로그인은 자동화할 수 없다. 그래서 `tests/e2e/fixtures.ts`의 `user` fixture가 Mongo에 사용자와 세션을 직접 넣고 쿠키를 설정한다. googleSub는 `e2e-…` 형식이고 테스트가 끝나면 삭제한다.

## CI (GitHub Actions)

| 워크플로 | 트리거 | 내용 | 필수 여부 |
|----------|--------|------|-----------|
| `verify-pull-request.yml` | main 대상 PR, main push | core-api: gofmt, vet, `go test -race`(Mongo 서비스 컨테이너로 store 테스트 포함) · web: npm ci, check, lint, build | 필수 게이트로 지정한다 |
| `e2e.yml` | main 대상 PR, 수동 | `COMPOSE_PROFILES=app`으로 `docker compose up --build --wait` → Playwright 전체 실행. 실패 시 trace와 리포트를 업로드한다 | 필수 아님 |
| `build-images.yml` | main push, `v*` 태그, 수동 | 두 서비스 이미지를 빌드만 한다. 레지스트리가 없어 로그인·푸시는 주석 처리했다. 활성화 방법은 파일 머리말에 있다 | — |

- PR 게이트는 PR head가 아니라 base와 병합한 결과를 검사한다. 각자는 통과하지만 합치면 깨지는 경우를 이 단계에서 잡는다.
- e2e는 필수 게이트에서 뺐다. 스택 전체와 브라우저가 필요해 느리고 불안정할 수 있는데, 불안정한 필수 게이트는 우회를 습관으로 만들기 때문이다.
- 아직 GitHub 원격 저장소가 없어서 워크플로는 push한 뒤부터 동작한다. 필수 체크는 Settings → Branches(또는 Rulesets)에서 `Verify pull request`의 두 잡을 지정한다.

## 명명·문서 컨벤션

- 기능 폴더는 kebab-case. ADR은 docs/adr/NNN-title.md.
- 새 기능 문서·ADR은 /init-project:feature가 플러그인 양식에서 만든다.
- 상태·체크박스·검증 결과는 실제 작업과 일치시킨다.
- 서비스 디렉토리는 `apps/<name>`이고 compose 서비스명은 `nn-<name>-server`다.
- API 경로는 모두 `/api` 접두사 아래에 둔다(nginx 프록시 기준). 로그인이 필요한 API에는 `auth.Service.RequireSession` 미들웨어를 붙이고, 핸들러에서 `auth.CurrentUser(c)`로 사용자를 꺼낸다.
- 비밀값은 `.env`에만 둔다(git 제외). 새 설정을 추가하면 `.env.example`과 compose 환경변수를 함께 갱신한다.
- web UI는 docs/STYLE.md를 따른다. 색은 `layout.css`의 토큰만 쓰고, 새 페이지는 `Header` + `PageContent`로 만든다.
