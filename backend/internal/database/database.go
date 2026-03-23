package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ApplyRuntimePatches применяет идемпотентные правки схемы для БД, созданных до появления новых SQL-миграций в initdb.
func ApplyRuntimePatches(db *gorm.DB) error {
	if err := ensureOrdersDeliveryCode(db); err != nil {
		return fmt.Errorf("orders.delivery_code: %w", err)
	}
	if err := ensureStoreInventory(db); err != nil {
		return fmt.Errorf("store_inventory: %w", err)
	}
	if err := ensureAdvancedStock(db); err != nil {
		return fmt.Errorf("advanced_stock: %w", err)
	}
	return nil
}

func ensureOrdersDeliveryCode(db *gorm.DB) error {
	if err := db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_code VARCHAR(6)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE orders
		SET delivery_code = upper(substr(md5(random()::text || id::text), 1, 6))
		WHERE delivery_code IS NULL OR btrim(delivery_code::text) = ''
	`).Error; err != nil {
		return err
	}
	return db.Exec(`ALTER TABLE orders ALTER COLUMN delivery_code SET NOT NULL`).Error
}

func ensureStoreInventory(db *gorm.DB) error {
	if err := db.Exec(`
CREATE TABLE IF NOT EXISTS store_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, product_id)
)`).Error; err != nil {
		return err
	}
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_store_inventory_store ON store_inventory(store_id)`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_store_inventory_product ON store_inventory(product_id)`).Error
	if err := db.Exec(`
INSERT INTO store_inventory (store_id, product_id, quantity)
SELECT p.store_id, p.id, GREATEST(p.stock_quantity, 0)
FROM products p
WHERE NOT EXISTS (
    SELECT 1 FROM store_inventory si
    WHERE si.store_id = p.store_id AND si.product_id = p.id
)`).Error; err != nil {
		return err
	}
	return nil
}

// ensureAdvancedStock зоны, партии, резерв, движения, поля товара (идемпотентно).
func ensureAdvancedStock(db *gorm.DB) error {
	stmts := []string{
		`ALTER TABLE store_inventory ADD COLUMN IF NOT EXISTS reserved_quantity INTEGER NOT NULL DEFAULT 0`,
		`UPDATE store_inventory SET reserved_quantity = 0 WHERE reserved_quantity IS NULL`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS stock_committed BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS inventory_unit VARCHAR(20) NOT NULL DEFAULT 'piece'`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS package_grams INTEGER NULL`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS is_seasonal BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS temporarily_unavailable BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS substitute_product_id UUID NULL`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS reorder_min_qty INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS cart_step_grams INTEGER NOT NULL DEFAULT 500`,
		`CREATE TABLE IF NOT EXISTS store_stock_zones (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			code VARCHAR(32) NOT NULL,
			name VARCHAR(100) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (store_id, code)
		)`,
		`CREATE TABLE IF NOT EXISTS suppliers (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS stock_batches (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			zone_id UUID NOT NULL REFERENCES store_stock_zones(id) ON DELETE RESTRICT,
			quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
			received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NULL,
			supplier_id UUID NULL REFERENCES suppliers(id) ON DELETE SET NULL,
			note TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_batches_store_product ON stock_batches(store_id, product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_batches_expires ON stock_batches(store_id, expires_at)`,
		`CREATE TABLE IF NOT EXISTS stock_movements (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			batch_id UUID NULL REFERENCES stock_batches(id) ON DELETE SET NULL,
			zone_id UUID NULL REFERENCES store_stock_zones(id) ON DELETE SET NULL,
			delta INTEGER NOT NULL,
			movement_type VARCHAR(32) NOT NULL,
			ref_order_id UUID NULL REFERENCES orders(id) ON DELETE SET NULL,
			reason TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_movements_store_created ON stock_movements(store_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS stock_receipts (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			supplier_id UUID NULL REFERENCES suppliers(id) ON DELETE SET NULL,
			note TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS stock_receipt_lines (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			receipt_id UUID NOT NULL REFERENCES stock_receipts(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			zone_id UUID NOT NULL REFERENCES store_stock_zones(id) ON DELETE RESTRICT,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			expires_at TIMESTAMPTZ NULL
		)`,
		`CREATE TABLE IF NOT EXISTS inventory_audit_sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
			note TEXT,
			started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ NULL
		)`,
		`CREATE TABLE IF NOT EXISTS inventory_audit_lines (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			session_id UUID NOT NULL REFERENCES inventory_audit_sessions(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			zone_id UUID NULL REFERENCES store_stock_zones(id) ON DELETE SET NULL,
			counted_qty INTEGER NOT NULL,
			system_qty_snapshot INTEGER NOT NULL,
			diff_qty INTEGER NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	_ = db.Exec(`
		INSERT INTO store_stock_zones (store_id, code, name, sort_order)
		SELECT s.id, v.code, v.name, v.ord
		FROM stores s
		CROSS JOIN (
			VALUES ('sales_floor', 'Зал', 1), ('fridge', 'Холодильник', 2), ('backroom', 'Подсобка', 3)
		) AS v(code, name, ord)
		ON CONFLICT (store_id, code) DO NOTHING
	`).Error
	_ = db.Exec(`
		INSERT INTO stock_batches (store_id, product_id, zone_id, quantity, received_at, expires_at)
		SELECT si.store_id, si.product_id, z.id, si.quantity, NOW(), NULL
		FROM store_inventory si
		JOIN store_stock_zones z ON z.store_id = si.store_id AND z.code = 'sales_floor'
		WHERE si.quantity > 0
		  AND NOT EXISTS (
			SELECT 1 FROM stock_batches b
			WHERE b.store_id = si.store_id AND b.product_id = si.product_id
		  )
	`).Error
	_ = db.Exec(`
		DO $c$
		BEGIN
		  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_substitute_product_id_fkey') THEN
			ALTER TABLE products
			  ADD CONSTRAINT products_substitute_product_id_fkey
			  FOREIGN KEY (substitute_product_id) REFERENCES products(id) ON DELETE SET NULL;
		  END IF;
		END $c$
	`).Error
	_ = db.Exec(`
		DO $c$
		BEGIN
		  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_store_inventory_reserved_le_qty') THEN
			ALTER TABLE store_inventory
			  ADD CONSTRAINT chk_store_inventory_reserved_le_qty CHECK (reserved_quantity <= quantity);
		  END IF;
		END $c$
	`).Error
	return nil
}

func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	slog.Info("Database connected")
	return db, nil
}

// ConnectWithRetry подключается к Postgres с повторными попытками (Docker: backend часто стартует раньше postgres).
func ConnectWithRetry(dsn string, attempts int, delay time.Duration) (*gorm.DB, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 1; i <= attempts; i++ {
		db, err := New(dsn)
		if err == nil {
			return db, nil
		}
		lastErr = err
		slog.Warn("database not ready, retrying", "attempt", i, "max", attempts, "error", err)
		if i < attempts {
			time.Sleep(delay)
		}
	}
	return nil, lastErr
}
