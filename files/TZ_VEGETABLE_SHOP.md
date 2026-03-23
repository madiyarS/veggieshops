# 🥬 ТЕХНИЧЕСКОЕ ЗАДАНИЕ: Система онлайн-доставки овощей и фруктов
## "VeggieShops.kz" — Multi-Store Delivery Platform

**Дата:** 23.03.2026  
**Статус:** Ready for Development  
**Приоритет:** HIGH  

---

## 📋 1. ОБЗОР ПРОЕКТА

### 1.1 Описание
Веб-платформа для онлайн-заказа овощей, фруктов и зелени с функциональностью:
- **Клиентская часть**: каталог товаров, корзина, оформление заказа, отслеживание
- **Админ-панель**: управление магазинами, товарами, заказами, доставкой, аналитикой
- **Система доставки**: разные типы доставки, управление курьерами, радиус доставки
- **Интеграции**: WhatsApp уведомления, платежи (Kaspi.kz, Halyk Bank, наличные)

### 1.2 Целевая аудитория
- **Клиенты**: жители Астаны и других городов Казахстана
- **Администраторы**: владельцы магазинов и сетей
- **Курьеры**: внешние доставщики и сотрудники магазинов

### 1.3 Ключевые требования
- ✅ Сеть магазинов (неограниченное количество локаций)
- ✅ Полная админ-панель как "конструктор" (все настраивается)
- ✅ Система доставки с расчетом по радиусу
- ✅ Платежи через Kaspi.kz, Halyk Bank, наличные
- ✅ WhatsApp уведомления
- ✅ Отслеживание заказов по номеру
- ✅ Подробная аналитика

---

## 🏗️ 2. АРХИТЕКТУРА И СТЕК ТЕХНОЛОГИЙ

### 2.1 Backend
```
Language: Go 1.21+
Framework: Gin (REST API) или Echo
Database: PostgreSQL 14+
ORM: GORM
Authentication: JWT
Task Queue: Redis + Bull (уведомления)
File Storage: AWS S3 или локальное хранилище
```

### 2.2 Frontend
```
Framework: React 18 + TypeScript
UI Library: Material-UI v5 или Tailwind CSS
State Management: Redux Toolkit / Zustand
HTTP Client: Axios
Build Tool: Vite
Deployment: Nginx / Vercel
```

### 2.3 Infrastructure
```
Server: Linux (Ubuntu 22.04+)
Container: Docker + Docker Compose
Reverse Proxy: Nginx
API Documentation: Swagger/OpenAPI 3.0
Monitoring: Prometheus + Grafana (опционально)
```

---

## 📊 3. СТРУКТУРА БАЗЫ ДАННЫХ

### 3.1 Основные таблицы

#### **Users** (Пользователи)
```sql
- id (UUID, PK)
- phone (String, unique)
- email (String)
- password_hash (String)
- first_name (String)
- last_name (String)
- role (Enum: customer, admin, manager, courier)
- is_active (Boolean)
- created_at (Timestamp)
- updated_at (Timestamp)
```

#### **Stores** (Магазины)
```sql
- id (UUID, PK)
- name (String)
- description (Text)
- address (String)
- latitude (Float)
- longitude (Float)
- phone (String)
- email (String)
- delivery_radius_km (Float, default: 3)
- min_order_amount (Integer, default: 2500)
- max_order_weight_kg (Float)
- is_active (Boolean)
- working_hours_start (Time)
- working_hours_end (Time)
- owner_id (UUID, FK -> Users)
- created_at (Timestamp)
- updated_at (Timestamp)
```

#### **Districts** (Районы)
```sql
- id (UUID, PK)
- store_id (UUID, FK)
- name (String)
- distance_km (Float)
- delivery_fee_regular (Integer) // в тенге
- delivery_fee_express (Integer)
- is_active (Boolean)
- created_at (Timestamp)
```

