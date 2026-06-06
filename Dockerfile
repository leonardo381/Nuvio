# ---------- Backoffice UI build stage ----------
FROM node:22-bookworm-slim AS ui-builder

WORKDIR /app/ui

COPY ui/package.json ui/package-lock.json ./
RUN npm ci

COPY ui/ ./

# VITE_* values are browser-exposed and must be provided at image build time.
ARG VITE_PB_BACKEND_URL=http://localhost:8090
ARG VITE_PUBLIC_SITE_BASE_URL=http://localhost:3000
ENV VITE_PB_BACKEND_URL=$VITE_PB_BACKEND_URL
ENV VITE_PUBLIC_SITE_BASE_URL=$VITE_PUBLIC_SITE_BASE_URL
RUN npm run build

# ---------- Go build stage ----------
FROM golang:1.24-bookworm AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nuvio ./examples/base

# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/pb_data

COPY --from=go-builder /out/nuvio /app/nuvio

EXPOSE 8090

VOLUME ["/app/pb_data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://127.0.0.1:8090/api/health || exit 1

CMD ["/app/nuvio", "serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
