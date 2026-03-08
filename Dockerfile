# ---------- Build stage ----------
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o nuvio ./examples/base

# ---------- Runtime stage ----------
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/nuvio /app/nuvio

EXPOSE 8090

CMD ["/app/nuvio", "serve", "--http=0.0.0.0:8090"]