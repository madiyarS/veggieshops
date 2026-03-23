-- Склад: остатки по паре (магазин, товар), отдельно от карточки номенклатуры
CREATE TABLE IF NOT EXISTS store_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_store_inventory_store ON store_inventory(store_id);
CREATE INDEX IF NOT EXISTS idx_store_inventory_product ON store_inventory(product_id);

COMMENT ON TABLE store_inventory IS 'Остатки склада магазина; списание при продаже';

INSERT INTO store_inventory (store_id, product_id, quantity)
SELECT p.store_id, p.id, GREATEST(p.stock_quantity, 0)
FROM products p
WHERE NOT EXISTS (
    SELECT 1 FROM store_inventory si
    WHERE si.store_id = p.store_id AND si.product_id = p.id
);
