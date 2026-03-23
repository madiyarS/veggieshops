package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents user role enum
type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleCourier  UserRole = "courier"
)

// OrderStatus represents order status enum
type OrderStatus string

const (
	OrderPending     OrderStatus = "pending"
	OrderConfirmed   OrderStatus = "confirmed"
	OrderPreparing   OrderStatus = "preparing"
	OrderInDelivery  OrderStatus = "in_delivery"
	OrderDelivered   OrderStatus = "delivered"
	OrderCancelled   OrderStatus = "cancelled"
)

// DeliveryType represents delivery type enum
type DeliveryType string

const (
	DeliveryRegular DeliveryType = "regular"
	DeliveryExpress DeliveryType = "express"
)

// PaymentMethod represents payment method enum
type PaymentMethod string

const (
	PaymentKaspi PaymentMethod = "kaspi"
	PaymentHalyk PaymentMethod = "halyk"
	PaymentCash  PaymentMethod = "cash"
)

// PaymentStatus represents payment status enum
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted  PaymentStatus = "completed"
	PaymentFailed     PaymentStatus = "failed"
)

// NotificationChannel represents notification channel enum
type NotificationChannel string

const (
	ChannelWhatsApp NotificationChannel = "whatsapp"
	ChannelSMS      NotificationChannel = "sms"
	ChannelEmail   NotificationChannel = "email"
)

// NotificationStatus represents notification status enum
type NotificationStatus string

const (
	NotifPending NotificationStatus = "pending"
	NotifSent    NotificationStatus = "sent"
	NotifFailed  NotificationStatus = "failed"
)

// User model
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Phone        string     `gorm:"uniqueIndex;size:20;not null" json:"phone"`
	Email        string     `gorm:"size:255" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	FirstName    string     `gorm:"size:100" json:"first_name"`
	LastName     string     `gorm:"size:100" json:"last_name"`
	Role         UserRole   `gorm:"type:varchar(20);not null;default:customer" json:"role"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Store model
type Store struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Name             string     `gorm:"size:255;not null" json:"name"`
	Description      string     `gorm:"type:text" json:"description"`
	Address          string     `gorm:"size:500;not null" json:"address"`
	Latitude         float64    `gorm:"not null" json:"latitude"`
	Longitude        float64    `gorm:"not null" json:"longitude"`
	Phone            string     `gorm:"size:20" json:"phone"`
	Email            string     `gorm:"size:255" json:"email"`
	DeliveryRadiusKm float64    `gorm:"not null;default:3" json:"delivery_radius_km"`
	MinOrderAmount   int        `gorm:"not null;default:2500" json:"min_order_amount"`
	MaxOrderWeightKg *float64   `json:"max_order_weight_kg"`
	IsActive         bool       `gorm:"not null;default:true" json:"is_active"`
	WorkingHoursStart *string   `gorm:"type:time" json:"working_hours_start"`
	WorkingHoursEnd   *string   `gorm:"type:time" json:"working_hours_end"`
	OwnerID          *uuid.UUID `gorm:"type:uuid" json:"owner_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (s *Store) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// Category model
type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"size:500" json:"icon_url"`
	Order       int       `gorm:"not null;default:0" json:"order"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// District model
type District struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StoreID            uuid.UUID `gorm:"type:uuid;not null" json:"store_id"`
	Name               string    `gorm:"size:255;not null" json:"name"`
	DistanceKm         float64   `gorm:"not null" json:"distance_km"`
	DeliveryFeeRegular int       `gorm:"not null" json:"delivery_fee_regular"`
	DeliveryFeeExpress int       `gorm:"not null" json:"delivery_fee_express"`
	IsActive           bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	Streets            []DistrictStreet `gorm:"foreignKey:DistrictID" json:"streets,omitempty"`
}

func (d *District) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DistrictStreet model
type DistrictStreet struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DistrictID uuid.UUID `gorm:"type:uuid;not null" json:"district_id"`
	StreetName string   `gorm:"size:255;not null" json:"street_name"`
	ZipCode    string   `gorm:"size:20" json:"zip_code"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ds *DistrictStreet) BeforeCreate(tx *gorm.DB) error {
	if ds.ID == uuid.Nil {
		ds.ID = uuid.New()
	}
	return nil
}

