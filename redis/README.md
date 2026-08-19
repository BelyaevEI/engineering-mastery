# Redis

## Что такое Redis?

Redis (Remote Dictionary Server) — это open-source хранилище структурированных данных, работающее в памяти (in-memory). Это не просто key-value хранилище, а сервер структурированных данных с поддержкой множества типов данных и расширенных операций.

**Основные характеристики:**
- **Скорость:** Операции происходят в памяти (RAM), что обеспечивает микросекундное время отклика
- **Структурированные данные:** Поддержка строк, списков, множеств, хэшей, сортированных множеств
- **Персистентность:** Возможность сохранения данных на диск (RDB и AOF)
- **Расширяемость:** Поддержка кластеров и репликации

## Типы данных

### 1. String (Строки)
Самый простой тип данных. Строки в Redis бинарно-безопасны и могут содержать любые данные (текст, JSON, бинарные данные).

```bash
SET key value
GET key
```

**Особенности:**
- Максимальный размер: 512 MB
- Поддержка числовых операций (INCR, DECR)
- Битовые операции (GETBIT, SETBIT)

### 2. List (Списки)
Упорядоченный список строк. Реализован как двусвязный список.

```bash
LPUSH key value    # Добавить в начало
RPUSH key value    # Добавить в конец
LPOP key           # Удалить из начала
RPOP key           # Удалить из конца
LRANGE key start end  # Получить диапазон
```

**Особенности:**
- O(1) для добавления/удаления элементов
- O(N) для доступа по индексу
- Максимальная длина: ~2^32 элементов

### 3. Set (Множества)
Неупорядоченный набор уникальных строк.

```bash
SADD key member    # Добавить элемент
SMEMBERS key       # Получить все элементы
SREM key member    # Удалить элемент
SUNION key1 key2   # Объединение множеств
SINTER key1 key2   # Пересечение множеств
SDIFF key1 key2    # Разность множеств
```

**Особенности:**
- Автоматическая уникальность элементов
- O(1) для добавления, удаления и проверки наличия
- Максимальное количество элементов: ~2^32

### 4. Hash (Хэши)
Словарь из полей и значений. Подходит для представления объектов.

```bash
HSET key field value    # Установить значение поля
HGET key field          # Получить значение поля
HGETALL key             # Получить все поля и значения
HDEL key field          # Удалить поле
HLEN key                # Количество полей
```

**Особенности:**
- Эффективное хранение объектов
- Меньше памяти по сравнению с сериализацией в строку

### 5. Sorted Set (Сортированные множества)
Множество, где каждый элемент имеет associated score, определяющий порядок.

```bash
ZADD key score member    # Добавить элемент с очками
ZRANGE key start end     # Получить диапазон (по возрастанию)
ZREVRANGE key start end  # Получить диапазон (по убыванию)
ZREMRANGEBYSCORE key min max  # Удалить по диапазону очков
ZCARD key                # Количество элементов
ZSCORE key member        # Получить очки элемента
```

**Особенности:**
- O(log N) для добавления/удаления
- Поддержка рангов (позиций в отсортированном списке)
- Используется для топ-листов, пагинации, сортировки

## Время жизни ключей (TTL)

Redis позволяет задать время жизни ключа, после которого он автоматически удаляется.

```bash
EXPIRE key seconds   # Установить TTL в секундах
PEXPIRE key ms       # TTL в миллисекундах
EXPIREAT key timestamp  # Установить дату/время истечения
TTL key              # Оставить TTL (в секундах, -1 = без TTL, -2 = не существует)
PTTL key             # TTL в миллисекундах
PERSIST key          # Убрать TTL
```

## Основные команды

### Работа со строками

