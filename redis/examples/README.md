# Примеры работы с Redis на Go

Этот каталог содержит примеры использования Redis с Go через библиотеку [go-redis/redis](https://github.com/redis/go-redis).

## Структура

Каждый пример находится в отдельной поддиректории:

```
examples/
├── basic/          # Основные операции
├── cluster/        # Redis Cluster
├── geo/            # Гео-запросы
├── pubsub/         # Pub/Sub
├── scripts/        # Lua-скрипты
├── sentinel/       # Sentinel
├── streams/        # Streams
└── transactions/   # Транзакции
```

## Требования

- Go 1.18+
- Redis server (локально или через Docker)

## Запуск примеров

Каждый пример имеет свой `go.mod` и `go.sum`. Запустите его из соответствующей папки:

```bash
# Базовые операции
cd basic && go run main.go

# Транзакции
cd transactions && go run main.go

# Pub/Sub
cd pubsub && go run main.go

# Streams
cd streams && go run main.go

# Lua-скрипты
cd scripts && go run main.go

# Гео-запросы
cd geo && go run main.go

# Sentinel
cd sentinel && go run main.go

# Кластер
cd cluster && go run main.go
```

### Быстрый запуск через Docker

```bash
# Запустить Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Проверить подключение
docker exec -it redis redis-cli ping
# Ответ: PONG

# Запустить пример
cd basic && go run main.go
```

## Описание примеров

| Директория | Описание |
|------------|----------|
| `basic/` | Основные операции: String, List, Hash, Set, Sorted Set |
| `transactions/` | WATCH/MULTI/EXEC, распределенные блокировки |
| `pubsub/` | Паттерн Pub/Sub для messaging |
| `streams/` | Redis Streams для очередей и event sourcing |
| `scripts/` | Lua-скрипты для атомарных операций |
| `geo/` | Геопространственные запросы |
| `sentinel/` | Redis Sentinel для HA |
| `cluster/` | Redis Cluster для horizontal scaling |

## Конфигурация подключения

По умолчанию все примеры подключаются к `localhost:6379`. Для изменения:

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "your-redis-host:6379",
    Password: "your-password", // если есть
    DB:       0,
})
```

## Использование в проекте

Добавьте в свой код:

```go
import "github.com/redis/go-redis/v9"

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer rdb.Close()
    
    // Используйте rdb для операций
    rdb.Set(ctx, "key", "value", 0)
}
```