#### **Categories** (Категории товаров)
```sql
- id (UUID, PK)
- name (String) // Овощи, Фрукты, Зелень, Ягоды, Травы, Молочные, Яйца
- description (String)
- icon_url (String)
- order (Integer) // для сортировки
- is_active (Boolean)
- created_at (Timestamp)
```

#### **Products** (Товары)
```sql
- id (UUID, PK)
- store_id (UUID, FK)
- category_id (UUID, FK)
- name (String)
- description (Text)
- price (Integer) // в тенге
- weight_gram (Integer) // вес основной единицы
- unit (String) // "шт", "кг", "л", "пучок"
- stock_quantity (Integer) // количество на складе
- image_url (String)
- origin (String) // происхождение (опционально)
- shelf_life_days (Integer) // срок годности (опционально)
- is_available (Boolean)
- is_active (Boolean)
- created_at (Timestamp)
- updated_at (Timestamp)
```

#### **DeliveryTimeSlots** (Временные окна доставки)
```sql
- id (UUID, PK)
- store_id (UUID, FK)
- day_of_week (Integer) // 0-6
- start_time (Time) // 09:00
- end_time (Time) // 11:00
- max_orders (Integer) // максимум заказов в слот
- is_active (Boolean)
- created_at (Timestamp)
```

#### **Orders** (Заказы)
```sql
- id (UUID, PK)
- order_number (String, unique) // ORD-20260323-001
- user_id (UUID, FK)
- store_id (UUID, FK)
- district_id (UUID, FK)
- status (Enum: pending, confirmed, preparing, in_delivery, delivered, cancelled)
- delivery_type (Enum: regular, express)
- delivery_time_slot_id (UUID, FK)
- delivery_address (String)
- customer_phone (String)
- customer_name (String)
- total_amount (Integer)
- delivery_fee (Integer)
- payment_method (Enum: kaspi, halyk, cash)
- payment_status (Enum: pending, completed, failed)
- courier_id (UUID, FK -> Users, nullable)
- notes (Text)
- created_at (Timestamp)
- updated_at (Timestamp)
```

#### **OrderItems** (Товары в заказе)
```sql
- id (UUID, PK)
- order_id (UUID, FK)
- product_id (UUID, FK)
- quantity (Integer)
- price_at_order (Integer) // цена в момент заказа
- subtotal (Integer)
- created_at (Timestamp)
```

#### **Couriers** (Курьеры - для сотрудников магазина)
```sql
- id (UUID, PK)
- user_id (UUID, FK)
- store_id (UUID, FK)
- vehicle_type (String) // "пешком", "велосипед", "мотоцикл", "авто"
- phone (String)
- is_active (Boolean)
- created_at (Timestamp)
```

#### **Notifications** (Уведомления)
```sql
- id (UUID, PK)
- order_id (UUID, FK)
- user_id (UUID, FK)
- channel (Enum: whatsapp, sms, email)
- status (Enum: pending, sent, failed)
- message (Text)
- sent_at (Timestamp)
- created_at (Timestamp)
```

#### **Analytics** (Аналитика - денормализованные данные)
```sql
- id (UUID, PK)
- store_id (UUID, FK)
- date (Date)
- total_orders (Integer)
- total_revenue (Integer)
- popular_product_id (UUID)
- avg_order_value (Integer)
- created_at (Timestamp)
```

---

## 🎨 4. ФУНКЦИОНАЛЬНОСТЬ

### 4.1 Клиентская часть (Веб-сайт)

#### 4.1.1 Главная страница
- [ ] Список доступных магазинов (по местоположению)
- [ ] Поисковая строка по товарам
- [ ] Категории товаров (слайдер/фильтр)
- [ ] Популярные товары этой недели
- [ ] Акции и спецпредложения

#### 4.1.2 Каталог товаров
- [ ] Фильтры по категориям
- [ ] Сортировка (по цене, популярности, новизне)
- [ ] Карточки товаров с фото, описанием, весом, ценой
- [ ] Быстрое добавление в корзину
- [ ] Модальное окно с полной информацией о товаре
- [ ] Система рейтинга и отзывов (опционально для v1)