```bash
# Установка и получение
SET key value
GET key

# Числовые операции
INCR key        # Инкремент (если значение число)
INCRBY key N    # Инкремент на N
DECR key        # Декремент
DECRBY key N    # Декремент на N

# Строковые операции
APPEND key value    # Добавить к строке
strlen key          # Длина строки
GETRANGE key start end  # Подстрока
SETRANGE key offset value  # Замена части строки

# Атомарные операции
SETNX key value     # Установить, если не существует (SET if Not eXists)
GETSET key value    # Получить старое значение, установить новое

# Партионная обработка
MSET key1 val1 key2 val2  # Множественная установка
MGET key1 key2            # Множественное получение
```

### Работа со списками

```bash
# Добавление элементов
LPUSH key value [value ...]   # В начало
RPUSH key value [value ...]   # В конец

# Извлечение элементов
LPOP key    # Из начала
RPOP key    # Из конца
RPOPLPUSH source destination  # Атомно переместить из конца одного в начало другого

# Доступ по индексу
LINDEX key index      # Получить элемент по индексу
LSET key index value  # Установить элемент по индексу

# Диапазон и длина
LRANGE key start stop  # Получить диапазон
LLEN key              # Длина списка

# Удаление
LREM key count value  # Удалить первые 'count' вхождений value
LTRIM key start stop  # Обрезать список
```

**Пример использования (очередь):**
```bash
# Добавление в очередь
RPUSH queue message

# Извлечение из очереди
LPOP queue

# Blocking pop (ожидание, пока список не станет непустым)
BLPOP queue 0
BRPOP queue 0
```

### Работа с множествами

```bash
# Добавление и удаление
SADD key member [member ...]
SREM key member [member ...]

# Информация
SMEMBERS key      # Все элементы
SCARD key         # Количество элементов
SISMEMBER key member  # Проверка наличия

# Операции над множествами
SUNION key [key ...]     # Объединение
SINTER key [key ...]     # Пересечение
SDIFF key [key ...]      # Разность

# Операции с результатом
SUNIONSTORE dest key [key ...]
SINTERSTORE dest key [key ...]
SDIFFSTORE dest key [key ...]
```

### Работа с хэшами

```bash
# Установка и получение
HSET key field value
HSETNX key field value  # Если поле не существует
HGET key field
HMSET key field1 val1 field2 val2  # Множественная установка
HMGET key field [field ...]        # Множественное получение

# Дополнительно
HGETALL key     # Все поля и значения
HKEYS key       # Только поля
HVALS key       # Только значения
HLEN key        # Количество полей

# Удаление
HDEL key field [field ...]

# Числовые операции
HINCRBY key field increment
HINCRBYFLOAT key field increment
```

### Работа с сортированными множествами

```bash
# Добавление
ZADD key score member [score member ...]

# Получение
ZRANGE key start stop [WITHSCORES]    # По возрастанию
ZREVRANGE key start stop [WITHSCORES] # По убыванию
ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]

# Информация
ZCARD key           # Количество элементов
ZCOUNT key min max  # Количество в диапазоне
ZSCORE key member   # Очки элемента
ZRANK key member    # Ранг (позиция)
ZREVRANK key member # Ранг (по убыванию)

# Удаление
ZREM key member [member ...]
ZREMRANGEBYRANK key start stop
ZREMRANGEBYSCORE key min max
```

### Управление ключами

```bash
# Проверка и удаление
EXISTS key       # Проверить существование
DEL key [key ...]  # Удалить ключ(и)

# TTL
EXPIRE key seconds
PEXPIRE key ms
TTL key
PTTL key
PERSIST key

# Получение информации
KEYS pattern         # Найти ключи по шаблону (осторожно!)
SCAN cursor [MATCH pattern] [COUNT count]  # Итеративный обход

# Перемещение
RENAME key newkey
RENAMENX key newkey  # Только если newkey не существует
MOVE key db          # Переместить в другую БД (0-15)

# Тип данных
TYPE key
```

## Продвинутые возможности

### Публикация/Подписка (Pub/Sub)

Redis поддерживает модель Pub/Sub для рассылки сообщений между подписчиками.

