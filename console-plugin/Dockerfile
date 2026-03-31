# Build context must be repo root: docker build -f console-plugin/Dockerfile .

# Stage 1: build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY console-plugin/web/package.json console-plugin/web/package-lock.json* ./
RUN npm ci --omit=optional 2>/dev/null || npm install
COPY console-plugin/web/ ./
RUN npm run build

# Stage 2: build Go backend
FROM golang:1.25-alpine AS backend
WORKDIR /go/src
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY api/ api/
COPY internal/ internal/
COPY console-plugin/cmd/ console-plugin/cmd/
COPY console-plugin/pkg/ console-plugin/pkg/
RUN go build -o /plugin ./console-plugin/cmd/plugin

# Stage 3: runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /plugin /app/plugin
COPY --from=frontend /app/web/dist /app/dist
EXPOSE 9443
ENTRYPOINT ["/app/plugin"]
CMD ["--port=9443", "--static-dir=/app/dist"]