// InventoryUnit: piece — количество в штуках; weight_gram — остаток и заказ в граммах, price = цена за 1 кг.
type InventoryUnit string

const (
	InventoryUnitPiece      InventoryUnit = "piece"
	InventoryUnitWeightGram InventoryUnit = "weight_gram"
)

// Product model
type Product struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StoreID        uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	CategoryID     uuid.UUID  `gorm:"type:uuid;not null" json:"category_id"`
	Name           string     `gorm:"size:255;not null" json:"name"`
	Description    string     `gorm:"type:text" json:"description"`
	Price          int        `gorm:"not null" json:"price"`
	WeightGram     int        `json:"weight_gram"`
	Unit           string     `gorm:"size:20;not null;default:шт" json:"unit"`
	StockQuantity  int        `gorm:"not null;default:0" json:"stock_quantity"`
	ImageURL       string     `gorm:"size:500" json:"image_url"`
	Origin         string     `gorm:"size:100" json:"origin"`
	ShelfLifeDays  *int       `json:"shelf_life_days"`
	IsAvailable    bool       `gorm:"not null;default:true" json:"is_available"`
	IsActive       bool       `gorm:"not null;default:true" json:"is_active"`
	// Склад и витрина
	InventoryUnit           InventoryUnit `gorm:"type:varchar(20);not null;default:piece" json:"inventory_unit"`
	PackageGrams            *int          `json:"package_grams,omitempty"`
	IsSeasonal              bool          `gorm:"not null;default:false" json:"is_seasonal"`
	TemporarilyUnavailable  bool          `gorm:"not null;default:false" json:"temporarily_unavailable"`
	SubstituteProductID     *uuid.UUID    `gorm:"type:uuid" json:"substitute_product_id,omitempty"`
	ReorderMinQty           int           `gorm:"not null;default:0" json:"reorder_min_qty"`
	CartStepGrams           int           `gorm:"not null;default:250" json:"cart_step_grams"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Category          *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	SubstituteProduct *Product  `gorm:"foreignKey:SubstituteProductID" json:"substitute,omitempty"`
	// Вычисляемые (не в БД как отдельные колонки)
	StockReserved int `gorm:"-" json:"stock_reserved,omitempty"`
	StockPhysical int `gorm:"-" json:"stock_physical,omitempty"`
	// Витрина: ближайший срок годности по партиям (FEFO); null если партий со сроком нет
	NearestBatchExpiresAt *time.Time `gorm:"-" json:"nearest_batch_expires_at,omitempty"`
	// Витрина: остаток ниже порога дозаказа (reorder_min_qty), но > 0
	CatalogLowStock bool `gorm:"-" json:"catalog_low_stock,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// StoreInventory остаток склада магазина по товару (отдельно от карточки номенклатуры).
type StoreInventory struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StoreID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_inv_store_product,priority:1" json:"store_id"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_inv_store_product,priority:2" json:"product_id"`
	Quantity          int       `gorm:"not null;default:0" json:"quantity"`
	ReservedQuantity  int       `gorm:"not null;default:0" json:"reserved_quantity"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (si *StoreInventory) BeforeCreate(tx *gorm.DB) error {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return nil
}

// TableName явное имя таблицы (иначе GORM использует store_inventories и запросы к складу падают).
func (StoreInventory) TableName() string {
	return "store_inventory"
}

