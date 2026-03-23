-- Код выдачи заказа курьеру (называет клиент при встрече)
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_code VARCHAR(6) NOT NULL DEFAULT '000000';

UPDATE orders
SET delivery_code = upper(substr(md5(random()::text || id::text), 1, 6))
WHERE delivery_code = '000000';

ALTER TABLE orders ALTER COLUMN delivery_code DROP DEFAULT;
