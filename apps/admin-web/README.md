# Novelity Admin

관리자 전용 SvelteKit SPA(adapter-static) 골격. 공통 `@novelity/ui`의 Sidebar·Header·PageContent를 사용한다. 사용자·작품·신고 관리 기능은 후속 작업이다. `GET /api/admin/users`는 빈 목록을 반환하는 권한 확인용 자리다.

루트에서 `npm ci` 후 `npm run dev:admin-api`와 `npm run dev:admin`을 실행한다. 기본 웹 출처는 http://localhost:8002, 관리자 API는 loopback :8081이다. `npm run dev`는 공개·관리자 앱을 함께 실행한다. `.env`의 `ADMIN_PORT`, `ADMIN_API_PORT`, `ADMIN_URL`, `ADMIN_CORE_API_URL`을 함께 맞춘다.

관리자 Google callback은 `ADMIN_URL` + `/api/auth/google/callback`이다. `ADMIN_EMAILS`에 검증된 Google 이메일을 쉼표로 나열하고 API를 재시작한다. 공개 로그인과 별도로 관리자 로그인이 필요하다. 운영 출처는 https://admin.novelity.kr이며 `nn_admin_session`은 host-only·HttpOnly·Secure·SameSite=Lax·Path=/·절대 TTL 2시간이다. 로컬 HTTP에서만 Secure가 생략된다. 서버는 admin audience와 현재 이메일 허용 목록을 요청마다 확인하며 관리자 로그아웃은 공개 세션을 지우지 않는다.

검증은 루트 `npm run check`, `npm run lint`, `npm run build`, `npm test --workspace @novelity/admin-web`로 실행한다. 최초 브라우저 설치는 `npm run test:e2e:install`이다. 관리자 guard 테스트는 자체 Vite 서버와 API 응답 mock을 쓰며 Mongo·Google 자격증명을 요구하지 않는다. 실제 세션·권한 경계는 core-api 테스트가 검증한다.

컨테이너는 `docker compose --profile admin up -d --build`로 실행한다. 관리자 API는 host port 없이 별도 admin 네트워크에 두고 웹만 127.0.0.1:8002로 바인딩한다. 공개 `app` 프로필은 관리자 서비스를 띄우지 않는다. admin 네트워크는 Google OIDC 통신을 위해 outbound를 허용한다. 운영 VPN·IP 허용 목록과 TLS 내부 게이트웨이는 배포 환경 확정 후 설정해야 한다. 공개 nginx에 admin upstream을 추가하거나 관리자 포트를 외부 주소에 바인딩하지 않는다.
