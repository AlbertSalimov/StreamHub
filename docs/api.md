# StreamHub REST API

## Общая информация

Базовый URL локального сервера:

`http://localhost:8080`

Формат тела запросов и ответов: JSON. Для запросов с телом используйте заголовок `Content-Type: application/json`.

## 1. Health check

### `GET /healthz`

Проверяет, что HTTP-приложение отвечает.

```bash
curl -i http://localhost:8080/healthz
```

### `GET /readyz`

Проверяет готовность приложения, включая соединение с PostgreSQL.

```bash
curl -i http://localhost:8080/readyz
```

## 2. Создание контента

### `POST /api/v1/content/`

Пример запроса:

```bash
curl -X POST http://localhost:8080/api/v1/content/ \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My first stream",
    "description": "An example video",
    "genre": "tech",
    "release_year": 2025,
    "is_published": true
  }'
```

Пример полей:

| Поле           | Тип     | Описание               |
| -------------- | ------- | ---------------------- |
| `title`        | string  | Название контента      |
| `description`  | string  | Описание               |
| `genre`        | string  | Жанр                   |
| `release_year` | integer | Год выпуска            |
| `is_published` | boolean | Опубликован ли контент |

При успешном создании API возвращает созданный объект с идентификатором UUID.

## 3. Получение каталога

### `GET /api/v1/content/`

Получает опубликованный контент с возможностью поиска, фильтрации, сортировки и пагинации.

Пример:

```bash
curl "http://localhost:8080/api/v1/content/?search=stream&genre=tech&sort=year&order=desc&limit=10&offset=0"
```

Параметры:

| Параметр | Назначение                                        |
| -------- | ------------------------------------------------- |
| `search` | Поиск по тексту                                   |
| `genre`  | Фильтр по жанру                                   |
| `year`   | Фильтр по году выпуска                            |
| `sort`   | Поле сортировки: `title`, `year` или `date`       |
| `order`  | Направление: `asc` или `desc`                     |
| `limit`  | Количество записей; по умолчанию 20, максимум 100 |
| `offset` | Число пропускаемых записей; по умолчанию 0        |

Пример структуры ответа каталога:

```json
{
  "items": [],
  "limit": 10,
  "offset": 0
}
```

`items` содержит найденные записи. Пустой массив означает, что по заданным условиям нет результатов.

## 4. Получение одной записи

### `GET /api/v1/content/{id}`

Вместо `{id}` укажите UUID записи.

```bash
curl -i http://localhost:8080/api/v1/content/00000000-0000-4000-8000-000000000001
```

Если запись не найдена, API должен вернуть HTTP-статус 404 Not Found.

## 5. Обновление контента

### `PATCH /api/v1/content/{id}`

```bash
curl -X PATCH http://localhost:8080/api/v1/content/00000000-0000-4000-8000-000000000001 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated stream",
    "description": "Updated description",
    "genre": "education",
    "release_year": 2024,
    "is_published": true
  }'
```

Передайте UUID существующей записи.

## 6. Удаление контента

### `DELETE /api/v1/content/{id}`

```bash
curl -i -X DELETE \
  http://localhost:8080/api/v1/content/00000000-0000-4000-8000-000000000001
```

При успешном удалении запись удаляется из PostgreSQL. Приложение также пытается инвалидировать кеш каталога и опубликовать событие в NATS.
