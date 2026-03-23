package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/veggieshop/backend/internal/services"
)

// StartStalePendingSweep периодически отменяет устаревшие pending-заказы (снятие резерва).
func StartStalePendingSweep(ctx context.Context, orderSvc *services.OrderService, interval, pendingOlderThan time.Duration) {
	if orderSvc == nil || interval <= 0 || pendingOlderThan <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-pendingOlderThan)
				n, err := orderSvc.RunStalePendingSweep(context.Background(), cutoff)
				if err != nil {
					slog.Warn("stale_pending_sweep_failed", "error", err)
					continue
				}
				if n > 0 {
					slog.Info("stale_pending_sweep", "cancelled_orders", n)
				}
			}
		}
	}()
}