#### 4.1.3 Корзина
- [ ] Список добавленных товаров
- [ ] Изменение количества
- [ ] Удаление товара
- [ ] Расчет промежуточной суммы
- [ ] Информация о минимальной сумме заказа

#### 4.1.4 Оформление заказа
- [ ] Выбор типа доставки (Regular/Express)
- [ ] Выбор района доставки
- [ ] Выбор временного окна доставки
- [ ] Расчет итоговой стоимости (товары + доставка)
- [ ] Ввод адреса доставки (текстовое поле или выбор из списка)
- [ ] Выбор способа оплаты (Kaspi.kz, Halyk Bank, Наличные)
- [ ] Поле для комментариев к заказу
- [ ] Кнопка "Оформить заказ"

#### 4.1.5 Отслеживание заказа
- [ ] Страница отслеживания (ввод номера заказа + телефон)
- [ ] Отображение статуса (pending → confirmed → preparing → in_delivery → delivered)
- [ ] Информация о магазине и курьере
- [ ] История статусов с временем обновления
- [ ] Кнопка "Связать с магазином" (WhatsApp)

#### 4.1.6 Личный кабинет (опционально для v1)
- [ ] История заказов
- [ ] Сохраненные адреса
- [ ] Профиль пользователя
- [ ] Настройки уведомлений

### 4.2 Админ-панель

#### 4.2.1 Управление магазинами
- [ ] Список всех магазинов
- [ ] Создание нового магазина
- [ ] Редактирование данных магазина
  - [ ] Название, адрес, координаты
  - [ ] Радиус доставки
  - [ ] Минимальная сумма заказа
  - [ ] Максимальный вес заказа
  - [ ] Часы работы
  - [ ] Контактные данные
- [ ] Удаление магазина
- [ ] Статус активности

#### 4.2.2 Управление районами доставки
- [ ] Список районов для каждого магазина
- [ ] Создание нового района
- [ ] Редактирование района
  - [ ] Название района
  - [ ] Расстояние до магазина (км)
  - [ ] Стоимость доставки Regular
  - [ ] Стоимость доставки Express
- [ ] Удаление района
- [ ] Статус активности

#### 4.2.3 Управление товарами
- [ ] Список товаров по магазинам
- [ ] Массовый импорт товаров (CSV/Excel для 50-70 товаров)
- [ ] Создание нового товара
- [ ] Редактирование товара
  - [ ] Категория
  - [ ] Название
  - [ ] Описание
  - [ ] Цена
  - [ ] Вес (грамм)
  - [ ] Единица измерения
  - [ ] Количество на складе
  - [ ] Загрузка фото
  - [ ] Происхождение
  - [ ] Срок годности
- [ ] Удаление товара
- [ ] Быстрое изменение цены/количества (inline редактирование)
- [ ] Фильтры по категориям и статусу

#### 4.2.4 Управление временными окнами доставки
- [ ] Таблица с интервалами доставки по дням недели
- [ ] Создание интервалов (2-часовые слоты: 09:00-11:00, 11:00-13:00 и т.д.)
- [ ] Редактирование времени и максимума заказов в слот
- [ ] Быстрое включение/отключение слотов
- [ ] Копирование расписания на другие дни

#### 4.2.5 Управление заказами
- [ ] Список всех заказов (с фильтрами и поиском)
  - [ ] По статусу
  - [ ] По магазину
  - [ ] По дате
  - [ ] По методу доплаты
- [ ] Детальная страница заказа
  - [ ] Полная информация о клиенте
  - [ ] Список товаров в заказе
  - [ ] История статусов с временем
  - [ ] Информация о доставке
- [ ] Управление статусом заказа
  - [ ] Кнопки для перехода между статусами
  - [ ] Подтверждение, отправка в производство
  - [ ] Отмена заказа (с причиной)
- [ ] Назначение курьера
  - [ ] Выбор из списка доступных курьеров
  - [ ] Отправка уведомления курьеру
- [ ] Отправка уведомления клиенту вручную
- [ ] Экспорт в PDF/печать квитанции

