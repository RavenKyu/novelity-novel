# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project
npx sv create my-app
```

To recreate this project with the same configuration:

```sh
# recreate this project
npx sv@0.17.1 create --template minimal --types ts --add prettier tailwindcss="plugins:none" sveltekit-adapter="adapter:static" --no-download-check --install npm apps/studio-web
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev:studio

# or start the server and open the app in a new browser tab
npm run dev:studio -- --open
```

## Building

To create a production version of your app:

```sh
npm run build --workspace @novelity/studio-web
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## 워크스페이스

설치는 저장소 루트에서 `npm ci` 한 번으로 한다. 루트 `npm run dev:studio`로 실행한다.
공통 컴포넌트·토큰은 `@novelity/ui`, API 타입·fetch는 `@novelity/api-client`에서 가져온다.
이 앱은 작가 공간이며 별도 서점·관리자 앱과의 구조·세션 경계는 docs/adr/002-spaces-identity-sessions.md를 따른다.
