FROM node:20-alpine AS builder

WORKDIR /app

# Must build with repo-root context (see deploy/docker-compose.yml)
# so pnpm-lock.yaml and pnpm-workspace.yaml are accessible.
RUN corepack enable

# Copy workspace config + lockfile
COPY pnpm-lock.yaml pnpm-workspace.yaml ./

# Copy the web app's manifest
COPY apps/web/package.json apps/web/

# Install all workspace dependencies
RUN pnpm install --frozen-lockfile

# Copy the web app source and build
COPY apps/web apps/web
RUN cd apps/web && pnpm build

FROM caddy:2-alpine

COPY --from=builder /app/apps/web/dist /usr/share/caddy
COPY apps/web/Caddyfile /etc/caddy/Caddyfile

EXPOSE 80

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