#### 4.2.6 Управление курьерами
- [ ] Список курьеров магазина
- [ ] Создание нового курьера
- [ ] Редактирование данных
  - [ ] ФИО
  - [ ] Телефон
  - [ ] Тип транспорта
  - [ ] Статус активности
- [ ] Удаление курьера
- [ ] Просмотр активных доставок курьера

#### 4.2.7 Управление платежами и платежными системами
- [ ] Интеграция с Kaspi.kz (webhook, проверка статуса)
- [ ] Интеграция с Halyk Bank (webhook, проверка статуса)
- [ ] Наличные платежи (ручное подтверждение при доставке)
- [ ] История платежей с фильтрами
- [ ] Статистика по методам оплаты

#### 4.2.8 Управление уведомлениями
- [ ] Настройка WhatsApp интеграции
  - [ ] API ключ
  - [ ] Номер отправителя
  - [ ] Шаблоны сообщений
- [ ] История отправленных сообщений
- [ ] Ручная отправка сообщений клиентам
- [ ] Статистика по доставке сообщений

#### 4.2.9 Аналитика и отчеты
- [ ] Дашборд с KPI
  - [ ] Количество заказов (сегодня, неделя, месяц)
  - [ ] Общая выручка
  - [ ] Среднее значение заказа
  - [ ] Конверсия
- [ ] График продаж (по дням, неделям, месяцам)
- [ ] Топ 10 популярных товаров
  - [ ] С количеством продаж
  - [ ] С выручкой
  - [ ] По категориям
- [ ] Аналитика по районам
  - [ ] Количество заказов по районам
  - [ ] Выручка по районам
  - [ ] Популярные товары в каждом районе
- [ ] Аналитика по магазинам (для главного админа)
  - [ ] Выручка по магазинам
  - [ ] Рейтинг магазинов по количеству заказов
- [ ] Отчеты по доставке
  - [ ] Время средней доставки
  - [ ] Процент вовремя доставленных заказов
- [ ] Экспорт отчетов (PDF, Excel)

#### 4.2.10 Управление администраторами и правами
- [ ] Список пользователей системы
- [ ] Создание администратора
- [ ] Редактирование ролей и прав доступа
  - [ ] Главный администратор (все права)
  - [ ] Менеджер магазина (только свой магазин)
  - [ ] Курьер (только свои доставки)
- [ ] Блокировка/разблокировка пользователя
- [ ] Логирование действий (опционально)

### 4.3 Мобильная версия для курьеров (опционально для v2)
- [ ] Приложение для iOS/Android
- [ ] Список активных заказов
- [ ] Маршрут доставки
- [ ] Отметить как "в доставке"
- [ ] Отметить как "доставлено"
- [ ] Фото доказательства доставки
- [ ] Сумма наличных платежей за день

---

## 🔌 5. ИНТЕГРАЦИИ И ВНЕШНИЕ СЕРВИСЫ

### 5.1 Платежные системы

#### Kaspi.kz
```
Метод: Webhook API
Документация: https://kaspi.kz/api
Функции:
- Создание платежа
- Проверка статуса платежа
- Webhook для подтверждения
- Возврат средств
```

#### Halyk Bank
```
Метод: API интеграция
Функции:
- Создание платежа
- Проверка статуса
- Webhook подтверждения
```

### 5.2 WhatsApp интеграция

```
Сервис: Twilio или WhatsApp Business API
Функционал:
- Отправка уведомлений о новом заказе
- Подтверждение заказа
- Уведомление о доставке
- Прямая ссылка на отслеживание
- Двусторонняя коммуникация (опционально)

Шаблоны сообщений:
1. "Заказ #{order_id} принят! Сумма: {amount}тг. Отследить: {tracking_link}"
2. "Ваш заказ подтвержден. Доставка: {time_slot}"
3. "Ваш заказ собирается... Курьер выедет через {minutes} минут"
4. "Заказ в доставке! Курьер {courier_name}. Номер: {courier_phone}"
5. "Заказ доставлен! Спасибо за покупку 🎉"
```

