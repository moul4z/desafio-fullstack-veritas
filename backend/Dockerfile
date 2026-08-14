# ---- build ----------------------------------------------------------------
FROM golang:1.22-alpine AS build
WORKDIR /src

# Sem dependências externas: basta o go.mod para o cache de camadas funcionar.
COPY go.mod ./
COPY *.go ./

# CGO desligado para gerar um binário estático que roda na imagem scratch.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/kanban .

# ---- runtime --------------------------------------------------------------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /bin/kanban /app/kanban

# O volume mantém data/tasks.json entre reinícios do container.
RUN mkdir -p /app/data && chown -R app:app /app
USER app

ENV PORT=8080 DATA_FILE=/app/data/tasks.json
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1

CMD ["/app/kanban"]
