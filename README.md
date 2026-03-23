# 🥬 VeggieShops.kz — система онлайн-доставки овощей и фруктов

## Быстрый старт

### Вариант 1: Docker (рекомендуется)

```bash
# Запуск всего стека
docker compose up -d

# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# Health: curl http://localhost:8080/api/v1/health
```

### Вариант 2: Локальная разработка

**Требования:** Go 1.21+, Node 18+, PostgreSQL 14+

```bash
# 1. База данных
docker compose up -d postgres

# 2. Backend
cd backend && cp .env.example .env
go run cmd/server/main.go

# 3. Frontend (новый терминал)
cd frontend && npm install && npm run dev
```

### Создание первого магазина

После запуска БД и backend:

```bash
# Регистрация админа
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"+77001234567","password":"password123","first_name":"Admin"}'

# Создание магазина (нужен JWT - залогиньтесь и возьмите токен)
# Или через SQL:
# INSERT INTO stores (name, address, latitude, longitude, delivery_radius_km, min_order_amount)
# VALUES ('Зелень & Здоровье', 'ул. Достык 10', 51.1694, 71.4491, 3, 2500);
```

### Документация
- `files/TZ_VEGETABLE_SHOP.md` - полное ТЗ
- `files/DELIVERY_SYSTEM_LOGIC.md` - логика доставки

## Структура проекта

```
backend/          # Go + Gin + GORM + PostgreSQL
frontend/         # React + TypeScript + Vite + Tailwind
files/            # Документация (ТЗ, инструкции)
docker-compose    # PostgreSQL + Backend + Frontend
```
