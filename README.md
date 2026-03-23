# Практическое занятие №7
# Саттаров Булат Рамилевич ЭФМО-01-25
# Dockerfile и контейнеризация Go-сервиса


---

## 1. Dockerfile

### Auth (multi-stage)

Используется двухэтапная сборка:

1. Builder — сборка бинарника
2. Runner — минимальный образ для запуска
```Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app
COPY . .

RUN cd services/auth && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o /bin/auth ./services/auth/cmd/auth

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/auth /bin/auth

WORKDIR /app

EXPOSE 8081 50051

CMD ["/bin/auth"]
```
---

### Tasks (multi-stage)
```Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app
COPY . .

RUN cd services/tasks && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o /bin/tasks ./services/tasks/cmd/tasks

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/tasks /bin/tasks

WORKDIR /app

EXPOSE 8082

CMD ["/bin/tasks"]
```
---

## 2. .dockerignore

Исключает лишние файлы из сборки:

- .git
- bin
- tmp
- *.log
- docs
- deploy

Уменьшает размер образа и ускоряет сборку.

---

## 3. Команды сборки и запуска

Запуск:
```bash
cd deploy
docker compose up -d --build
```
![docker compose up.png](docs/screenshots/docker%20compose%20up.png)

Проверка:
![docker ps.png](docs/screenshots/docker%20ps.png)

---

## 4. Переменные окружения

.env:
```env
AUTH_PORT=8081
AUTH_GRPC_PORT=50051

TASKS_PORT=8082
AUTH_GRPC_ADDR=auth:50051
DATABASE_URL=postgres://tasksuser:taskspass@postgres:5432/tasksdb?sslmode=disable

POSTGRES_DB=tasksdb
POSTGRES_USER=tasksuser
POSTGRES_PASSWORD=taskspass
```
.env не коммитится, используется .env.example.

---

## 5. Взаимодействие сервисов

Все сервисы находятся в одной сети docker-compose

- auth доступен как auth:50051
- tasks обращается к auth по сети docker
- postgres доступен как postgres:5432


---

## 6. Проверка

![post.png](docs/screenshots/post.png)
---

