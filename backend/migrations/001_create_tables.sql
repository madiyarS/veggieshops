-- VeggieShops.kz Database Migrations
-- Version: 001
-- Description: Create all tables for the vegetable delivery system

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- ENUM TYPES
-- =============================================================================

CREATE TYPE user_role AS ENUM ('customer', 'admin', 'manager', 'courier');
CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'preparing', 'in_delivery', 'delivered', 'cancelled');
CREATE TYPE delivery_type AS ENUM ('regular', 'express');
CREATE TYPE payment_method AS ENUM ('kaspi', 'halyk', 'cash');
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed');
CREATE TYPE notification_channel AS ENUM ('whatsapp', 'sms', 'email');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed');

-- =============================================================================
-- USERS - Пользователи
-- =============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role user_role NOT NULL DEFAULT 'customer',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

COMMENT ON TABLE users IS 'Пользователи системы: клиенты, администраторы, менеджеры, курьеры';

-- =============================================================================
-- STORES - Магазины
-- =============================================================================
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address VARCHAR(500) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    delivery_radius_km DOUBLE PRECISION NOT NULL DEFAULT 3.0,
    min_order_amount INTEGER NOT NULL DEFAULT 2500,
    max_order_weight_kg DOUBLE PRECISION,
    is_active BOOLEAN NOT NULL DEFAULT true,
    working_hours_start TIME,
    working_hours_end TIME,
    owner_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stores_is_active ON stores(is_active);
CREATE INDEX idx_stores_owner ON stores(owner_id);

COMMENT ON TABLE stores IS 'Магазины сети с координатами и параметрами доставки';
COMMENT ON COLUMN stores.delivery_radius_km IS 'Радиус доставки в км (по умолчанию 3)';

-- =============================================================================
-- CATEGORIES - Категории товаров
-- =============================================================================
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url VARCHAR(500),
    "order" INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_categories_order ON categories("order");

COMMENT ON TABLE categories IS 'Категории: Овощи, Фрукты, Зелень, Ягоды, Травы, Молочные, Яйца';

-- =============================================================================
-- DISTRICTS - Районы доставки
-- =============================================================================
CREATE TABLE districts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    distance_km DOUBLE PRECISION NOT NULL,
    delivery_fee_regular INTEGER NOT NULL,
    delivery_fee_express INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_districts_store ON districts(store_id);
CREATE INDEX idx_districts_active ON districts(store_id, is_active);

COMMENT ON TABLE districts IS 'Районы доставки с разной стоимостью Regular/Express';

-- =============================================================================
-- DISTRICT_STREETS - Улицы в районах
-- =============================================================================
CREATE TABLE district_streets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    street_name VARCHAR(255) NOT NULL,
    zip_code VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_district_streets_district ON district_streets(district_id);
CREATE INDEX idx_district_streets_name ON district_streets(street_name);

COMMENT ON TABLE district_streets IS 'Улицы, входящие в каждый район доставки';

-- =============================================================================
-- PRODUCTS - Товары
-- =============================================================================
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    weight_gram INTEGER,
    unit VARCHAR(20) NOT NULL DEFAULT 'шт',
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(500),
    origin VARCHAR(100),
    shelf_life_days INTEGER,
    is_available BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_products_store ON products(store_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_available ON products(store_id, is_available, is_active);

COMMENT ON TABLE products IS 'Товары с фото, весом, ценой и остатками';

-- =============================================================================
-- DELIVERY_TIME_SLOTS - Временные окна доставки
-- =============================================================================
CREATE TABLE delivery_time_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    max_orders INTEGER NOT NULL DEFAULT 10,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_time_slots_store ON delivery_time_slots(store_id);
CREATE INDEX idx_time_slots_day ON delivery_time_slots(store_id, day_of_week);

COMMENT ON TABLE delivery_time_slots IS '2-часовые окна доставки (09:00-11:00 и т.д.)';

-- =============================================================================
-- ORDERS - Заказы
-- =============================================================================
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id),
    store_id UUID NOT NULL REFERENCES stores(id),
    district_id UUID NOT NULL REFERENCES districts(id),
    status order_status NOT NULL DEFAULT 'pending',
    delivery_type delivery_type NOT NULL,
    delivery_time_slot_id UUID NOT NULL REFERENCES delivery_time_slots(id),
    delivery_address VARCHAR(500) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    customer_name VARCHAR(200) NOT NULL,
    total_amount INTEGER NOT NULL,
    delivery_fee INTEGER NOT NULL,
    payment_method payment_method NOT NULL,
    payment_status payment_status NOT NULL DEFAULT 'pending',
    courier_id UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_store ON orders(store_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at);
CREATE INDEX idx_orders_customer_phone ON orders(customer_phone);

COMMENT ON TABLE orders IS 'Заказы со статусами жизненного цикла';

-- =============================================================================
-- ORDER_ITEMS - Товары в заказе
-- =============================================================================
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_order INTEGER NOT NULL,
    subtotal INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

COMMENT ON TABLE order_items IS 'Товары в заказе с фиксированной ценой на момент заказа';

-- =============================================================================
-- COURIERS - Курьеры
-- =============================================================================
CREATE TABLE couriers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    vehicle_type VARCHAR(50),
    phone VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_couriers_store ON couriers(store_id);
CREATE INDEX idx_couriers_user ON couriers(user_id);

COMMENT ON TABLE couriers IS 'Курьеры магазина (сотрудники и внешние)';

-- =============================================================================
-- NOTIFICATIONS - Уведомления
-- =============================================================================
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    channel notification_channel NOT NULL,
    status notification_status NOT NULL DEFAULT 'pending',
    message TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_order ON notifications(order_id);
CREATE INDEX idx_notifications_status ON notifications(status);

COMMENT ON TABLE notifications IS 'Очередь WhatsApp/SMS/Email уведомлений';

-- =============================================================================
-- ANALYTICS - Аналитика
-- =============================================================================
CREATE TABLE analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_orders INTEGER NOT NULL DEFAULT 0,
    total_revenue INTEGER NOT NULL DEFAULT 0,
    popular_product_id UUID REFERENCES products(id),
    avg_order_value INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_analytics_store_date ON analytics(store_id, date);
CREATE INDEX idx_analytics_date ON analytics(date);

COMMENT ON TABLE analytics IS 'Денормализованные данные для аналитики';

-- =============================================================================
-- UPDATED_AT TRIGGER
-- =============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_stores_updated_at
    BEFORE UPDATE ON stores
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