// DeliveryTimeSlot model
type DeliveryTimeSlot struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StoreID   uuid.UUID `gorm:"type:uuid;not null" json:"store_id"`
	DayOfWeek int       `gorm:"not null" json:"day_of_week"`
	StartTime string    `gorm:"type:time;not null" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null" json:"end_time"`
	MaxOrders int       `gorm:"not null;default:10" json:"max_orders"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (d *DeliveryTimeSlot) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// Order model
type Order struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	OrderNumber         string         `gorm:"uniqueIndex;size:50;not null" json:"order_number"`
	UserID              *uuid.UUID     `gorm:"type:uuid" json:"user_id"`
	StoreID             uuid.UUID      `gorm:"type:uuid;not null" json:"store_id"`
	DistrictID          uuid.UUID      `gorm:"type:uuid;not null" json:"district_id"`
	Status              OrderStatus    `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	DeliveryType        DeliveryType   `gorm:"type:varchar(20);not null" json:"delivery_type"`
	DeliveryTimeSlotID  uuid.UUID      `gorm:"type:uuid;not null" json:"delivery_time_slot_id"`
	DeliveryAddress     string         `gorm:"size:500;not null" json:"delivery_address"`
	CustomerPhone       string         `gorm:"size:20;not null" json:"customer_phone"`
	CustomerName        string         `gorm:"size:200;not null" json:"customer_name"`
	TotalAmount         int            `gorm:"not null" json:"total_amount"`
	DeliveryFee         int            `gorm:"not null" json:"delivery_fee"`
	PaymentMethod       PaymentMethod  `gorm:"type:varchar(20);not null" json:"payment_method"`
	PaymentStatus       PaymentStatus  `gorm:"type:varchar(20);not null;default:pending" json:"payment_status"`
	CourierID           *uuid.UUID     `gorm:"type:uuid" json:"courier_id"`
	DeliveryCode        string         `gorm:"size:6;not null" json:"delivery_code"`
	Notes               string         `gorm:"type:text" json:"notes"`
	StockCommitted      bool           `gorm:"not null;default:false" json:"stock_committed"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	Items               []*OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// OrderItem model
type OrderItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	OrderID      uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	ProductID    uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	PriceAtOrder int       `gorm:"not null" json:"price_at_order"`
	Subtotal     int       `gorm:"not null" json:"subtotal"`
	CreatedAt    time.Time `json:"created_at"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}

// Courier model
type Courier struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	StoreID     uuid.UUID `gorm:"type:uuid;not null" json:"store_id"`
	VehicleType string    `gorm:"size:50" json:"vehicle_type"`
	Phone       string    `gorm:"size:20;not null" json:"phone"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Courier) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Notification model
type Notification struct {
	ID        uuid.UUID          `gorm:"type:uuid;primary_key" json:"id"`
	OrderID   uuid.UUID          `gorm:"type:uuid;not null" json:"order_id"`
	UserID    *uuid.UUID         `gorm:"type:uuid" json:"user_id"`
	Channel   NotificationChannel `gorm:"type:varchar(20);not null" json:"channel"`
	Status    NotificationStatus  `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	Message   string             `gorm:"type:text" json:"message"`
	SentAt    *time.Time         `json:"sent_at"`
	CreatedAt time.Time          `json:"created_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// Analytics model
type Analytics struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StoreID          uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	Date             string     `gorm:"type:date;not null" json:"date"`
	TotalOrders      int        `gorm:"not null;default:0" json:"total_orders"`
	TotalRevenue     int        `gorm:"not null;default:0" json:"total_revenue"`
	PopularProductID *uuid.UUID `gorm:"type:uuid" json:"popular_product_id"`
	AvgOrderValue    *int       `json:"avg_order_value"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (a *Analytics) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// StoreStockZone зона хранения в магазине (зал / холодильник / подсобка).
type StoreStockZone struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StoreID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_zone_store_code,priority:1" json:"store_id"`
	Code      string    `gorm:"size:32;not null;uniqueIndex:ux_zone_store_code,priority:2" json:"code"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (z *StoreStockZone) BeforeCreate(tx *gorm.DB) error {
	if z.ID == uuid.Nil {
		z.ID = uuid.New()
	}
	return nil
}

// StockBatch партия товара (FEFO по expires_at, затем по received_at).
type StockBatch struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StoreID     uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	ProductID   uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	ZoneID      uuid.UUID  `gorm:"type:uuid;not null" json:"zone_id"`
	Quantity    int        `gorm:"not null;default:0" json:"quantity"`
	ReceivedAt  time.Time  `gorm:"not null" json:"received_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	SupplierID  *uuid.UUID `gorm:"type:uuid" json:"supplier_id,omitempty"`
	Note        string     `gorm:"type:text" json:"note,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	Zone        *StoreStockZone `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
}

