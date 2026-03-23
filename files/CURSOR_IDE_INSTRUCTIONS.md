# 🚀 ИНСТРУКЦИИ ДЛЯ CURSOR IDE
## Как использовать ТЗ и начать разработку

---

## 1. ПОДГОТОВКА К РАЗРАБОТКЕ

### 1.1 Предварительные требования

Убедитесь, что установлено:
```bash
# Проверить версии
go version        # Go 1.21+ (https://go.dev/dl)
node --version    # Node 18+ (https://nodejs.org)
npm --version     # npm 9+
docker --version  # Docker 20+
postgresql --version # PostgreSQL 14+ (для локальной разработки)

# Если чего-то нет - установите
# macOS:
brew install go node postgresql

# Ubuntu/Debian:
sudo apt update && sudo apt install golang-go nodejs postgresql postgresql-contrib

# Windows:
# Скачайте инсталляторы с официальных сайтов
```

### 1.2 Настройка Cursor IDE

1. **Откройте Cursor** (https://www.cursor.sh/)
2. **Создайте новый проект:**
   - File → Open Folder → выберите `/path/to/veggies-shop`
3. **Установите расширения:**
   - Go extension (для разработки backend)
   - ESLint & Prettier (для frontend)
   - PostgreSQL (для БД)

### 1.3 Структура проекта

```bash
# Создайте базовую структуру
mkdir -p veggies-shop
cd veggies-shop

# Backend
mkdir -p backend/{cmd,internal/{api,models,repositories,services,config,utils},migrations}

# Frontend
npx create-vite@latest frontend --template react-ts

# Docker & Config
touch docker-compose.yml .env.example .gitignore README.md
```

---

## 2. ПОШАГОВЫЙ ПЛАН РАЗРАБОТКИ

### 2.1 ШАГ 1: Backend Setup (День 1-2)

**Prompt для Cursor (скопируйте и отправьте):**

```
Помоги мне создать базовый Golang проект для интернет-магазина овощей.

Требования:
- Framework: Gin или Echo для REST API
- Database: PostgreSQL с GORM ORM
- Structure: cmd, internal (api, models, repositories, services, config, utils), migrations
- Authentication: JWT
- Port: 8080

Нужно создать:
1. go.mod с необходимыми зависимостями
2. main.go с инициализацией сервера
3. config/config.go для переменных окружения
4. Пример структуры для моделей (User, Store, Product)
5. docker-compose.yml для PostgreSQL
6. .env.example с примерами переменных

Структура должна быть готова к масштабированию и следовать best practices Go
```

**Что будет создано:**
- ✅ Структура проекта Go
- ✅ docker-compose.yml с PostgreSQL
- ✅ go.mod с зависимостями
- ✅ config.go для конфигурации
- ✅ main.go с базовым сервером

---

### 2.2 ШАГ 2: Database Design (День 3-4)

**Prompt для Cursor:**

```
Создай SQL миграции для системы онлайн-доставки овощей на PostgreSQL.

Таблицы (согласно ТЗ, раздел 3.1):
1. users - пользователи (клиенты, администраторы, курьеры)
2. stores - магазины
3. districts - районы доставки
4. categories - категории товаров (Овощи, Фрукты, Зелень и т.д.)
5. products - товары (с фото, весом, ценой, остатками)
6. delivery_time_slots - временные окна доставки (2-часовые интервалы)
7. orders - заказы (со статусами: pending, confirmed, preparing, in_delivery, delivered)
8. order_items - товары в заказе
9. couriers - информация о курьерах
10. notifications - очередь уведомлений в WhatsApp

Требования:
- Все с UUID primary keys
- Timestamps (created_at, updated_at)
- Правильные Foreign Keys
- Индексы для часто запрашиваемых полей
- Enum типы для статусов

Создай файл migrations/001_create_tables.sql с полной структурой.
Файлы сохрани в backend/migrations/
```

**Что будет создано:**
- ✅ SQL миграции для всех таблиц
- ✅ Правильные типы данных
- ✅ Индексы и Foreign Keys
- ✅ Комментарии к полям

---

### 2.3 ШАГ 3: Models & Repositories (День 4-5)

**Prompt для Cursor:**

```
Создай Go модели и GORM структуры для таблиц БД.

Модели нужны для:
1. User (с ролями: customer, admin, manager, courier)
2. Store (с координатами, радиусом доставки)
3. District (районы доставки магазина)
4. Category (категории товаров)
5. Product (товары - с весом, остатками, фото)
6. DeliveryTimeSlot (временные окна)
7. Order (заказы с полным циклом статусов)
8. OrderItem (товары в заказе)
9. Courier (информация о курьерах)
10. Notification (для WhatsApp уведомлений)

Также создай Repository слой (GORM):
- UserRepository (Create, GetByID, GetByPhone, Update)
- StoreRepository (GetAll, GetByID, Create, Update)
- ProductRepository (GetByStoreID, GetByCategory, Create, Update, Delete)
- OrderRepository (Create, GetByID, GetByNumber, UpdateStatus)
- И т.д.

Структура:
- internal/models/models.go - все struct'ы
- internal/repositories/user_repo.go
- internal/repositories/store_repo.go
- ... (один файл на repository)

Используй GORM и следуй best practices.
```

**Что будет создано:**
- ✅ GORM модели для всех таблиц
- ✅ Repository interfaces
- ✅ Реализация CRUD операций
- ✅ Query методы с фильтрацией

---

### 2.4 ШАГ 4: API Handlers (День 5-6)

**Prompt для Cursor:**

```
Создай REST API handlers для основных функций.

Endpoints (согласно ТЗ раздел 7):

PUBLIC API:
- POST /api/v1/auth/register - регистрация
- POST /api/v1/auth/login - вход
- GET /api/v1/products - список товаров
- GET /api/v1/products/:id - детали товара
- GET /api/v1/categories - список категорий
- GET /api/v1/stores - список магазинов
- GET /api/v1/stores/:id/districts - районы магазина
- POST /api/v1/orders/check-delivery - проверка доставки
- POST /api/v1/orders - создать заказ
- GET /api/v1/orders/track - отследить заказ

ADMIN API (с JWT middleware):
- GET /api/v1/admin/stores - все магазины
- POST /api/v1/admin/stores - создать магазин
- PATCH /api/v1/admin/stores/:id - редактировать магазин
- GET /api/v1/admin/stores/:storeId/products - товары магазина
- POST /api/v1/admin/stores/:storeId/products - создать товар
- GET /api/v1/admin/orders - все заказы
- PATCH /api/v1/admin/orders/:id/status - изменить статус заказа
- И т.д.

Требования:
- Используй Gin framework
- JWT middleware для защищенных роутов
- Правильные HTTP статусы (200, 201, 400, 401, 404, 500)
- JSON responses с структурированными ошибками
- Валидация input данных
- CORS для фронтенда

Структура:
- internal/api/handlers/*.go
- internal/api/middleware/auth.go
- internal/api/routes.go

Создай сначала базовые handlers, потом усложни.
```

**Что будет создано:**
- ✅ HTTP handlers для всех endpoints
- ✅ JWT middleware
- ✅ Валидация данных
- ✅ Error handling
- ✅ CORS middleware

---

### 2.5 ШАГ 5: Services Layer (День 6)

**Prompt для Cursor:**

```
Создай сервис слой (Business Logic) между handlers и repositories.

Сервисы:
1. AuthService - регистрация, логин, JWT токены, refresh токены
2. ProductService - логика работы с товарами (фильтрация, поиск, остатки)
3. OrderService - создание заказа, проверка доступности, статусы
4. DeliveryService - расчет стоимости доставки, проверка радиуса (ВАЖНО!)
5. StoreService - управление магазинами, районами, временными окнами
6. NotificationService - отправка WhatsApp уведомлений
7. CourierService - управление курьерами

САМОЕ ВАЖНОЕ:
- DeliveryService должен содержать функцию:
  func (s *DeliveryService) CheckDeliveryAvailability(
      storeID string, 
      customerLat, customerLon float64
  ) (*DeliveryAvailability, error)
  
  Она должна:
  1. Получить координаты магазина
  2. Рассчитать расстояние (haversine formula)
  3. Проверить: distance <= store.delivery_radius_km
  4. Если да - вернуть доступные районы и стоимость доставки
  5. Если нет - вернуть ошибку

Требования:
- Все сервисы имеют interfaces
- Dependency injection через конструкторы
- Обработка ошибок
- Логирование (используй log/slog из Go 1.21)
- Unit тестируемость

Структура:
- internal/services/*.go

Используй best practices для слоистой архитектуры.
```

**Что будет создано:**
- ✅ Business logic слой
- ✅ DeliveryService с расчетом расстояния
- ✅ Все необходимые сервисы
- ✅ Dependency Injection

---

### 2.6 ШАГ 6: Frontend Setup (День 7)

**Prompt для Cursor:**

```
Помоги настроить React TypeScript проект для онлайн-магазина овощей.

Setup:
- Framework: Vite (уже должен быть из create-vite)
- UI Library: Tailwind CSS или Material-UI v5
- State Management: Redux Toolkit или Zustand
- HTTP Client: Axios
- Routing: React Router v6
- Form Validation: React Hook Form + Zod

Структура проекта:
frontend/src/
├── pages/
│   ├── Home.tsx
│   ├── Catalog.tsx
│   ├── Cart.tsx
│   ├── Checkout.tsx
│   ├── TrackOrder.tsx
│   └── Admin/
│       ├── AdminDashboard.tsx
│       ├── StoresPage.tsx
│       ├── ProductsPage.tsx
│       ├── OrdersPage.tsx
│       └── AnalyticsPage.tsx
├── components/
│   ├── ProductCard.tsx
│   ├── CartItem.tsx
│   ├── Header.tsx
│   ├── Footer.tsx
│   └── ...
├── hooks/
│   ├── useAuth.ts
│   ├── useCart.ts
│   └── ...
├── services/
│   └── api.ts
├── store/
│   ├── slices/
│   │   ├── authSlice.ts
│   │   ├── cartSlice.ts
│   │   └── ...
│   └── index.ts
├── App.tsx
└── main.tsx

Требования:
- TypeScript strict mode
- Правильная типизация всех компонентов
- Environment переменные для API URL
- Dark mode support (опционально)
- Responsive design для мобильных

Установи все зависимости и сделай первые компоненты.
```

**Что будет создано:**
- ✅ Vite конфигурация
- ✅ Tailwind CSS или MUI настройка
- ✅ Redux/Zustand setup
- ✅ Routing конфигурация
- ✅ Base структура компонентов

---

### 2.7 ШАГ 7: Frontend Pages & Components (День 8-10)

**Prompt для Cursor:**

```
Создай страницы и компоненты для клиентской части:

ГЛАВНАЯ СТРАНИЦА (Home.tsx):
- Список доступных магазинов (по локации)
- Поисковая строка
- Популярные товары этой недели
- Акции/спецпредложения

КАТАЛОГ (Catalog.tsx):
- Список товаров с фото
- Фильтры по категориям
- Сортировка (цена, популярность)
- Быстрое добавление в корзину
- Modal с полной информацией о товаре

КОРЗИНА (Cart.tsx):
- Список товаров в корзине
- Изменение количества
- Удаление товара
- Показать минимальную сумму заказа
- Кнопка "Оформить заказ"

ОФОРМЛЕНИЕ ЗАКАЗА (Checkout.tsx):
- Этап 1: Выбор типа доставки (Regular/Express)
- Этап 2: Выбор района доставки
- Этап 3: Ввод адреса доставки
- Этап 4: Выбор временного окна доставки
- Этап 5: Выбор способа оплаты
- Расчет итоговой суммы

ОТСЛЕЖИВАНИЕ ЗАКАЗА (TrackOrder.tsx):
- Ввод номера заказа и телефона
- Отображение статуса заказа
- История статусов

Требования:
- Responsive дизайн
- Правильная обработка ошибок
- Loading states
- Form validation
- TypeScript типы для всех props

Используй Tailwind CSS или Material-UI для стилей.
Компоненты должны быть переиспользуемы.
```

**Что будет создано:**
- ✅ Все страницы для клиентов
- ✅ Компоненты с нужными функциями
- ✅ Стили и responsive дизайн
- ✅ Формы и валидация

---

### 2.8 ШАГ 8: Admin Panel (День 11-13)

**Prompt для Cursor:**

```
Создай админ-панель для управления магазинами и заказами.

ДАШБОРД (AdminDashboard.tsx):
- KPI: количество заказов, выручка, среднее значение заказа
- Графики продаж
- Топ товаров

УПРАВЛЕНИЕ МАГАЗИНАМИ (StoresPage.tsx):
- Таблица со всеми магазинами
- Кнопка создания нового магазина
- Форма редактирования магазина с полями:
  * Название, адрес, координаты
  * Радиус доставки
  * Минимальная сумма заказа
  * Максимальный вес заказа
  * Часы работы

УПРАВЛЕНИЕ ТОВАРАМИ (ProductsPage.tsx):
- Таблица товаров для выбранного магазина
- Импорт товаров из CSV (для 50-70 товаров)
- Форма создания товара
- Inline редактирование цены и остатков
- Фильтры по категориям

УПРАВЛЕНИЕ ЗАКАЗАМИ (OrdersPage.tsx):
- Таблица заказов с фильтрами (статус, дата, магазин)
- Детальная страница заказа
- Управление статусом (кнопки переходов)
- Назначение курьера
- История статусов

АНАЛИТИКА (AnalyticsPage.tsx):
- Дашборд с графиками
- Популярные товары
- Аналитика по районам
- Экспорт отчетов в PDF/Excel

Требования:
- Role-based access control (супер-админ, менеджер магазина)
- Таблицы с фильтрацией и сортировкой
- Модальные окна для форм
- Уведомления об успехе/ошибке
- Правильные права доступа

Используй таблицы (Material-UI DataGrid или react-table),
модальные окна, формы с валидацией.
```

**Что будет создано:**
- ✅ Все страницы админ-панели
- ✅ Таблицы с сортировкой и фильтрацией
- ✅ Формы управления
- ✅ Графики и аналитика

---

### 2.9 ШАГ 9: Интеграции (День 14-15)

**Prompt для Cursor:**

```
Интегрируй внешние сервисы в приложение.

WHATSAPP УВЕДОМЛЕНИЯ:
1. Backend:
   - Установи библиотеку для Twilio (или другого WhatsApp провайдера)
   - Создай WhatsAppService
   - Функция SendNotification(phone, message)
   - Шаблоны сообщений для разных статусов

2. Когда отправлять уведомления:
   - Новый заказ создан
   - Заказ подтвержден
   - Заказ собирается
   - Заказ в доставке
   - Заказ доставлен

KASPI.KZ ИНТЕГРАЦИЯ:
1. Backend:
   - Endpoint для создания платежа в Kaspi
   - Webhook endpoint для обработки callback'ов Kaspi
   - Проверка статуса платежа
   - Обновление статуса заказа после успешного платежа

2. Frontend:
   - Redirect на страницу оплаты Kaspi
   - Обработка успеха/ошибки
   - Показать информацию о платеже

HALYK BANK ИНТЕГРАЦИЯ:
- Аналогично Kaspi.kz

НАЛИЧНЫЕ ПЛАТЕЖИ:
- Просто отметить при доставке как "ожидание наличными"
- Курьер подтверждает при доставке

Требования:
- Secure API ключи (в .env)
- Error handling и retry логика
- Логирование платежей
- Уведомления при ошибке платежа

Используй Go библиотеки для интеграций.
```

**Что будет создано:**
- ✅ WhatsApp интеграция
- ✅ Kaspi.kz платежи
- ✅ Halyk Bank платежи
- ✅ Webhook обработчики

---

### 2.10 ШАГ 10: Docker & Deployment (День 15-16)

**Prompt для Cursor:**

```
Подготовь приложение к развертыванию с Docker.

Создай:
1. backend/Dockerfile
   - Multi-stage build (builder + runtime)
   - Go 1.21 базовый образ
   - Компиляция в production бинарник
   - Запуск на порту 8080

2. frontend/Dockerfile
   - Node image для build
   - Nginx для serving
   - Оптимизированный production build

3. docker-compose.yml (для локальной разработки)
   - PostgreSQL 14
   - Backend сервис (Go)
   - Frontend сервис (React/Vite)
   - Nginx reverse proxy
   - Volumes для БД и кода
   - Networks для коммуникации

4. nginx.conf
   - Routing для frontend (/)
   - Routing для backend API (/api/v1)
   - CORS headers
   - Кэширование статики

5. .dockerignore и .gitignore

Требования:
- Оптимизированные образы (минимальный размер)
- Правильные переменные окружения (.env)
- Health checks для контейнеров
- Volumes для persistent data (БД)
- Логирование в stdout

После создания тестируй:
docker-compose up
Проверь: http://localhost:3000 (frontend)
         http://localhost:8080/api/v1/health (backend)
```

**Что будет создано:**
- ✅ Dockerfile для backend и frontend
- ✅ docker-compose.yml для локальной разработки
- ✅ nginx конфигурация
- ✅ Production-ready setup

---

## 3. ПОЛНЫЙ ПРОЦЕСС ДЛЯ CURSOR

### 3.1 Mega Prompt (все в одном)

Если хотите быстрее - отправьте в Cursor этот mega prompt:

```
ВАЖНО: Это задание для создания системы онлайн-доставки овощей и фруктов.

СТЕК ТЕХНОЛОГИЙ:
- Backend: Go 1.21+ с Gin framework
- Database: PostgreSQL 14+
- Frontend: React 18 + TypeScript + Vite
- Deployment: Docker + docker-compose

ТЕКУЩЕЕ СОСТОЯНИЕ:
Нужно создать приложение с нуля.

ТРЕБОВАНИЯ:

1. BACKEND (Go + PostgreSQL):
   ✓ REST API с JWT аутентификацией
   ✓ GORM ORM для работы с БД
   ✓ Models: User, Store, Product, Order, District, DeliveryTimeSlot и т.д.
   ✓ Services: Auth, Product, Order, Delivery, Store, Notification
   ✓ API endpoints согласно раздела 7 в ТЗ
   ✓ Расчет расстояния для проверки доставки (haversine formula)
   ✓ WhatsApp интеграция для уведомлений
   ✓ Интеграция с Kaspi.kz и Halyk Bank платежами

2. FRONTEND (React + TypeScript):
   ✓ Страницы: Home, Catalog, Cart, Checkout, TrackOrder
   ✓ Admin Panel: Dashboard, Stores, Products, Orders, Analytics
   ✓ Компоненты: ProductCard, Cart, Header, etc
   ✓ Redux/Zustand для state management
   ✓ Axios для API запросов
   ✓ React Router для навигации
   ✓ Tailwind CSS или Material-UI для стилей
   ✓ Responsive дизайн

3. DATABASE (PostgreSQL):
   ✓ Миграции для всех таблиц (users, stores, products, orders и т.д.)
   ✓ Правильные типы данных, индексы, foreign keys
   ✓ UUID primary keys
   ✓ Timestamps (created_at, updated_at)

4. DOCKER:
   ✓ Dockerfile для backend (multi-stage)
   ✓ Dockerfile для frontend
   ✓ docker-compose.yml с PostgreSQL, Backend, Frontend, Nginx
   ✓ nginx.conf для reverse proxy
   ✓ .env.example с примерами переменных

ГЛАВНОЕ:
- Чистая архитектура (handlers → services → repositories)
- TypeScript везде
- Правильная обработка ошибок
- Логирование
- Готово к production

Начни с Backend Setup, потом Database, затем Frontend.
Работай пошагово, создавай структуру и файлы.

ФАЙЛЫ ТЗ:
1. TZ_VEGETABLE_SHOP.md - полное техническое задание (раздел 3, 4, 5, 6, 7, 8 - главные)
2. DELIVERY_SYSTEM_LOGIC.md - логика системы доставки (как работает радиус, районы)

Используй эти документы как справочник для деталей.

Готов? Давайте начнем!
```

---

## 4. QUICK START GUIDE

### 4.1 Локальная разработка (3 команды)

```bash
# 1. Клонируй и перейди в проект
git clone <repo> && cd veggies-shop

# 2. Запусти всё через Docker
docker-compose up -d

# 3. Готово!
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# API Docs: http://localhost:8080/swagger/index.html
```

### 4.2 Во время разработки

```bash
# Backend (в отдельном терминале)
cd backend
go run cmd/server/main.go

# Frontend (в отдельном терминале)
cd frontend
npm run dev

# БД (уже running в docker)
docker-compose logs postgres
```

### 4.3 Тестирование API

```bash
# Используйте Postman, Insomnia или curl

# Пример: Создать заказ
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": "...",
    "items": [{"product_id": "...", "quantity": 2}],
    "delivery_type": "regular",
    "district_id": "...",
    "address": "..."
  }'
```

---

## 5. ПОМОЩЬ И ПРОБЛЕМЫ

### 5.1 Cursor IDE Shortcuts

```
Ctrl+K (Cmd+K на Mac) - Open Cursor commands
Ctrl+L (Cmd+L) - Inline edit
Ctrl+Shift+L - Full page edit
Ctrl+I (Cmd+I) - Start new chat
Tab - Accept suggestion
```

### 5.2 Если что-то не работает

1. **PostgreSQL не запускается:**
   ```bash
   docker-compose down
   docker volume rm veggies-shop_postgres_data
   docker-compose up -d
   ```

2. **Port уже используется:**
   ```bash
   # Найти процесс
   lsof -i :8080
   # Убить его
   kill -9 <PID>
   ```

3. **Go модули не устанавливаются:**
   ```bash
   cd backend
   go clean -modcache
   go mod download
   ```

4. **Node modules конфликт:**
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   ```

---

## 6. ЧЕКЛИСТ РАЗРАБОТКИ

### Фаза 1: MVP (Неделя 1-2)
- [ ] ✅ Backend структура + DB миграции
- [ ] ✅ Basic CRUD API (товары, заказы)
- [ ] ✅ Frontend: Catalog, Cart, Checkout (без платежей)
- [ ] ✅ Админ-панель: управление товарами и заказами
- [ ] ✅ Docker setup
- [ ] ✅ Базовое отслеживание заказа

### Фаза 2: Интеграции (Неделя 3)
- [ ] ✅ WhatsApp уведомления
- [ ] ✅ Kaspi.kz платежи
- [ ] ✅ Halyk Bank платежи
- [ ] ✅ DeliveryService с проверкой радиуса

### Фаза 3: Полнота (Неделя 4)
- [ ] ✅ Управление районами
- [ ] ✅ Временные окна доставки (2-часовые слоты)
- [ ] ✅ Система назначения курьеров
- [ ] ✅ Улучшенная админ-панель

### Фаза 4: Аналитика (Неделя 5)
- [ ] ✅ Дашборд с KPI
- [ ] ✅ Графики продаж
- [ ] ✅ Отчеты по товарам и районам
- [ ] ✅ Экспорт отчетов

### Фаза 5: Финализ (Неделя 6)
- [ ] ✅ Тестирование
- [ ] ✅ Код ревью
- [ ] ✅ Documentation
- [ ] ✅ Production deployment

---

## 7. ПОЛЕЗНЫЕ РЕСУРСЫ

### Документация
- Go Gin Framework: https://gin-gonic.com/
- GORM: https://gorm.io/
- React: https://react.dev/
- Vite: https://vitejs.dev/
- PostgreSQL: https://www.postgresql.org/docs/

### Инструменты
- Postman: https://www.postman.com/ (тестирование API)
- DBeaver: https://dbeaver.io/ (управление БД)
- GitHub: https://github.com (версионирование)

### Примеры
- Golang Web API: https://github.com/golang-standards/project-layout
- React Admin: https://github.com/marmelab/react-admin

---

## ✅ ГОТОВО!

Теперь вы:
✅ Понимаете полную архитектуру системы
✅ Знаете стек технологий
✅ Имеете пошаговый план разработки
✅ Готовы использовать Cursor IDE для написания кода

**Следующий шаг:** Откройте Cursor IDE и начните с ШАГ 1 (Backend Setup).

**Вопросы?** Всегда можно вернуться к файлам ТЗ:
- `TZ_VEGETABLE_SHOP.md` - полное описание
- `DELIVERY_SYSTEM_LOGIC.md` - логика доставки

---

**Успехов в разработке! 🚀**

*версия 1.0 | 23.03.2026*