```bash
# Подписка
SUBSCRIBE channel [channel ...]
PSUBSCRIBE pattern [pattern ...]  # По шаблону

# Публикация
PUBLISH channel message

# Отписка
UNSUBSCRIBE [channel ...]
PUNSUBSCRIBE [pattern ...]
```

**Особенности:**
- Сообщения не сохраняются (fire and forget)
- Подписчик получает только сообщения после подписки

### Транзакции (MULTI/EXEC)

Redis поддерживает базовую транзакционность через команды `MULTI` и `EXEC`.

```bash
MULTI        # Начало транзакции
COMMANDS     # Буферизация команд
EXEC         # Выполнение всех команд атомарно
DISCARD      # Отмена транзакции
```

**Особенности:**
- Атомарность: все команды выполняются последовательно без прерываний
- Нет отката (rollback): если одна команда失败, остальные выполняются

### Пакетирование команд (Pipeline)

Pipeline позволяет отправить несколько команд за один проход по сети.

**В коде (Python):**
```python
import redis
r = redis.Redis()
pipe = r.pipeline()
pipe.set('key1', 'value1')
pipe.set('key2', 'value2')
pipe.execute()  # Отправить все команды сразу
```

**Важно:** Pipeline ≠ Транзакция. Pipeline просто группирует команды для сетевой передачи.

### Lua-скриптирование

Redis позволяет выполнять Lua-скрипты на стороне сервера для атомарных операций.

```bash
# EVAL - выполнить скрипт
EVAL script numkeys key [key ...] arg [arg ...]

# EVALSHA - выполнить по хэшу скрипта (для кэширования)
EVALSHA sha1 numkeys key [key ...] arg [arg ...]
```

**Преимущества:**
- Атомарность: скрипт выполняется без прерываний
- Снижение количества запросов

### Битовые операции

Redis позволяет работать с отдельными битами строк.

```bash
# Установка и получение бита
SETBIT key offset value      # Установить бит
GETBIT key offset            # Получить бит

# Операции над битами
BITOP operation destkey key [key ...]  # Битовые операции (AND, OR, XOR, NOT)
BITCOUNT key [start end]          # Подсчет установленных битов
BITPOS key bit [start] [end]      # Позиция первого бита
```

### Geospatial индексы (GEO)

Redis поддерживает геопространственные индексы с помощью Sorted Sets.

```bash
# Добавление геоданных
GEOADD key longitude latitude member [longitude latitude member ...]

# Получение координат
GEOPOS key member [member ...]

# Расстояние между точками
GEODIST key member1 member2 [unit]  # unit: m, km, mi, ft

# Поиск ближайших
GEOSEARCH key [FROMMEMBER member] [FROMLONLAT longitude latitude] [BYRADIUS radius m|km|mi|ft] [BYBOX width height m|km|mi|ft] [ASC|DESC] [COUNT count] [WITHCOORD] [WITHDIST] [WITHHASH]
```

## Персистентность

### RDB (Redis Database)

RDB создает снапшот (snapshot) базы данных в определенные моменты времени.

**Конфигурация:**
```conf
save 900 1       # Сохранять каждые 900 сек, если изменился 1 ключ
save 300 10      # Сохранять каждые 300 сек, если изменилось 10 ключей
save 60 10000    # Сохранять каждые 60 сек, если изменилось 10000 ключей
dbfilename dump.rdb
dir /var/lib/redis
```

**Преимущества:**
- Компактный файл
- Быстрое восстановление

**Недостатки:**
- Возможна потеря данных между снапшотами

### AOF (Append Only File)

AOF записывает каждую операцию записи в лог-файл.

**Конфигурация:**
```conf
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec  # или always, no
```

**Режимы синхронизации:**
- `always` - каждый запрос синхронно на диск (медленно, но безопасно)
- `everysec` - каждую секунду (компромисс)
- `no` - делегирует системе (быстро, но рискованно)

## Производительность

### Многопоточность