func (b *StockBatch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (StockBatch) TableName() string { return "stock_batches" }

// StockMovementType движение склада.
type StockMovementType string

const (
	MovementReceipt       StockMovementType = "receipt"
	MovementSale          StockMovementType = "sale"
	MovementWriteOffDamage StockMovementType = "write_off_damage"
	MovementWriteOffShrink StockMovementType = "write_off_shrink"
	MovementWriteOffResort StockMovementType = "write_off_resort"
	MovementAdjustment    StockMovementType = "adjustment"
	MovementAudit         StockMovementType = "audit_adjustment"
)

// StockMovement журнал движений.
type StockMovement struct {
	ID           uuid.UUID         `gorm:"type:uuid;primary_key" json:"id"`
	StoreID      uuid.UUID         `gorm:"type:uuid;not null" json:"store_id"`
	ProductID    uuid.UUID         `gorm:"type:uuid;not null" json:"product_id"`
	BatchID      *uuid.UUID        `gorm:"type:uuid" json:"batch_id,omitempty"`
	ZoneID       *uuid.UUID        `gorm:"type:uuid" json:"zone_id,omitempty"`
	Delta        int               `gorm:"not null" json:"delta"`
	MovementType StockMovementType `gorm:"type:varchar(32);not null" json:"movement_type"`
	RefOrderID   *uuid.UUID        `gorm:"type:uuid" json:"ref_order_id,omitempty"`
	Reason       string            `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

func (m *StockMovement) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (StockMovement) TableName() string { return "stock_movements" }

// Supplier поставщик магазина.
type Supplier struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StoreID   uuid.UUID `gorm:"type:uuid;not null" json:"store_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Phone     string    `gorm:"size:20" json:"phone"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// StockReceipt приходная накладная (агрегат).
type StockReceipt struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StoreID    uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	SupplierID *uuid.UUID `gorm:"type:uuid" json:"supplier_id,omitempty"`
	Note       string     `gorm:"type:text" json:"note,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	Lines      []StockReceiptLine `gorm:"foreignKey:ReceiptID" json:"lines,omitempty"`
}

func (r *StockReceipt) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (StockReceipt) TableName() string { return "stock_receipts" }

// StockReceiptLine строка прихода.
type StockReceiptLine struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	ReceiptID  uuid.UUID  `gorm:"type:uuid;not null" json:"receipt_id"`
	ProductID  uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	ZoneID     uuid.UUID  `gorm:"type:uuid;not null" json:"zone_id"`
	Quantity   int        `gorm:"not null" json:"quantity"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (l *StockReceiptLine) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func (StockReceiptLine) TableName() string { return "stock_receipt_lines" }

// InventoryAuditSession сессия инвентаризации.
type InventoryAuditSession struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StoreID     uuid.UUID  `gorm:"type:uuid;not null" json:"store_id"`
	Note        string     `gorm:"type:text" json:"note,omitempty"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Lines       []InventoryAuditLine `gorm:"foreignKey:SessionID" json:"lines,omitempty"`
}

func (a *InventoryAuditSession) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (InventoryAuditSession) TableName() string { return "inventory_audit_sessions" }

// InventoryAuditLine строка пересчёта.
type InventoryAuditLine struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	SessionID         uuid.UUID  `gorm:"type:uuid;not null" json:"session_id"`
	ProductID         uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	ZoneID            *uuid.UUID `gorm:"type:uuid" json:"zone_id,omitempty"`
	CountedQty        int        `gorm:"not null" json:"counted_qty"`
	SystemQtySnapshot int        `gorm:"not null" json:"system_qty_snapshot"`
	DiffQty           int        `gorm:"not null" json:"diff_qty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (l *InventoryAuditLine) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func (InventoryAuditLine) TableName() string { return "inventory_audit_lines" }
