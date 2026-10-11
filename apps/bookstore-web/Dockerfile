FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/studio-web/package.json apps/studio-web/package.json
COPY apps/bookstore-web/package.json apps/bookstore-web/package.json
COPY packages/ui/package.json packages/ui/package.json
COPY packages/api-client/package.json packages/api-client/package.json
RUN npm ci
COPY apps/bookstore-web apps/bookstore-web
COPY packages packages
RUN npm run build --workspace @novelity/bookstore-web

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production HOST=0.0.0.0 PORT=3000
COPY --from=build --chown=node:node /app/apps/bookstore-web/build ./build
COPY --from=build --chown=node:node /app/apps/bookstore-web/package.json ./package.json
USER node
EXPOSE 3000
CMD ["node", "build"]
