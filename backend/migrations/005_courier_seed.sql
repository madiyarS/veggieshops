-- Тестовый курьер: телефон +77000000001, пароль как у админа (admin123)
INSERT INTO users (id, phone, email, password_hash, first_name, last_name, role, is_active)
SELECT uuid_generate_v4(), '77000000001', 'courier@veggieshops.kz',
       (SELECT password_hash FROM users WHERE phone = '77000000000' LIMIT 1),
       'Курьер', 'Тестовый', 'courier', true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE phone = '77000000001');

INSERT INTO couriers (id, user_id, store_id, phone, is_active)
SELECT uuid_generate_v4(), u.id, (SELECT id FROM stores ORDER BY created_at LIMIT 1), '+77000000001', true
FROM users u
WHERE u.phone = '77000000001'
  AND EXISTS (SELECT 1 FROM stores)
  AND NOT EXISTS (SELECT 1 FROM couriers c WHERE c.user_id = u.id);
