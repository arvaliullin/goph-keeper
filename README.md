# GophKeeper

Клиент-серверная система для безопасного хранения приватных данных: логинов/паролей, текстовых заметок, бинарных файлов и данных банковских карт.

## Архитектура

```
keeper (CLI) --HTTP/JSON--> keeperd (сервер) --> PostgreSQL (секреты)
                                             --> MinIO (бинарные файлы)
```

- **Сервер** (`cmd/keeperd`) - REST API на Go (Chi), JWT-аутентификация, AES-256-GCM шифрование данных в БД.
- **Клиент** (`cmd/keeper`) - CLI-приложение на Cobra, поддержка Windows/Linux/macOS.

## Типы хранимых данных

| Тип | Описание |
|-----|----------|
| `login` | Пара логин/пароль |
| `text` | Произвольные текстовые данные |
| `card` | Данные банковской карты |
| `binary` | Произвольные бинарные файлы |

Для каждого типа поддерживается произвольная текстовая метаинформация (`--meta`).

## Быстрый старт

### Запуск через Docker Compose

```bash
make up
```

Поднимает PostgreSQL, MinIO и сервер. API доступен на `http://localhost:8080`.

### Сборка

```bash
make build-server       # сервер
make build-client       # клиент (текущая платформа)
make build-client-all   # клиент для Linux, macOS, Windows
```

### Swagger

После запуска сервера документация доступна по адресу `http://localhost:8080/swagger/`.

## Переменные окружения сервера

| Переменная | Описание | По умолчанию |
|------------|----------|-------------|
| `ADDRESS` | Адрес HTTP-сервера | `localhost:8080` |
| `DATABASE_URI` | URI подключения к PostgreSQL | `postgres://keeper:keeperpassword@localhost:5432/gophkeeper?sslmode=disable` |
| `MINIO_ENDPOINT` | Адрес MinIO | `localhost:9000` |
| `MINIO_ACCESS_KEY` | Ключ доступа MinIO | `minioadmin` |
| `MINIO_SECRET_KEY` | Секретный ключ MinIO | `minioadmin` |
| `MINIO_SECURE` | Использовать TLS для MinIO | `false` |
| `JWT_SECRET` | Секрет для подписи JWT (обязательный) | - |
| `ENCRYPTION_KEY` | Ключ шифрования AES-256 (32 байта, обязательный) | - |
| `MAX_BINARY_SIZE_BYTES` | Максимальный размер бинарного файла | `10485760` (10 MB) |

## Команды CLI

### Аутентификация

```bash
keeper register -l <логин> -p <пароль>   # регистрация
keeper login -l <логин> -p <пароль>      # вход
keeper logout                             # выход (удаление токена)
```

Если `-p` не указан, пароль запрашивается интерактивно.

### Управление секретами

```bash
keeper add -t <тип> -d <данные> [-m <метаданные>]      # создать секрет
keeper add-binary -f <путь_к_файлу> [-m <метаданные>]  # создать бинарный секрет
keeper update <id> [-d <данные>] [-m <метаданные>]     # обновить
keeper list                                             # список
keeper get <id> [-o <файл>]                             # получить (для binary -o сохраняет файл)
keeper download-binary <id> -o <файл>                   # скачать бинарный файл
keeper delete <id>                                      # удалить
```

Допустимые значения `-t` (тип секрета): `login`, `text`, `card`, `binary`.
Для бинарных данных рекомендуется использовать `add-binary` вместо `add -t binary`.

### Синхронизация

```bash
keeper sync   # получить изменения с сервера
```

### Прочее

```bash
keeper version              # версия и дата сборки
keeper --server <url> ...   # указать сервер для команды
```

Конфигурация клиента хранится в `<домашняя_директория>/.gophkeeper/config.json`
(определяется через `os.UserHomeDir()`: `$HOME` на Linux/macOS, `%USERPROFILE%` на Windows).

## REST API

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/register` | Регистрация |
| POST | `/api/v1/login` | Аутентификация |
| POST | `/api/v1/secrets` | Создать секрет |
| GET | `/api/v1/secrets` | Список (+ `?updated_after=` для синхронизации) |
| GET | `/api/v1/secrets/{id}` | Получить секрет |
| PUT | `/api/v1/secrets/{id}` | Обновить секрет |
| DELETE | `/api/v1/secrets/{id}` | Удалить секрет |
| POST | `/api/v1/binary/{id}` | Загрузить бинарный файл |
| GET | `/api/v1/binary/{id}` | Скачать бинарный файл |
| DELETE | `/api/v1/binary/{id}` | Удалить бинарный файл |

Защищенные эндпоинты требуют заголовок `Authorization: Bearer <JWT>`.

## Тестирование

```bash
make test                # юнит-тесты с проверкой покрытия (порог 70%)
make test-integration    # интеграционные тесты (требуют запущенную инфраструктуру)
make lint                # линтер (golangci-lint)
```

## Makefile

```bash
make help   # список всех доступных команд
```
