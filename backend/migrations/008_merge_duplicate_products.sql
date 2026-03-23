-- Слияние дублей товара: одинаковые store_id + name (часто после повторного запуска 003_seed_examples.sql).
-- Оставляем самую раннюю карточку (created_at, id), остальные переносим в неё и удаляем.

DO $$
DECLARE
  grp RECORD;
  keep_id UUID;
  dup_id UUID;
BEGIN
  FOR grp IN
    SELECT store_id, name
    FROM products
    GROUP BY store_id, name
    HAVING COUNT(*) > 1
  LOOP
    SELECT p.id INTO keep_id
    FROM products p
    WHERE p.store_id = grp.store_id AND p.name = grp.name
    ORDER BY p.created_at NULLS LAST, p.id
    LIMIT 1;

    FOR dup_id IN
      SELECT p.id
      FROM products p
      WHERE p.store_id = grp.store_id AND p.name = grp.name AND p.id <> keep_id
      ORDER BY p.id
    LOOP
      -- store_inventory: сначала снимаем строки dupe, где уже есть строка keep на тот же склад
      WITH del AS (
        DELETE FROM store_inventory d
        WHERE d.product_id = dup_id
          AND EXISTS (
            SELECT 1 FROM store_inventory k
            WHERE k.store_id = d.store_id AND k.product_id = keep_id
          )
        RETURNING d.store_id, d.quantity, d.reserved_quantity
      )
      UPDATE store_inventory k
      SET quantity = k.quantity + del.quantity,
          reserved_quantity = k.reserved_quantity + del.reserved_quantity
      FROM del
      WHERE k.store_id = del.store_id AND k.product_id = keep_id;

      UPDATE store_inventory SET product_id = keep_id WHERE product_id = dup_id;

      UPDATE stock_batches SET product_id = keep_id WHERE product_id = dup_id;
      UPDATE stock_movements SET product_id = keep_id WHERE product_id = dup_id;
      UPDATE stock_receipt_lines SET product_id = keep_id WHERE product_id = dup_id;
      UPDATE inventory_audit_lines SET product_id = keep_id WHERE product_id = dup_id;

      UPDATE products SET substitute_product_id = keep_id
      WHERE substitute_product_id = dup_id;

      UPDATE order_items SET product_id = keep_id WHERE product_id = dup_id;
      UPDATE analytics SET popular_product_id = keep_id WHERE popular_product_id = dup_id;

      DELETE FROM products WHERE id = dup_id;
    END LOOP;
  END LOOP;
END $$;
