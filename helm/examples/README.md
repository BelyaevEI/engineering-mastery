# Пример использования Helm с простым приложением

## Структура проекта

```
simple-app/
├── Chart.yaml          # Метаданные чарта
├── values.yaml         # Значения по умолчанию
└── templates/          # Шаблоны Kubernetes ресурсов
    ├── _helpers.tpl    # Вспомогательные шаблоны
    ├── deployment.yaml # Deployment
    ├── service.yaml    # Service
    ├── ingress.yaml    # Ingress (опционально)
    └── configmap.yaml  # ConfigMap
```

---

## Установка чарта

### Базовая установка (с значениями по умолчанию)

```bash
# Установка чарта
helm install my-app ./simple-app

# Проверка статуса релиза
helm status my-app

# Просмотр установленных ресурсов
kubectl get all -l app.kubernetes.io/instance=my-app
```

### Установка с переопределением значений

```bash
# Через командную строку
helm install my-app ./simple-app \
  --set replicaCount=3 \
  --set image.tag=1.25 \
  --set service.type=LoadBalancer

# Через values файл
helm install my-app -f custom-values.yaml ./simple-app
```

### Установка с включенным Ingress

Создайте `values-ingress.yaml`:

```yaml
ingress:
  enabled: true
  hosts:
    - host: my-app.example.com
      paths:
        - path: /
          pathType: Prefix
```

Установка:

```bash
helm install my-app -f values-ingress.yaml ./simple-app
```

---

## Просмотр информации

```bash
# Информация о релизе
helm status my-app

# Значения релиза
helm get values my-app

# Манифесты релиза
helm get manifest my-app

# История релиза
helm history my-app
```

---

## Обновление релиза

```bash
# Обновление количества реплик
helm upgrade my-app ./simple-app --set replicaCount=5

# Обновление версии образа
helm upgrade my-app ./simple-app --set image.tag=1.25

# Обновление с values файлом
helm upgrade my-app -f values-prod.yaml ./simple-app
```

---

## Откат к предыдущей версии

```bash
# Просмотр истории
helm history my-app
# Пример вывода:
# REVISION  UPDATED                  STATUS     CHART         APP VERSION DESCRIPTION
# 1         Mon Aug  8 15:00:00 2024   deployed   simple-app-1.0.0 1.0.0       Initial install
# 2         Mon Aug  8 15:05:00 2024   deployed   simple-app-1.0.0 1.0.0       Upgraded with 3 replicas
# 3         Mon Aug  8 15:10:00 2024   deployed   simple-app-1.0.0 1.0.0       Upgraded with 5 replicas

# Откат к предыдущей версии
helm rollback my-app 2

# Откат к первой версии
helm rollback my-app 1
```

---

## Удаление релиза

```bash
# Удаление релиза
helm uninstall my-app

# Удаление с сохранением истории
helm uninstall my-app --keep-history
```

---

## Тестирование чарта перед установкой

```bash
# Проверка синтаксиса
helm lint ./simple-app

# Dry-run (предпросмотр создаваемых ресурсов)
helm install my-test ./simple-app --dry-run --debug

# Генерация манифестов без установки
helm template my-test ./simple-app
```

---

## Примеры значений

### Продуктивная конфигурация (values-prod.yaml)

```yaml
replicaCount: 3

image:
  repository: myregistry/myapp
  tag: "1.0.0"
  pullPolicy: Always

service:
  type: ClusterIP
  port: 80
  targetPort: 80

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: app.example.com
      paths:
        - path: /
          pathType: Prefix

app:
  name: myapp-prod
  logLevel: warn
```

### Dev конфигурация (values-dev.yaml)

```yaml
replicaCount: 1

image:
  repository: myregistry/myapp
  tag: "dev"
  pullPolicy: Always

service:
  type: ClusterIP
  port: 80
  targetPort: 80

ingress:
  enabled: false

app:
  name: myapp-dev
  logLevel: debug
```

---

## Управление несколькими релизами

```bash
# Установка разных версий в разные namespace
helm install my-app-dev ./simple-app -n development --create-namespace
helm install my-app-prod ./simple-app -n production --create-namespace

# Список всех релизов
helm list -A

# Удаление из конкретного namespace
helm uninstall my-app-dev -n development
```

---

## Common Commands

| Команда | Описание |
|---------|----------|
| `helm install <name> <chart>` | Установка релиза |
| `helm upgrade <name> <chart>` | Обновление релиза |
| `helm rollback <name> <revision>` | Откат релиза |
| `helm uninstall <name>` | Удаление релиза |
| `helm list` | Список релизов |
| `helm status <name>` | Статус релиза |
| `helm lint <chart>` | Проверка чарта |
| `helm template <name> <chart>` | Предпросмотр манифестов |
| `helm get values <name>` | Просмотр значений |
| `helm history <name>` | История релиза |
