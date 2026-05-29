ARG APP=admin

FROM node:22-alpine AS builder
ARG APP
WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
COPY frontend/packages/api-client/package.json packages/api-client/
COPY frontend/apps/admin/package.json apps/admin/
COPY frontend/apps/member/package.json apps/member/
RUN npm ci

COPY frontend/ .
RUN npm run build:${APP}

FROM nginx:1.27-alpine
ARG APP
COPY --from=builder /app/apps/${APP}/dist /usr/share/nginx/html
COPY docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
