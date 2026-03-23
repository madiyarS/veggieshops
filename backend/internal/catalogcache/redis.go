package catalogcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Store инвалидация каталога через счётчик поколения на магазин (ключи с TTL).
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// New подключение к Redis; при пустом addr возвращает nil (кэш отключён).
func New(addr string) *Store {
	if addr == "" {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Store{rdb: rdb, ttl: 90 * time.Second}
}

func (s *Store) Close() error {
	if s == nil || s.rdb == nil {
		return nil
	}
	return s.rdb.Close()
}

func (s *Store) Bump(ctx context.Context, storeID uuid.UUID) {
	if s == nil || s.rdb == nil {
		return
	}
	_ = s.rdb.Incr(ctx, genKey(storeID)).Err()
}

func genKey(storeID uuid.UUID) string {
	return fmt.Sprintf("catalog:gen:%s", storeID.String())
}

func (s *Store) generation(ctx context.Context, storeID uuid.UUID) int64 {
	if s == nil || s.rdb == nil {
		return 0
	}
	v, err := s.rdb.Get(ctx, genKey(storeID)).Int64()
	if err == redis.Nil || err != nil {
		return 0
	}
	return v
}

func dataKey(storeID uuid.UUID, gen int64, catPart string) string {
	return fmt.Sprintf("catalog:data:%s:%d:%s", storeID.String(), gen, catPart)
}

// GetProductsJSON кэш списка товаров после расчёта доступных остатков.
func (s *Store) GetProductsJSON(ctx context.Context, storeID uuid.UUID, catPart string) ([]byte, bool) {
	if s == nil || s.rdb == nil {
		return nil, false
	}
	gen := s.generation(ctx, storeID)
	b, err := s.rdb.Get(ctx, dataKey(storeID, gen, catPart)).Bytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

func (s *Store) SetProductsJSON(ctx context.Context, storeID uuid.UUID, catPart string, payload interface{}) error {
	if s == nil || s.rdb == nil {
		return nil
	}
	gen := s.generation(ctx, storeID)
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, dataKey(storeID, gen, catPart), raw, s.ttl).Err()
}
