# Protocol Buffers (protobuf)

Protocol Buffers — это язык сериализации данных от Google. Используется для структурирования данных, хранения и обмена данными между сервисами.

## Оглавление

- [Что включено](#что-включено)
- [Основные концепции и устройство gRPC](#основные-концепции-и-устройство-grpc)
  - [gRPC — что это?](#grpc---что-это)
  - [Основные концепции](#основные-концепции)
  - [Архитектура gRPC](#архитектура-grpc)
  - [Каналы связи](#каналы-связи)
  - [Преимущества gRPC](#преимущества-grpc)
  - [Базовая структура `.proto` файла](#базовая-структура-proto-файла)
  - [Базовый поток команд в gRPC](#базовый-поток-команд-в-grpc)
- [QA](#qa)
  - [Часто проверяемые items при работе с gRPC/protobuf](#часто-проверяемые-items-при-работе-с-grpcprotobuf)
  - [Обработка ошибок](#обработка-ошибок)
  - [Тестирование](#тестирование)
  - [Отладка](#отладка)
- [FAQ](#faq)
  - [Общие вопросы](#общие-вопросы)
  - [gRPC вопросы](#grpc-вопросы)
  - [Производительность](#производительность)
  - [Разработка](#разработка)
  - [Отладка и диагностика](#отладка-и-диагностика)
  - [Безопасность](#безопасность)
- [TLS и mTLS в gRPC](#tls-и-mtls-в-grpc)
  - [Типы защиты](#типы-защиты)
    - [1. TLS (Transport Layer Security)](#1-tls-transport-layer-security)
    - [2. mTLS (Mutual TLS)](#2-mtls-mutual-tls)
  - [Сравнение TLS vs mTLS](#сравнение-tls-vs-mtls)
  - [Управление сертификатами](#управление-сертификатами)
  - [Генерация сертификатов (скрипт)](#генерация-сертификатов-скрипт)
  - [Проверка сертификатов](#проверка-сертификатов)
  - [Интеграция с Kubernetes](#интеграция-с-kubernetes)
  - [Миграция с HTTP/JSON на gRPC with TLS](#миграция-с-httpjson-grpc-with-tls)
  - [Best Practices](#best-practices)
- [Полезные ссылки](#полезные-ссылки)

## Что включено

- `examples/` — примеры использования protobuf с Go

## Основные концепции и устройство gRPC

### gRPC — что это?

gRPC (Google Remote Procedure Call) — это современный фреймворк для построения распределённых систем,allowing services to communicate с использованием HTTP/2 для передачи сообщений. gRPC использует Protocol Buffers как язык описания интерфейсов и формат сериализации данных.

### Основные концепции

| Концепция | Описание |
|-----------|----------|
| **Service** | Интерфейс удалённого вызова процедур, определяемый в `.proto` файле |
| **Method** | Функция в сервисе, вызываемая клиентом (unary, server streaming, client streaming, bidirectional streaming) |
| **Unary RPC** | Простейший тип: клиент отправляет один запрос и получает один ответ |
| **Server Streaming** | Клиент отправляет один запрос, сервер возвращает поток данных |
| **Client Streaming** | Клиент отправляет поток данных, сервер возвращает один ответ |
| **Bidirectional Streaming** | Оба участника обмениваются потоками данных одновременно |

### Архитектура gRPC

```
┌─────────────────┐                              ┌─────────────────┐
│   Client App    │                              │   Server App    │
│                 │                              │                 │
│  ┌───────────┐  │           HTTP/2             │  ┌───────────┐  │
│  │ gRPC      │  │─────────────────────────────►│  │ gRPC      │  │
│  │ Client    │  │     (protobuf over HTTP/2)   │  │ Server    │  │
│  └───────────┘  │                              │  └───────────┘  │
└─────────────────┘                              └─────────────────┘
```

### Каналы связи

- **HTTP/2** — транспортный протокол с multiplexing, header compression и flow control
- **Protocol Buffers** — формат сериализации (легковесный, быстрый, строго типизированный)

### Преимущества gRPC

- **Производительность**: бинарный формат protobuf быстрее JSON/XML
- **Типизация**: строгая схема данных из `.proto` файлов
- **Мультиплексинг**: HTTP/2 позволяет использовать одно соединение для нескольких потоков
- **Поддержка потоков**: встроенная поддержка stream-коммуникации
- **Cross-language**: официальная поддержка для множества языков (Go, Java, Python, Node.js и др.)

### Базовая структура `.proto` файла

```protobuf
syntax = "proto3";

package example;

// Определение сервиса
service Greeter {
  // Unary RPC
  rpc SayHello (HelloRequest) returns (HelloResponse);
  
  // Server streaming
  rpc ListMessages (ListRequest) returns (stream Message);
}

// Запрос
message HelloRequest {
  string name = 1;
}

// Ответ
message HelloResponse {
  string message = 1;
}
```

### Базовый поток команд в gRPC

1. **Определение сервиса** в `service.proto`
2. **Генерация кода**: `protoc --go_out=. --go-grpc_out=. service.proto`
3. **Реализация сервера** с логикой методов
4. **Создание клиента** с автоматически сгенерированным кодом
5. **Вызов методов** как локальных функций

## QA

### Часто проверяемые items при работе с gRPC/protobuf

| Категория | Что проверять | Инструменты/Подсказки |
|-----------|---------------|----------------------|
| **.proto файлы** | - Корректность синтаксиса<br>- Уникальность ID полей<br>- Согласованность package имён | `protoc --check_unimplemented` |
| **Генерация кода** | - Все `.pb.go` файлы обновлены<br>- Отсутствие ошибок компиляции | `protoc --go_out=. --go-grpc_out=. *.proto` |
| **API контракт** | - Изменения в `.proto` не ломают клиентов<br>- Backward compatibility | Versioning: v1, v2; `protoc --decode` для тестов |
| **Сервер** | - Все методы реализованы<br>- Обработка ошибок (status codes)<br>- Stream закрывается корректно | Logs, `grpcurl`, тесты |
| **Клиент** | - Правильная обработка stream<br>- Retry логика<br>- Timeout и cancellation | Context, telemetry |

### Обработка ошибок

gRPC использует стандартные статус-коды:

| Код | Описание | Когда использовать |
|-----|----------|-------------------|
| `OK` | 0 | Успех |
| `INVALID_ARGUMENT` | 3 | Клиент передал неверные аргументы |
| `NOT_FOUND` | 5 | Ресурс не найден |
| `ALREADY_EXISTS` | 6 | Ресурс уже существует |
| `PERMISSION_DENIED` | 7 | Нет прав на операцию |
| `RESOURCE_EXHAUSTED` | 8 | Лимиты превышены |
| `FAILED_PRECONDITION` | 9 | Система не в том состоянии |
| `ABORTED` | 10 | Конфликт параллельных операций |
| `OUT_OF_RANGE` | 11 | Выход за границы |
| `UNIMPLEMENTED` | 12 | Метод не реализован |
| `INTERNAL` | 13 | Внутренняя ошибка сервера |
| `UNAVAILABLE` | 14 | Сервис недоступен |
| `DATA_LOSS` | 15 | Потеря данных |
| `UNKNOWN` | 2 | Неизвестная ошибка |

### Тестирование

```bash
# Проверка .proto файлов
protoc --proto_path=. --decode=example.HelloRequest example.proto < request.bin

# Тестовый вызов через grpcurl (установка: brew install grpcurl)
grpcurl -plaintext -d '{"name": "test"}' localhost:50051 example.Greeter/SayHello

# Валидация схемы
protoc --proto_path=. --include_imports service.proto --descriptor_set_out=desc.pb
```

### Отладка

- **Логирование**: логируй full error message с `status.FromError()`
- **Tracing**: используй `context` для передачи trace ID
- **Метрики**: count ошибок по статус-кодам, latency по методам

## FAQ

### Общие вопросы

**Q: В чём разница между proto2 и proto3?**

|Proto2|Proto3|
|------|------|
| Опциональные поля по умолчанию | Поля optional (начиная с proto3.15)|
| Необязательные поля имеют default| Нет default значений|
| Использует `optional` ключевое слово| Поля по умолчанию считаются optional|
| Более строгая типизация| Более простой синтаксис|

**Q: Почему ID полей должны быть уникальными?**

ID используются для идентификации полей в бинарном представлении. Если ID совпадут, protobuf не сможет корректно распознать данные.

**Q: Можно ли менять ID полей после релиза?**

**Нельзя!** Это нарушает backward compatibility. При изменении ID старые клиенты будут неправильно читать данные.

---

### gRPC вопросы

**Q: Когда использовать streaming, а когда unary RPC?**

| Сценарий | Рекомендуемый тип |
|----------|-------------------|
| Клиент запрашивает один ответ | Unary |
| Сервер возвращает много данных (файл, логи) | Server Streaming |
| Клиент отправляет поток (загрузка, события) | Client Streaming |
| Оба участника обмениваются потоками (чат, игра) | Bidirectional |

**Q: Как обеспечить backward compatibility при изменении сервиса?**

- Добавляй новые методы/поля с новыми ID (не меняй существующие)
- Используй versioning в названиях сервисов: `GreeterV1`, `GreeterV2`
- Деплой новых версий параллельно, постепенно переключай клиентов

**Q: gRPC работает только с protobuf?**

Нет, формат можно заменить на any (в т.ч. JSON), но protobuf — рекомендуемый и оптимальный вариант.

---

### Производительность

**Q: Почему бинарный формат эффективнее JSON?**

См. раздел "Преимущества gRPC". Кратко:
- Размер данных на 50-80% меньше (нет дублирования ключей)
- Скорость парсинга в 5-20 раз выше
- Меньше memory footprint (нет промежуточных объектов)

**Q: Какой overhead у gRPC по сравнению с REST/JSON?**

- **Latency**: +1-2ms из-за HTTP/2 и protobuf
- **Throughput**: gRPC обходит REST при нагрузке >100 req/s
- **Потребление ресурсов**: на 30-50% меньше CPU/памяти на сервере

**Q: Нужно ли сжатие данных (gzip) при использовании protobuf?**

Обычно нет — protobuf уже сжат эффективно. Gzip может добавить overhead на CPU без существенной выгоды для размера.

---

### Разработка

**Q: Как проверить, что .proto файл корректен?**

```bash
# Проверка синтаксиса
protoc --proto_path=. --descriptor_set_out=/dev/null service.proto

# Генерация кода (проверка + генерация)
protoc --go_out=. --go-grpc_out=. *.proto
```

**Q: Как обновить .proto и не сломать клиентов?**

1. Добавь новые поля с новыми ID
2. Не удаляй существующие поля (можно пометить как `reserved`)
3. Не меняй типы полей
4. Добавь миграцию/тесты

**Q: Что такое reserved в .proto?**

Зарезервированные ID или имена, которые нельзя использовать:

```protobuf
message User {
  reserved 10 to 20;
  reserved "old_field", "deprecated";
}
```

---

### Отладка и диагностика

**Q: Как протестировать gRPC метод локально?**

Используй `grpcurl` (установка: `brew install grpcurl`):

```bash
# Unary RPC
grpcurl -plaintext -d '{"name": "test"}' localhost:50051 example.Greeter/SayHello

# List методов
grpcurl -plaintext localhost:50051 list

# Describe сервиса
grpcurl -plaintext localhost:50051 describe example.Greeter
```

**Q: Как понять, почему клиент не получает ответ?**

Проверь по шагам:
1. Сервер запущен и слушает на порту (`netstat -an | grep 50051`)
2. Клиент использует правильный IP/порт
3. нет таймаута (`context.WithTimeout`)
4. логи сервера на предмет ошибок

**Q: Как отладить ошибку "stream terminated by RST_STREAM"?**

Типичные причины:
- Сервер отправил ERROR код в stream
- Клиент отменил запрос (canceled context)
- Проблемы с HTTP/2 connection (ping timeout)

---

### Безопасность

**Q: Нужно ли шифрование для gRPC?**

Да, в продакшене используй **mTLS** или **TLS**:

```go
// Сервер с TLS
creds, _ := credentials.NewServerTLSFromFile(cert, key)
s := grpc.NewServer(grpc.Creds(creds))

// Клиент с TLS
conn, _ := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
```

**Q: Как защититься от атак через protobuf?**

- Валидируй входные данные (protobuf не делает этого автоматически)
- Ограничивай размер сообщений (`grpc.MaxRecvMsgSize`)
- Используй `proto.Unmarshal` с проверкой ошибок

---

## TLS и mTLS в gRPC

### Типы защиты

#### 1. TLS (Transport Layer Security)

Односторонняя аутентификация — клиент проверяет сервер, сервер не проверяет клиента.

```
Client  ──────►  Server (verified)
  │              └─ certificates trusted by CA
  └── encrypted channel
```

**Используется, когда:**
- Сервер публичный (API для внешних клиентов)
- Клиенты — браузеры или мобильные приложения
- Аутентификация реализуется на уровне приложения (токены, сессии)

**Пример Go:**

```go
// Сервер
creds, _ := credentials.NewServerTLSFromFile("server.crt", "server.key")
s := grpc.NewServer(grpc.Creds(creds))

// Клиент
conn, _ := grpc.Dial(
    "api.example.com:443",
    grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "api.example.com")),
)
```

---

#### 2. mTLS (Mutual TLS)

Двусторонняя аутентификация — клиент и сервер проверяют друг друга.

```
Client (verified)  ◄──►  Server (verified)
   certificates          certificates
   trusted by server     trusted by client
```

**Используется, когда:**
- Сервис-к-сервису коммуникация в микросервисной архитектуре
- Требуется строгий контроль доступа
- Zero-trust security model

**Пример Go:**

```go
// Генерация сертификатов (однажды)
// ./gen-certs.sh
# Серверный сертификат
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key -out server.crt \
  -subj "/CN=grpc-server"

# Клиентский сертификат
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout client.key -out client.crt \
  -subj "/CN=grpc-client"

// Сервер - проверяет клиентский сертификат
serverCreds, _ := credentials.NewServerTLSFromFile("server.crt", "server.key")
clientCreds := credentials.NewClientCertAuth([]byte(`CA cert content`))
s := grpc.NewServer(
    grpc.Creds(serverCreds),
    grpc.ClientCertAuth(clientCreds),
)

// Клиент - проверяет серверный сертификат и отправляет свой
clientCreds, _ := credentials.NewClientTLSFromFile("server.crt", "grpc-server")
clientKeyCert, _ := tls.X509KeyPair(clientCert, clientKey)
creds := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{clientKeyCert},
    ServerName:   "grpc-server",
})
conn, _ := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
```

---

### Сравнение TLS vs mTLS

| Критерий | TLS | mTLS |
|----------|-----|------|
| **Аутентификация** | Сервер → Клиент | Сервер ↔ Клиент |
| **Безопасность** | Выше, чем HTTP | Максимальная |
| **Сложность** | Простая | Сложнее (управление сертификатами) |
| **Использование** | Внешние API | Микросервисы, внутренние сервисы |
| **Overhead** | Небольшой | Умеренный |

---

### Управление сертификатами

#### Структура директорий

```
tls/
├── ca/
│   ├── ca.crt          # Root CA certificate
│   └── ca.key          # Root CA key (храни в секрете!)
├── server/
│   ├── server.crt      # Server certificate (signed by CA)
│   ├── server.key      # Server private key
│   └── server.csr      # Certificate signing request
└── client/
    ├── client.crt      # Client certificate (signed by CA)
    ├── client.key      # Client private key
    └── client.csr      # Certificate signing request
```

#### Генерация сертификатов (скрипт)

```bash
#!/bin/bash
# gen-certs.sh

# Создаём CA
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 \
  -out ca.crt -subj "/CN=MyCA/O=MyCompany"

# Генерируем серверный сертификат
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/CN=grpc-server/O=MyCompany"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365 -sha256

# Генерируем клиентский сертификат
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/CN=grpc-client/O=MyCompany"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 365 -sha256
```

---

### Проверка сертификатов

```bash
# Проверить сертификат сервера
openssl x509 -in server.crt -text -noout

# Проверить сертификат клиента
openssl x509 -in client.crt -text -noout

# Проверить цепочку доверия
openssl verify -CAFile ca.crt server.crt
openssl verify -CAFile ca.crt client.crt

# Посмотреть детали сертификата
openssl x509 -in server.crt -noout -issuer -dates
```

---

### Интеграция с Kubernetes

#### Secret для TLS

```yaml
# tls-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: grpc-tls
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0t...
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0t...
---
# Клиентский CA
apiVersion: v1
kind: Secret
metadata:
  name: grpc-ca
type: Opaque
data:
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0t...
```

#### Под с mTLS

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: grpc-client
spec:
  containers:
  - name: client
    image: grpc-client:latest
    env:
    - name: TLS_ENABLED
      value: "true"
    - name: CA_CERT
      value: "/etc/ssl/certs/ca.crt"
    volumeMounts:
    - name: tls-secret
      mountPath: /etc/grpc/tls
      readOnly: true
    - name: ca-secret
      mountPath: /etc/ssl/certs
      readOnly: true
  volumes:
  - name: tls-secret
    secret:
      secretName: grpc-tls
  - name: ca-secret
    secret:
      secretName: grpc-ca
```

---

### Миграция с HTTP/JSON на gRPC with TLS

#### Этап 1: Подготовка

1. Сгенерируй сертификаты
2. Обнови конфиги сервисов
3. Настрой DNS (SAN в сертификате должен содержать hostname)

#### Этап 2: Деплой

1. Деплой сервера с TLS (parallel с HTTP)
2. Деплой клиентов с TLS
3. Трансфер трафика на gRPC

#### Этап 3: Валидация

```bash
# Проверить, что gRPC работает
grpcurl -cacert ca.crt -proto service.proto localhost:443 list

# Проверить сертификат
openssl s_client -connect localhost:443 -tls1_2
```

---

### Best Practices

| Практика | Описание |
|----------|----------|
| **Используй короткий TTL** | Сертификаты на 90-365 дней (не 5 лет) |
| **Автоматизируй ротацию** | cert-manager в Kubernetes, Let's Encrypt |
| **Храни ключи в секрете** | Не коммить `.key` файлы в git |
| **Проверяй SAN** | Сертификат должен содержать IP и hostname в SAN |
| **Используй CRL/OCSP** | Проверяй отозванность сертификатов |
| **Логируй аутентификацию** | Записывай неудачные попытки подключения |

---

## Полезные ссылки

- [Официальная документация](https://protobuf.dev/)
- [Protocol Buffers Go Guide](https://protobuf.dev/getting-started/gotutorial/)
- [gRPC Documentation](https://grpc.io/docs/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
