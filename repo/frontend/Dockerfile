FROM node:20-alpine AS builder
WORKDIR /app
COPY .npmrc package.json package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY . .
RUN npm run build

FROM nginx:alpine
RUN apk add --no-cache openssl
RUN mkdir -p /etc/nginx/certs && openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/nginx/certs/server.key -out /etc/nginx/certs/server.crt \
    -subj "/C=US/ST=Local/L=Local/O=LocalInsights/CN=localhost" 2>/dev/null
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 443
CMD ["nginx", "-g", "daemon off;"]
