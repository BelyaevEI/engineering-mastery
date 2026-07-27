# Пример использования Protobuf с Go

Этот пример демонстрирует работу с Protocol Buffers для создания gRPC сервиса с REST-интерфейсом через GRPC-Gateway.

## Структура проекта

```
go/
├── Makefile                 # Файлы для генерации кода
├── go.mod                   # Зависимости Go
├── api/                     # Protobuf контракты
│   └── user_service.proto
├── doc/                     # Swagger спецификации (автогенерация)
│   └── user_service.swagger.json
├── pkg/                     # Сгенерированный код (gitignored)
│   ├── user_service.pb.go
│   ├── user_service.pb.gw.go
│   └── user_service_grpc.pb.go
├── server/                  # gRPC + HTTP Gateway сервер
│   └── main.go
└── client/                  # gRPC клиент
    └── main.go
```

## Установка зависимостей

```bash
# Установить зависимости Go
go mod download

# Установить необходимые плагины для protoc
make deps
```

## Генерация кода

```bash
# Сгенерировать Go код из .proto файлов
make generate

# Очистить сгенерированные файлы
make clean
```

В результате генерации будут созданы:
- `pkg/user_service.pb.go` - код protobuf сообщений
- `pkg/user_service.pb.gw.go` - HTTP gateway код
- `pkg/user_service_grpc.pb.go` - gRPC серверный код
- `doc/user_service.swagger.json` - Swagger спецификация

## Запуск примера

### 1. Запустить gRPC + HTTP Gateway сервер

```bash
cd server
go run main.golang
```

Сервер будет запущен на двух портах:
- **gRPC**: `localhost:50051`
- **HTTP**: `localhost:8080`

### 2. Запустить gRPC клиент

В другом терминале:

```bash
cd client
go run main.golang
```

Клиент отправит запрос на получение пользователя с ID `12345` и выведет ответ.

### 3. Использовать REST API через curl

```bash
# Получить пользователя по ID
curl http://localhost:8080/api/v1/user/12345

# Или с JSON ответом
curl -H "Accept: application/json" http://localhost:8080/api/v1/user/12345
```

## Swagger API документация

После генерации Swagger спецификация доступна в файле `doc/user_service.swagger.json`. Вы можете использовать её для:

- Просмотр API в Swagger UI: `swagger-ui-dist` или `swagger ui`
- Генерация SDK для различных языков
- Тестирование API через инструменты вроде curl или Postman

Пример endpoint'а:
```
GET /api/v1/users/{id}
```