### 5.3 Хранилище файлов

```
Опция 1: AWS S3
Опция 2: MinIO (самостоятельное хранилище)
Опция 3: Локальное хранилище на сервере

Структура папок:
/products/{store_id}/{product_id}/
/orders/{store_id}/{order_id}/
```

---

## 🔐 6. БЕЗОПАСНОСТЬ И АУТЕНТИФИКАЦИЯ

### 6.1 Аутентификация
- [ ] JWT токены (Access + Refresh)
- [ ] Хеширование паролей (bcrypt)
- [ ] Rate limiting на API
- [ ] CORS настройки
- [ ] HTTPS обязателен

### 6.2 Авторизация
- [ ] Role-Based Access Control (RBAC)
- [ ] Проверка прав при каждом запросе
- [ ] Изоляция данных между магазинами

### 6.3 Защита данных
- [ ] Валидация всех входных данных
- [ ] SQL injection prevention (используется ORM)
- [ ] XSS protection
- [ ] CSRF tokens
- [ ] Логирование действий администраторов

---

## 📱 7. API ENDPOINTS (REST)

### 7.1 Public API (для клиентов)

```
// Товары
GET  /api/v1/products              - Список товаров
GET  /api/v1/products/{id}         - Детали товара
GET  /api/v1/categories            - Список категорий

// Магазины и доставка
GET  /api/v1/stores                - Список магазинов
GET  /api/v1/stores/{id}/districts - Районы доставки магазина
GET  /api/v1/stores/{id}/slots     - Временные окна доставки

// Заказы
POST /api/v1/orders                - Создать заказ
GET  /api/v1/orders/track          - Отследить заказ (query: order_number, phone)

// Аутентификация
POST /api/v1/auth/register         - Регистрация
POST /api/v1/auth/login            - Вход
POST /api/v1/auth/refresh          - Обновить токен
```

### 7.2 Admin API (для админ-панели)

```
// Магазины
GET    /api/v1/admin/stores
POST   /api/v1/admin/stores
PATCH  /api/v1/admin/stores/{id}
DELETE /api/v1/admin/stores/{id}

// Товары
GET    /api/v1/admin/stores/{storeId}/products
POST   /api/v1/admin/stores/{storeId}/products
POST   /api/v1/admin/stores/{storeId}/products/import  - Импорт из CSV
PATCH  /api/v1/admin/stores/{storeId}/products/{id}
DELETE /api/v1/admin/stores/{storeId}/products/{id}

// Заказы
GET    /api/v1/admin/orders
PATCH  /api/v1/admin/orders/{id}/status
PATCH  /api/v1/admin/orders/{id}/assign-courier
POST   /api/v1/admin/orders/{id}/notify

// Районы
GET    /api/v1/admin/stores/{storeId}/districts
POST   /api/v1/admin/stores/{storeId}/districts
PATCH  /api/v1/admin/stores/{storeId}/districts/{id}

// Временные окна
GET    /api/v1/admin/stores/{storeId}/time-slots
POST   /api/v1/admin/stores/{storeId}/time-slots
PATCH  /api/v1/admin/stores/{storeId}/time-slots/{id}

// Курьеры
GET    /api/v1/admin/stores/{storeId}/couriers
POST   /api/v1/admin/stores/{storeId}/couriers
PATCH  /api/v1/admin/stores/{storeId}/couriers/{id}

// Аналитика
GET    /api/v1/admin/analytics/dashboard
GET    /api/v1/admin/analytics/sales          - Продажи по датам
GET    /api/v1/admin/analytics/products       - Популярные товары
GET    /api/v1/admin/analytics/districts      - По районам
GET    /api/v1/admin/analytics/export         - Экспорт отчета

// Пользователи
GET    /api/v1/admin/users
POST   /api/v1/admin/users
PATCH  /api/v1/admin/users/{id}/role
PATCH  /api/v1/admin/users/{id}/block
```

---

