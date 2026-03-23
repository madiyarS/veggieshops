# 🚀 Как запустить VeggieShops.kz

> ⚠️ **Если Docker не запущен:** откройте приложение Docker Desktop (Mac/Windows) и дождитесь готовности. Либо используйте Вариант 2 без Docker для БД — потребуется локальный PostgreSQL.

## Вариант 1: Docker (рекомендуется)

Один раз запустить всё:

```bash
cd /Users/madi/ovoshnoy_bro
docker compose up -d
```

Сервисы:
- **Frontend:** http://localhost:3001
- **Backend API:** http://localhost:8081
- **PostgreSQL:** localhost:5433 (user: postgres, pass: password, db: veggies_shop)

> Если порты 3000, 8080, 5432 заняты — в docker-compose.yml используются 3001, 8081, 5433.

Проверка:
```bash
curl http://localhost:8081/api/v1/health
# Ответ: {"status":"ok",...}
```

Остановка:
```bash
docker compose down
```

---

## Вариант 2: Локальная разработка

### Требования
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+ (или Docker только для БД)

### Шаг 1: База данных

```bash
cd /Users/madi/ovoshnoy_bro
docker compose up -d postgres
```

Миграции выполнятся автоматически при первом запуске PostgreSQL.

### Шаг 2: Backend

```bash
cd backend
cp .env.example .env
# При необходимости отредактируйте .env
go run cmd/server/main.go
```

Сервер: http://localhost:8080

### Шаг 3: Frontend

В **новом терминале**:

```bash
cd frontend
npm install
npm run dev
```

Откроется: http://localhost:5173 (Vite dev server)

---

## 🧪 Тестовые данные и админ-панель

После `docker compose up -d` выполните seed для примеров:

```bash
docker exec -i veggieshopskz-postgres psql -U postgres -d veggies_shop < backend/migrations/003_seed_examples.sql
```

Это создаст:
- **Админ:** +77000000000 / **admin123**
- **Магазин:** Зелень & Здоровье (ул. Достык 10, Астана)
- **Районы:** Астана-Центр, Есиль, Сарыарка
- **Временные окна:** 09-11, 11-13, 13-15 (все дни)
- **Товары:** Помидоры, Огурцы, Картофель, Морковь, Яблоки, Бананы, Укроп, Петрушка

### Вход в админ-панель

1. Откройте http://localhost:3001 (или 3000)
2. Нажмите **Админ** в шапке
3. Логин: **+77000000000** | Пароль: **admin123**
4. После входа — список магазинов

### Полный цикл тестирования

1. **Главная** — выбрать магазин «Зелень & Здоровье»
2. **Каталог** — добавить товары в корзину
3. **Корзина** — проверить сумму
4. **Оформление** — выбрать район, адрес, время, оплату
5. **Отслеживание** — ввести номер заказа и телефон

---

## Создание первого магазина (вручную)

Backend и БД должны быть запущены.

### 1. Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"+77001234567","password":"password123","first_name":"Admin","last_name":"User"}'
```

### 2. Сделать пользователя админом (через psql)

```bash
# Подключиться к БД
docker exec -it veggieshopskz-postgres psql -U postgres -d veggies_shop

# В psql:
UPDATE users SET role = 'admin' WHERE phone = '+77001234567';
\q
```

### 3. Получить токен (логин)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"+77001234567","password":"password123"}'
```

Скопируйте `access_token` из ответа.

### 4. Создать магазин

```bash
curl -X POST http://localhost:8080/api/v1/admin/stores \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ВАШ_ACCESS_TOKEN" \
  -d '{
    "name": "Зелень & Здоровье",
    "address": "ул. Достык 10, Астана",
    "latitude": 51.1694,
    "longitude": 71.4491,
    "phone": "+77071234567",
    "delivery_radius_km": 3,
    "min_order_amount": 2500
  }'
```

### 5. Добавить район и слоты (через psql или API)

Через SQL:
```sql
-- Получить store_id из таблицы stores
INSERT INTO districts (store_id, name, distance_km, delivery_fee_regular, delivery_fee_express, is_active)
VALUES ('UUID_МАГАЗИНА', 'Астана-Центр', 1.5, 500, 1000, true);

INSERT INTO delivery_time_slots (store_id, day_of_week, start_time, end_time, max_orders, is_active)
VALUES ('UUID_МАГАЗИНА', 1, '09:00', '11:00', 10, true),
       ('UUID_МАГАЗИНА', 1, '11:00', '13:00', 10, true);
```

### 6. Добавить категории и товары

Категории уже созданы seed-миграцией (002_seed_data.sql). Товары добавляются через Admin API или SQL.

---

## Быстрый запуск без Docker (только проверка)

Если Docker не установлен или не запущен:

```bash
# Backend (health check будет работать, остальное — без БД)
cd backend
go run cmd/server/main.go

# Frontend (в новом терминале)
cd frontend
npm run dev
```

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Магазины/товары/заказы не работают без PostgreSQL.

---

## Полезные команды

| Действие | Команда |
|----------|---------|
| Логи backend | `docker compose logs -f backend` |
| Логи postgres | `docker compose logs -f postgres` |
| Перезапуск | `docker compose restart` |
| Очистить volumes | `docker compose down -v` |

---

## Переменные окружения (.env)

### Backend (backend/.env)

```
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=veggies_shop
JWT_SECRET=your-secret-key
```

### Frontend (frontend/.env)

```
VITE_API_URL=http://localhost:8080/api/v1
```

При `npm run dev` можно использовать proxy (см. vite.config.ts) — тогда API URL: `/api/v1`.

---

## Каталог и остатки (API)

- **GET** `/api/v1/products?store_id=UUID` — список витрины. Параметры: `category_id`, `q` (поиск), `in_stock_only=true`, `sort=name|price_asc|price_desc|expiry_asc`. В ответе у товара: `stock_quantity` (доступно), `nearest_batch_expires_at`, `catalog_low_stock`.
- **GET** `/api/v1/products/availability?store_id=UUID&product_ids=id1,id2` — доступный остаток по списку товаров (до 50 UUID), для веса — **граммы**.

## Склад (админ, JWT)

- **POST** `/api/v1/admin/stores/{storeId}/stock/receive-simple` — тело `{ "product_id", "quantity", "note?" }`.
- **GET** `/api/v1/admin/stores/{storeId}/stock/moves-journal?limit=` — журнал с названием товара.
- **POST** `/api/v1/admin/stores/{storeId}/stock/set-actual` — тело `{ "product_id", "actual", "note?" }`.

Подробнее см. `backend/openapi/openapi.yaml`.

## Ошибки API

Ответы с ошибкой: `{ "success": false, "error": "текст", "code": "КОД" }`. Примеры кодов: `INSUFFICIENT_STOCK`, `MIN_ORDER_AMOUNT`, `PRODUCT_UNAVAILABLE`, `VALIDATION_ERROR`, `UNAUTHORIZED`. Поле `code` удобно передавать в поддержку.
