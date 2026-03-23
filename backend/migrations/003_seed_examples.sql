-- Примеры данных для тестирования
-- Запускать после 001 и 002

-- 1. Admin пользователь (телефон 77000000000 или +77000000000, пароль admin123)
INSERT INTO users (id, phone, email, password_hash, first_name, last_name, role, is_active)
VALUES (
  uuid_generate_v4(),
  '77000000000',
  'admin@veggieshops.kz',
  '$2a$10$mWYzZuxmlddSPsBBp9ISyO867.8myuB/yij8/aaqPYCRQKv8iYXL.',
  'Админ',
  'VeggieShops.kz',
  'admin',
  true
) ON CONFLICT (phone) DO UPDATE SET role = 'admin', password_hash = EXCLUDED.password_hash;

-- 2. Магазин (если ещё нет)
INSERT INTO stores (id, name, description, address, latitude, longitude, phone, delivery_radius_km, min_order_amount, is_active)
SELECT uuid_generate_v4(), 'Зелень & Здоровье', 'Свежие овощи и фрукты с доставкой', 'ул. Достык 10, Астана', 51.1694, 71.4491, '+77071234567', 3, 2500, true
WHERE NOT EXISTS (SELECT 1 FROM stores);

-- 3. Районы доставки (для первого магазина)
INSERT INTO districts (store_id, name, distance_km, delivery_fee_regular, delivery_fee_express, is_active)
SELECT s.id, 'Астана-Центр', 1.5, 500, 1000, true FROM stores s
WHERE NOT EXISTS (SELECT 1 FROM districts d WHERE d.store_id = s.id AND d.name = 'Астана-Центр')
LIMIT 1;

INSERT INTO districts (store_id, name, distance_km, delivery_fee_regular, delivery_fee_express, is_active)
SELECT s.id, 'Есиль', 2.8, 700, 1200, true FROM stores s
WHERE NOT EXISTS (SELECT 1 FROM districts d WHERE d.store_id = s.id AND d.name = 'Есиль')
LIMIT 1;

INSERT INTO districts (store_id, name, distance_km, delivery_fee_regular, delivery_fee_express, is_active)
SELECT s.id, 'Сарыарка', 2.2, 600, 1100, true FROM stores s
WHERE NOT EXISTS (SELECT 1 FROM districts d WHERE d.store_id = s.id AND d.name = 'Сарыарка')
LIMIT 1;

-- 4. Временные окна (Пн-Вс, слоты 09-11, 11-13, 13-15)
INSERT INTO delivery_time_slots (store_id, day_of_week, start_time, end_time, max_orders, is_active)
SELECT s.id, d, '09:00', '11:00', 10, true FROM stores s CROSS JOIN generate_series(0, 6) AS d
WHERE NOT EXISTS (SELECT 1 FROM delivery_time_slots WHERE store_id = s.id);

INSERT INTO delivery_time_slots (store_id, day_of_week, start_time, end_time, max_orders, is_active)
SELECT s.id, d, '11:00', '13:00', 10, true FROM stores s CROSS JOIN generate_series(0, 6) AS d
WHERE (SELECT COUNT(*) FROM delivery_time_slots) < 8;

INSERT INTO delivery_time_slots (store_id, day_of_week, start_time, end_time, max_orders, is_active)
SELECT s.id, d, '13:00', '15:00', 8, true FROM stores s CROSS JOIN generate_series(0, 6) AS d
WHERE (SELECT COUNT(*) FROM delivery_time_slots) < 15;

-- 5. Товары (первый магазин по дате создания; без дублей при повторном запуске сида)
INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Помидоры', 'Свежие томаты', 450, 1000, 'кг', 50, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Овощи'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Помидоры');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Огурцы', 'Свежие огурцы', 350, 1000, 'кг', 40, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Овощи'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Огурцы');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Картофель', 'Молодой картофель', 250, 1000, 'кг', 100, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Овощи'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Картофель');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Морковь', 'Сладкая морковь', 200, 1000, 'кг', 60, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Овощи'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Морковь');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Яблоки', 'Красные яблоки', 550, 1000, 'кг', 30, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Фрукты'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Яблоки');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Бананы', 'Спелые бананы', 650, 1000, 'кг', 25, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Фрукты'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Бананы');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Укроп', 'Свежий укроп', 150, 50, 'пучок', 20, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Зелень'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Укроп');

INSERT INTO products (store_id, category_id, name, description, price, weight_gram, unit, stock_quantity, is_available, is_active)
SELECT st.id, c.id, 'Петрушка', 'Свежая петрушка', 120, 50, 'пучок', 15, true, true
FROM (SELECT id FROM stores ORDER BY created_at LIMIT 1) st
JOIN categories c ON c.name = 'Зелень'
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.store_id = st.id AND p.name = 'Петрушка');
