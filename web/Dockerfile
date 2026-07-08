# ─── Build Stage ──────────────────────────────────────────────────────────────
FROM node:20-slim AS build

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Build the app
COPY . .
RUN npm run build

# ─── Runtime Stage ───────────────────────────────────────────────────────────
FROM nginx:alpine

# Copy the build output to nginx public folder
COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
