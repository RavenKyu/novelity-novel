# 브랜치를 로컬에서 직접 확인하기

자동 검사가 닿지 않는 것을 사람이 직접 확인하는 절차다. 실제 Google 로그인, 화면 배치, `npm run dev` 기동이 여기에 든다. 마일스톤 브랜치를 develop에 넣기 전에 한 번 하고, 어떤 브랜치가 실제로 도는 모습을 보고 싶을 때도 쓴다. CI와 pre-commit이 무엇을 검사하는지는 [Conventions의 검증 흐름](CONVENTIONS.md#검증-흐름)에 있다.

## 처음 한 번: Google 로그인 준비

실제 Google 계정으로 로그인해 볼 때만 필요하다. 자격 증명이 없으면 로그인 버튼이 그 공간의 `/login`으로 `not_configured` 오류와 함께 돌아온다. 나머지 확인과 e2e는 그대로 할 수 있다.

1. Google Cloud Console에서 OAuth 클라이언트(웹 애플리케이션)의 승인된 리디렉션 URI에 세 공간의 콜백을 등록한다. 공간마다 자기 출처로 콜백을 받는다([ADR-002](adr/002-spaces-identity-sessions.md)).

   ```
   http://localhost:8000/api/auth/google/callback
   http://localhost:8001/api/auth/google/callback
   http://localhost:8002/api/auth/google/callback
   ```

2. 확인할 checkout의 루트에 `.env`를 둔다. `.env.example`을 복사하고 세 값을 채운다. 나머지는 `scripts/dev`가 포트에서 만든다.

   ```
   GOOGLE_CLIENT_ID=...
   GOOGLE_CLIENT_SECRET=...
   ADMIN_EMAILS=<내 Google 이메일>,e2e-admin@example.com
   ```

   `e2e-admin@example.com`은 관리자 e2e가 쓰는 계정이다. 빼면 관리자 spec이 실패한다.

## 확인용 worktree 만들기

확인할 브랜치를 브랜치 없이(`--detach`) 새 worktree로 연다. 마일스톤 브랜치는 병합 전용 worktree `.worktrees/milestone-<마일스톤 id>`에 이미 체크아웃돼 있어 한 번 더 체크아웃할 수 없다. 병합 전용 worktree에는 `node_modules`와 `.env`를 남기지 않는다.

```sh
git worktree add --detach .worktrees/check <브랜치>     # 예: milestone/<마일스톤 id>
cd .worktrees/check
cp .env.example .env      # 또는 저장소 루트의 .env를 복사
npm ci
```

## 서버 띄우기

`npm run dev` 하나로 MongoDB(compose)와 다섯 프로세스를 한 터미널에 띄운다. Ctrl-C 한 번으로 함께 멈춘다. 파일을 저장하면 바로 반영된다.

| 프로세스 | 주소 | 무엇 |
|---|---|---|
| api | http://localhost:8080 | 공개 API. studio·store 경로만 켠다 |
| studio | http://localhost:8000 | 작업실 SPA. `/api`를 8080으로 넘긴다 |
| bookstore | http://localhost:8001 | 서점 SSR. 서버에서 `CORE_API_URL`로 API를 부른다 |
| admin-api | http://127.0.0.1:8081 | 관리자 전용 API. loopback에만 연다 |
| admin | http://localhost:8002 | 관리자 SPA. `/api`를 8081로 넘긴다 |

- compose 프로젝트 이름이 `novelity-novel`로 고정돼 있어서 어느 worktree에서 띄워도 같은 개발용 MongoDB(127.0.0.1:27017)와 데이터를 쓴다. 브랜치마다 Mongo 서비스 설정(네트워크 등)이 다르면 처음 띄울 때 컨테이너가 다시 만들어진다. 데이터는 볼륨에 남는다.
- 다른 checkout에서 `npm run dev`가 돌고 있으면 포트가 겹친다. 먼저 멈춘다.
- 로컬 쿠키는 포트를 구분하지 않는다. 그래서 localhost에서도 운영(`SESSION_DOMAIN=novelity.kr`)처럼 작업실과 서점이 로그인을 함께 쓴다.

## 확인할 것

### 기동과 공개 경계

다른 터미널에서 다섯 프로세스가 응답하는지, 공개 쪽에 관리자 경로가 아예 없는지 본다.

```sh
curl -s localhost:8080/api/healthz                                        # {"status":"ok"}
curl -s 127.0.0.1:8081/api/healthz                                        # {"status":"ok"}
curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/api/admin/users   # 404
curl -s -o /dev/null -w '%{http_code}\n' localhost:8000/api/admin/users   # 404
```

### 로그인 후 돌아갈 주소

Google 자격 증명이 없어도 된다. API는 돌아갈 주소(`returnTo`)부터 검사한다. 인코딩된 공백은 통과(302)하고, 다른 호스트로 가는 주소는 거절(400)해야 한다.

```sh
curl -s -o /dev/null -w '%{http_code}\n' 'http://localhost:8000/api/auth/google/login?returnTo=%2Fsearch%3Fq%3Da%2520b'   # 302
curl -s -o /dev/null -w '%{http_code}\n' 'http://localhost:8000/api/auth/google/login?returnTo=%2F%2Fevil.test'           # 400
```

두 요청은 작업실 개발 서버(8000)를 거친다. 첫 요청이 400 `invalid_origin`이면 [잘 안 될 때](#잘-안-될-때)를 본다.

### 작업실 (http://localhost:8000)

- 로그인하지 않은 상태로 열면 `/login`으로 간다.
- Google로 로그인하면 작업실 홈으로 돌아온다.
- 로그인한 채 `/login?returnTo=/..//evil.test`를 열면 작업실 홈에 머문다.

### 서점 (http://localhost:8001)

- 로그인하지 않아도 홈이 보인다.
- 페이지 소스(⌘⌥U)에 `og:title` 같은 메타 태그와 본문이 들어 있다. 서버에서 그린 화면이다.
- 작업실에서 로그인한 뒤 서점을 새로고침하면 로그인 상태로 보인다.
- 서점에서 로그아웃하면 작업실도 로그아웃된다. 둘은 세션 하나를 함께 쓴다.

### 관리자 (http://localhost:8002)

- 작업실에 로그인돼 있어도 따로 로그인을 요구한다. 관리자 세션은 공개 세션과 audience가 다르다.
- `ADMIN_EMAILS`에 있는 계정으로 로그인하면 관리자 홈이 보인다.
- 관리자에서 로그아웃해도 작업실 로그인은 남는다. 반대도 같다.
- `ADMIN_EMAILS`에서 내 이메일을 빼고 `npm run dev`를 다시 띄운 뒤 새로고침하면 권한 거절 화면이 나온다. 세션이 살아 있어도 요청마다 목록을 다시 본다.
- 관리자 세션은 로그인한 지 2시간이 지나면 끝난다(절대 만료).

### 화면

- 세 앱을 브라우저 개발자 도구의 기기 모드(⌘⇧M)에서 폭 375px과 데스크톱 폭으로 본다.
- 가로 스크롤이 생기지 않고, 콘솔에 에러가 없어야 한다.

## 자동 테스트도 같이 돌리기

개발 서버를 띄운 채로 같은 worktree의 다른 터미널에서 돌린다. 모두 통과해야 한다.

```sh
npm run test:e2e:install   # 처음 한 번, Chromium 설치
npm run test:e2e
scripts/test-release
```

- e2e는 Google 없이 `tests/e2e/fixtures.ts`가 개발용 MongoDB에 사용자와 세션을 넣고, 끝나면 지운다.
- 개발 서버가 페이지를 처음 컴파일할 때 느려서, 첫 실행에서 재시도 끝에 통과(flaky)하는 테스트가 나올 수 있다. 다시 돌려 통과하면 문제가 아니다.

## 정리

```sh
# npm run dev 터미널에서 Ctrl-C
cd <저장소 루트>
git worktree remove .worktrees/check
```

`.env`도 worktree와 함께 지워진다. Google 값을 계속 쓰려면 지우기 전에 저장소 루트로 복사한다. `git worktree remove`가 거절하면 `--force`를 붙인다. 확인용 복사본이라 잃을 것이 없다.

## 잘 안 될 때

- **로그인·로그아웃이 400 `invalid_origin`:** 개발 서버의 `/api` 프록시가 Host를 대상 주소로 바꿔 보냈다. API는 Host로 공간을 정한다. `apps/studio-web`과 `apps/admin-web`의 vite 설정에서 프록시는 `changeOrigin: false`여야 한다. 문자열 축약형으로 쓰면 Vite가 Host를 바꾼다(nn-pjr4.4u8).
- **Google 화면의 `redirect_uri_mismatch`:** 그 공간의 콜백 URI가 등록되지 않았다. 준비 1번의 세 URI를 확인한다.
- **로그인 버튼이 `/login?error=not_configured`로 돌아온다:** `.env`의 `GOOGLE_CLIENT_ID`·`GOOGLE_CLIENT_SECRET`이 비어 있다.
- **관리자 로그인이 `/login?error=forbidden`으로 돌아온다:** 로그인한 이메일이 `ADMIN_EMAILS`에 없다. Google에서 확인된 이메일만 받는다.
- **`npm run dev`가 `node_modules missing`이나 `go not found`로 멈춘다:** 그 worktree에서 `npm ci`를 하지 않았거나 Go가 없다(`brew install go`).
- **포트가 이미 쓰이는 중:** 다른 checkout의 `npm run dev`나 컨테이너(`docker compose --profile app`)가 같은 포트를 쓴다. 하나만 띄운다.
- **`git worktree add`가 already checked out으로 거절한다:** `--detach`를 빠뜨렸다.

Decided in: nn-l4cg, nn-aa6i, nn-w097, nn-4oyg, nn-pjr4
