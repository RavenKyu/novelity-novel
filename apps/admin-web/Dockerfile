FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/studio-web/package.json apps/studio-web/package.json
COPY apps/bookstore-web/package.json apps/bookstore-web/package.json
COPY apps/admin-web/package.json apps/admin-web/package.json
COPY packages/ui/package.json packages/ui/package.json
COPY packages/api-client/package.json packages/api-client/package.json
RUN npm ci
COPY apps/admin-web apps/admin-web
COPY packages packages
RUN npm run build --workspace @novelity/admin-web

FROM nginx:alpine
COPY apps/admin-web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/admin-web/build /usr/share/nginx/html
EXPOSE 80
