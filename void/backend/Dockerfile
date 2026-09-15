# ---- Build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/void ./cmd/void

# ---- Runtime stage ----
FROM alpine:3.20
RUN adduser -D -u 10001 void && apk add --no-cache ca-certificates tzdata
USER void
WORKDIR /app
COPY --from=build /out/void /app/void
ENV VOID_HTTP_ADDR=:8080 \
    VOID_DATA_DIR=/app/data
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/void"]
CMD ["serve"]