- Redis 6.0+ поддерживает многопоточную обработку сетевых запросов
- Сама логика выполняется в одном потоке (single-threaded for commands)

**Настройка:**
```conf
# Количество потоков для I/O (по умолчанию 4)
io-threads 4
io-threads-do-reads yes
```

### Оптимизация памяти

**Упаковка хэшей:**
```conf
# Хэши меньшего размера кодируются более эффективно
hash-max-ziplist-entries 512
hash-max-ziplist-value 64

# Списки
list-max-ziplist-size -2

# Множества с числами
set-max-intset-entries 512

# Сортированные множества
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
```

**Политики удаления при нехватке памяти:**
```conf
maxmemory-policy allkeys-lru
# Варианты:
# noeviction - возвращать ошибку (по умолчанию)
# allkeys-lru - удалить наименее используемые ключи
# volatile-lru - удалить наименее используемые с TTL
# allkeys-random - случайные ключи
# volatile-random - случайные с TTL
# volatile-ttl - сначала те, у кого меньше TTL
```

## Безопасность

### Аутентификация

```bash
# В конфигурации
requirepass your_strong_password

# Или через команду
CONFIG SET requirepass "new_password"

# Подключение
redis-cli -a password
AUTH password
```

### ACL (Access Control Lists)

Redis 6.0+ поддерживает пользователей и ACL.

```bash
# Создать пользователя
ACL SETUSER alice on >password ~cache:* +get +set

# Просмотреть пользователей
ACL LIST

# Удалить пользователя
ACL DELUSER alice

# Проверить права
ACL WHOAMI
```

### TLS/SSL шифрование

```conf
# Включить TLS
tls-port 6379
port 0  # Отключить незашифрованный порт
tls-cert-file /path/to/redis.crt
tls-key-file /path/to/redis.key
tls-ca-cert-file /path/to/ca.crt
tls-auth-clients yes
```

## Кластеризация

### Redis Cluster

Redis Cluster обеспечивает децентрализованное хранение данных.

**Архитектура:**
- 16384 hash slots
- Каждый ключ соответствует slot: `CRC16(key) % 16384`
- Каждый slot может иметь реплику

**Требования:**
- Минимум 6 нод (3 мастер + 3 слейв)
- Redis 3.0+

**Создание кластера:**
```bash
# Конфигурация для каждой ноды (redis-7000.conf)
port 7000
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000

# Запуск нод
redis-server redis-7000.conf
# ... для всех 6 нод

# Создание кластера
redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 --cluster-replicas 1
```

### Репликация (Master-Slave)

Redis поддерживает асинхронную репликацию.

```bash
# На слейфе
SLAVEOF host port
# Или в конфиге
replicaof host port
```

## Использование в приложениях

### Python

```python
import redis

# Подключение
r = redis.Redis(host='localhost', port=6379, db=0, password='password')

# Работа со строками
r.set('key', 'value')
value = r.get('key')

# Работа с хэшами
r.hset('user:1', mapping={'name': 'John', 'age': 30})

# Работа с списками
r.rpush('queue', 'item1')
r.lpop('queue')

# TTL
r.setex('key', 3600, 'value')  # 1 час

# Pipeline
pipe = r.pipeline()
pipe.set('a', '1')
pipe.set('b', '2')
pipe.execute()
```

### Node.js

```javascript
const redis = require('redis');

const client = redis.createClient({
  socket: { host: 'localhost', port: 6379 },
  password: 'password'
});

await client.connect();

// Работа со строками
await client.set('key', 'value');
const value = await client.get('key');

// Работа с хэшами
await client.hSet('user:1', 'name', 'John');

// Pipeline
const multi = client.multi();
multi.set('a', '1');
multi.get('a');
await multi.exec();
```

### Кэширование (Cache-aside pattern)

```python
def get_user(user_id):
    cache_key = f"user:{user_id}"
    
    # Попытка из кэша
    user = r.get(cache_key)
    if user:
        return json.loads(user)
    
    # Из базы данных
    user = db.get_user(user_id)
    if user:
        r.setex(cache_key, 3600, json.dumps(user))  # 1 час
    return user
```

