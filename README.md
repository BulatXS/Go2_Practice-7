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
cp .env.example .env
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

Для запуска стенда используется docker-compose.yml, который поднимает четыре контейнера:
-	postgres — база данных для сервиса tasks
-	auth — сервис авторизации
-	tasks — основной сервис работы с задачами
-	nginx — reverse proxy для приёма HTTPS-запросов

Все контейнеры запускаются в одной сети docker-compose, поэтому могут обращаться друг к другу по именам сервисов, а не по localhost.
Например:
-	tasks подключается к базе по адресу postgres:5432
-	tasks обращается к сервису авторизации по адресу auth:50051

Сервисы auth и tasks не берутся из готового образа, 
а собираются локально по services/auth/Dockerfile и services/tasks/Dockerfile
```yml
services:
  postgres:
    image: postgres:16
    container_name: pr7-postgres
    env_file:
      - .env
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "5433:5432"
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 10

  auth:
    build:
      context: ..
      dockerfile: services/auth/Dockerfile
    container_name: pr7-auth
    command: ["/bin/auth"]
    env_file:
      - .env
    environment:
      AUTH_PORT: ${AUTH_PORT}
      AUTH_GRPC_PORT: ${AUTH_GRPC_PORT}
    ports:
      - "8081:8081"
      - "50051:50051"

  tasks:
    build:
      context: ..
      dockerfile: services/tasks/Dockerfile
    container_name: pr7-tasks
    command: ["/bin/tasks"]
    env_file:
      - .env
    environment:
      TASKS_PORT: ${TASKS_PORT}
      AUTH_GRPC_ADDR: ${AUTH_GRPC_ADDR}
      DATABASE_URL: ${DATABASE_URL}
    ports:
      - "8082:8082"
    depends_on:
      postgres:
        condition: service_healthy
      auth:
        condition: service_started

  nginx:
    image: nginx:latest
    container_name: pr7-nginx
    ports:
      - "8443:8443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./cert.pem:/etc/nginx/tls/cert.pem:ro
      - ./key.pem:/etc/nginx/tls/key.pem:ro
    depends_on:
      tasks:
        condition: service_started
```
---

## 6. Проверка

![post.png](docs/screenshots/post.png)
---