## 📁 8. СТРУКТУРА ПРОЕКТА

```
project-root/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── routes.go
│   │   ├── models/
│   │   │   └── *.go
│   │   ├── repositories/
│   │   │   └── *.go
│   │   ├── services/
│   │   │   └── *.go
│   │   ├── config/
│   │   │   └── config.go
│   │   └── utils/
│   │       ├── jwt.go
│   │       ├── password.go
│   │       └── validators.go
│   ├── migrations/
│   │   └── *.sql
│   ├── go.mod
│   ├── go.sum
│   ├── .env.example
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Home.tsx
│   │   │   ├── Catalog.tsx
│   │   │   ├── Cart.tsx
│   │   │   ├── Checkout.tsx
│   │   │   ├── TrackOrder.tsx
│   │   │   └── Admin/
│   │   │       ├── Dashboard.tsx
│   │   │       ├── Stores.tsx
│   │   │       ├── Products.tsx
│   │   │       ├── Orders.tsx
│   │   │       ├── Districts.tsx
│   │   │       ├── TimeSlots.tsx
│   │   │       ├── Couriers.tsx
│   │   │       └── Analytics.tsx
│   │   ├── components/
│   │   │   ├── ProductCard.tsx
│   │   │   ├── Cart.tsx
│   │   │   ├── Header.tsx
│   │   │   └── ...
│   │   ├── hooks/
│   │   ├── services/
│   │   │   └── api.ts
│   │   ├── store/
│   │   │   ├── slices/
│   │   │   └── index.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── public/
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── docker-compose.yml
├── nginx.conf
├── README.md
└── .gitignore
```

---

## 🚀 9. ЭТАПЫ РАЗРАБОТКИ (ФАЗЫ)

### **Фаза 1: MVP (2-3 недели)**
- [x] Структура БД
- [x] Basic Backend API (CRUD товары, заказы)
- [x] Frontend каталог + корзина
- [x] Оформление заказа (без платежей)
- [x] Админ-панель: управление товарами, заказами
- [x] Отслеживание заказа (базовое)

### **Фаза 2: Интеграции (1-2 недели)**
- [ ] Kaspi.kz интеграция
- [ ] Halyk Bank интеграция
- [ ] WhatsApp уведомления
- [ ] Улучшенное отслеживание заказа

### **Фаза 3: Расширенная функциональность (1-2 недели)**
- [ ] Управление районами и доставкой
- [ ] Управление временными окнами
- [ ] Назначение курьеров
- [ ] Улучшенная админ-панель

### **Фаза 4: Аналитика (1 неделя)**
- [ ] Дашборд с KPI
- [ ] Граф продаж
- [ ] Отчеты по товарам и районам
- [ ] Экспорт отчетов

### **Фаза 5: Оптимизация и тестирование (1-2 недели)**
- [ ] Тестирование всех функций
- [ ] Оптимизация производительности
- [ ] Подготовка к production
- [ ] Documentation

---

## 🎯 10. ТЕХНИЧЕСКИЕ ТРЕБОВАНИЯ

### 10.1 Performance
- [ ] Время загрузки страницы < 2 сек
- [ ] API response time < 500ms
- [ ] Поддержка 1000+ одновременных пользователей
- [ ] CDN для статики

### 10.2 Масштабируемость
- [ ] Горизонтальное масштабирование backend
- [ ] Кэширование (Redis)
- [ ] Database репликация
- [ ] Load balancing

### 10.3 Надежность
- [ ] 99.5% uptime
- [ ] Резервные копии БД (ежедневно)
- [ ] Health checks и мониторинг
- [ ] Graceful shutdown

---

## 📝 11. ПРИМЕРЫ И СЦЕНАРИИ

### 11.1 Сценарий: Клиент делает заказ