### Очереди сообщений

```python
# Простая очередь
def enqueue(task):
    r.rpush('tasks', json.dumps(task))

def dequeue():
    task = r.blpop('tasks', timeout=5)
    if task:
        return json.loads(task[1])
    return None
```

### Rate Limiting

```python
import time

def is_rate_limited(user_id, max_requests, window_seconds):
    key = f"ratelimit:{user_id}"
    now = time.time()
    
    pipe = r.pipeline()
    pipe.zadd(key, {now: now})
    pipe.zremrangebyscore(key, 0, now - window_seconds)
    pipe.zcard(key)
    pipe.expire(key, window_seconds)
    _, _, count, _ = pipe.execute()
    
    return count > max_requests
```

## Инструменты и мониторинг

### redis-cli

```bash
# Латентность
redis-cli --latency
redis-cli --latency-history

# Статистика
redis-cli --stat

# Мониторинг команд
redis-cli MONITOR

# Информация
redis-cli INFO
redis-cli INFO server
redis-cli INFO memory

# Slow log
redis-cli CONFIG SET slowlog-log-slower-than 10000
redis-cli SLOWLOG GET
```

### redis-benchmark

```bash
# Базовый тест
redis-benchmark -n 100000 -q

# Сpecific test
redis-benchmark -t ping,set,get -n 100000 -q

# С параллелизмом
redis-benchmark -p 6379 -n 100000 -c 50 -q
```

### RedisInsight

GUI для Redis от создания Redis:
- Просмотр данных
- Мониторинг производительности
- Анализ памяти
- Работа с ключами

### Prometheus + Redis Exporter

```yaml
# docker-compose.yml
version: '3'
services:
  redis:
    image: redis:7
    ports:
      - "6379:6379"
  
  redis-exporter:
    image: oliver006/redis_exporter
    ports:
      - "9121:9121"
    environment:
      - REDIS_ADDR=redis:6379
```

## Лучшие практики

### Нейминг ключей

```bash
# Использовать префиксы и разделители
user:1000:profile
user:1000:sessions
order:2024:01:001

# Избегать пробелов и спецсимволов
```

### Размер значений

- Избегать очень больших значений (>1MB)
- Для больших данных использовать файловое хранилище
- Разбивать большие списки/множества

### Pipeline для батчей

```python
# Плохо: 1000 отдельных запросов
for i in range(1000):
    r.set(f"key:{i}", i)

# Хорошо: один pipeline
pipe = r.pipeline()
for i in range(1000):
    pipe.set(f"key:{i}", i)
pipe.execute()
```

## Соседние технологии

### Redis vs Memcached

| Feature | Redis | Memcached |
|---------|-------|-----------|
| Типы данных | строка, список, множество, хэш, сортированное множество | только строка |
| Персистентность | RDB, AOF | нет |
| Репликация | да | нет |

### RedisJSON

Модуль для работы с JSON-документами:

```bash
JSON.SET key . '{"name": "John", "age": 30}'
JSON.GET key
JSON.NUMINCRBY key $.age 1
```

### RedisSearch

Модуль для полнотекстового поиска:

```bash
FT.CREATE idx ON HASH PREFIX 1 "doc:" SCHEMA title TEXT body TEXT
FT.SEARCH idx "hello world"
```

### RedisTimeSeries

Модуль для временных рядов:

```bash
TS.CREATE temperature
TS.ADD temperature * 25.5
TS.RANGE temperature 1600000000 1600000100
```

## Полезные ссылки

- [Официальный сайт Redis](https://redis.io/)
- [Документация](https://redis.io/documentation)
- [Redis Commands](https://redis.io/commands/)
- [Redis GitHub](https://github.com/redis/redis)
- [RedisInsight](https://redis.com/redis-enterprise/redis-insight/)
