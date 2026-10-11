# Novelity Studio

작가 공간의 SvelteKit SPA다. 시스템 구조와 세션 정책은 [ADR-002](../../docs/adr/002-spaces-identity-sessions.md)를 따른다.

저장소 루트에서 실행한다.

```sh
npm ci
npm run dev:studio
npm run check --workspace @novelity/studio-web
npm run lint --workspace @novelity/studio-web
npm run build --workspace @novelity/studio-web
npm run preview --workspace @novelity/studio-web
```

API와 MongoDB까지 함께 실행하려면 루트 `npm run dev`를 쓴다. 로컬 설정과 callback 등록은 [개발 환경](../../docs/CONVENTIONS.md#개발-환경)을 따른다.

공통 컴포넌트·토큰은 `@novelity/ui`, API 타입·fetch는 `@novelity/api-client`에서 가져온다. Sidebar에 앱 제목·버전·경로·메뉴와 사용자 메뉴 snippet을 전달한다. 세션 상태와 페이지 이동은 앱이 소유한다.
