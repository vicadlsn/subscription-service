# Сервис управления подписками пользователей

## Описание

Демонстрационный HTTP-сервис для управления онлайн-подписками пользователей. Реализованы CRUDL-операции и расчет суммарной стоимости подписок.

Функционал:

-   Создание, чтение, обновление, удаление и получение списка подписок
-   Фильтрация подписок по ID пользователя и названию сервиса
-   Подсчет суммарной стоимости подписок за выбранный период с фильтрацией по
-   Поддержка опциональной даты окончания подписки

## Запуск проекта

### Предварительные требования

-   Docker
-   Docker Compose

### Шаги для запуска

1. Создайте файл `.env` в корневой директории проекта со следующими переменными:

```env
DB_USERNAME=db_username
DB_PASSWORD=db_password
DB_NAME=db_name
DB_SCHEME=public
DB_PORT=5432
DB_MAX_ATTEMPTS=5
DB_RETRY_DELAY=5
DB_CONNECTION_TIMEOUT=10
SERVER_PORT=8081
```

2. Запустите сервис:

```bash
docker-compose up --build
```

3. Для применения миграций базы данных выполните:

```bash
docker-compose run --rm migrate
```

4. Просмотр логов приложения

```bash
docker-compose logs -f app
```

5. Остановка сервиса

```bash
docker-compose down
```

6. Полная очистка (включая удаление томов)

```bash
docker-compose down -v
```

## Команды разработки (Makefile)

-   `make run` - Запуск приложения локально
-   `make build` - Сборка бинарного файла
-   `make docker-build` - Сборка Docker-образа
-   `make docker-run` - Запуск контейнера
-   `make compose-up` - Запуск всех сервисов через Docker Compose
-   `make compose-down` - Остановка всех сервисов
-   `make compose-migrate` - Применение миграций базы данных
-   `make docker-clean` - Полная очистка Docker-окружения
-   `make logs-app` - Просмотр логов

## API Документация

### Swagger

Swagger документация доступна по адресу `http://localhost:8081/swagger/index.html` после запуска сервиса.

### API Endpoints

#### Создание подписки

-   `POST /api/subscriptions`
-   Тело запроса:

```json
{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025" // опционально
}
```

#### Получение списка подписок

-   `GET /api/v1/subscriptions`

-   Параметры:

    -   `user_id`? - UUID пользователя
    -   `service_name`? - название сервиса
    -   `start_date`? - начало периода (формат: MM-YYYY)
    -   `end_date`? - конец периода (формат: MM-YYYY)
    -   `limit`? - количество записей (по умолчанию: 50)
    -   `offset`? - смещение (по умолчанию: 0)
    -   `sort_by`? - поле для сортировки (по умолчанию: created_at)
    -   `sort_order`? - порядок сортировки (asc/desc, по умолчанию: desc)

#### Получение подписки по ID

-   `GET /api/subscriptions/{id}`

#### Обновление подписки

-   `PUT /api/subscriptions/{id}`
-   Тело запроса:

```json
{
    "service_name": "Yandex Plus",
    "price": 400,
    "start_date": "07-2025",
    "end_date": "12-2025" // опционально
}
```

#### Удаление подписки

-   `DELETE /api/subscriptions/{id}`

#### Подсчет суммарной стоимости подписок

-   `GET /api/subscriptions/total`
-   Параметры запроса:
    -   `start_date` - начало периода (формат: MM-YYYY)
    -   `end_date` - конец периода (формат: MM-YYYY)
    -   `user_id`? - UUID пользователя
    -   `service_name`? - название сервиса
