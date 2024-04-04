# Backend from ulr-shorter
## Сервис сокращения ссылок
### Стек технологий:
1. Go 1.22
2. Postgres:latest - бд
3. Docker | Docker-compose
4. go-chi/chi/v5 - роутинг/мидлвары
5. database/sql - для подключения к бд

### Запустить докер с Postgres:latest
```
docker run --name url_shortener-pg -p 5432:5432 -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=url_shortener -d postgres
```
### Запустить внутри докер контейнера Postgres:latest
```
psql -U admin -d url_shortener
```

### Другое
## для генирации swagger файла используется команда
```
swag init -d "./" -g "cmd/url-shortener/main.go"
```