1. Клиент заходит на сайт
2. Выбирает магазин (по его локации или из списка)
3. Смотрит каталог товаров
4. Добавляет товары в корзину
5. Переходит в корзину
6. Выбирает тип доставки (Regular/Express)
7. Выбирает район доставки
8. Система показывает стоимость доставки
9. Выбирает временное окно доставки
10. Вводит адрес доставки
11. Выбирает способ оплаты
12. Оформляет заказ
13. Перенаправляется на платеж (если онлайн)
14. Получает номер заказа
15. Может отследить по номеру + телефон

### 11.2 Сценарий: Администратор управляет магазином

1. Логинится в админ-панель
2. Видит дашборд с KPI
3. Идет в "Товары"
4. Импортирует 50-70 товаров из CSV
5. Редактирует цены и остатки
6. Идет в "Районы доставки"
7. Добавляет новый район
8. Устанавливает стоимость доставки
9. Идет в "Временные окна"
10. Добавляет 2-часовые слоты на день
11. Идет в "Заказы"
12. Видит новый заказ
13. Подтверждает его
14. Отправляет уведомление на WhatsApp
15. Переводит в статус "Собирается"
16. Назначает курьера
17. Отмечает как "В доставке"
18. После доставки отмечает как "Доставлено"
19. Получает уведомление в аналитике

---

## 📚 12. ДОКУМЕНТАЦИЯ И ТЕСТИРОВАНИЕ

- [ ] Swagger документация для всех API endpoints
- [ ] README с инструкциями по установке
- [ ] Переменные окружения (.env.example)
- [ ] Unit тесты (минимум 70% покрытие)
- [ ] Integration тесты для критических сценариев
- [ ] E2E тесты для основных workflows

---

## ⚙️ 13. РАЗВЕРТЫВАНИЕ И DEVOPS

### 13.1 Локальная разработка
```bash
docker-compose up
# Backend: http://localhost:8080
# Frontend: http://localhost:3000
# PostgreSQL: localhost:5432
```

### 13.2 Production
- [ ] Docker контейнеризация
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Автоматические тесты перед деплоем
- [ ] Blue-green deployment
- [ ] Мониторинг и логирование (ELK stack опционально)

---

## 🔄 14. ПРОЦЕСС РАЗРАБОТКИ ДЛЯ CURSOR IDE

### Инструкции для Cursor IDE:

1. **Инициализация проекта:**
   ```bash
   mkdir veggies-shop && cd veggies-shop
   
   # Backend
   mkdir -p backend && cd backend
   go mod init github.com/yourusername/veggies-shop
   
   # Frontend
   cd ../
   npm create vite@latest frontend -- --template react-ts
   ```

2. **Создать структуру файлов согласно п. 8**

3. **Реализовать в порядке приоритета:**
   - Database migrations (PostgreSQL)
   - Backend API (Golang + Gin/Echo)
   - Frontend компоненты (React)
   - Интеграции (Kaspi, WhatsApp)
   - Админ-панель
   - Аналитика

4. **Тестирование:**
   - Unit тесты для критических функций
   - Postman/Insomnia коллекция для API
   - Manual тестирование UI

5. **Deployment:**
   - Docker образы
   - docker-compose для локальной разработки
   - Инструкции для production

---

## ✅ 15. ЧЕКЛИСТ ЗАВЕРШЕНИЯ

- [ ] База данных спроектирована и протестирована
- [ ] Backend API реализован и документирован (Swagger)
- [ ] Frontend приложение функционально
- [ ] Интеграции с платежами работают
- [ ] WhatsApp уведомления отправляются
- [ ] Админ-панель полностью функциональна
- [ ] Аналитика работает и отображает корректные данные
- [ ] Тестирование завершено (unit, integration, e2e)
- [ ] Code review пройден
- [ ] Documentation написана
- [ ] Готово к production deployment

---

## 📞 16. КОНТАКТЫ И ВОПРОСЫ

**Заказчик:** Veggies Shop Network  
**Регион:** Казахстан (Астана)  
**Язык интерфейса:** Русский / Казахский  
**Валюта:** Казахстанский Тенге (KZT)  

---

**Версия документа:** 1.0  
**Дата создания:** 23.03.2026  
**Статус:** READY FOR DEVELOPMENT ✅
