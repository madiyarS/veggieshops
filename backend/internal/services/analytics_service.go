package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/repositories"
)

type AnalyticsService struct {
	orderRepo repositories.OrderRepository
}

func NewAnalyticsService(orderRepo repositories.OrderRepository) *AnalyticsService {
	return &AnalyticsService{orderRepo: orderRepo}
}

// AdminRevenueSummaryDTO сводка по выручке (заказы кроме отменённых).
type AdminRevenueSummaryDTO struct {
	TotalRevenue      int64  `json:"total_revenue"`
	OrdersCount       int64  `json:"orders_count"`
	AverageCheck      int64  `json:"average_check"`
	TotalDeliveryFees int64  `json:"total_delivery_fees"`
	DateFrom          string `json:"date_from"`
	DateTo            string `json:"date_to"`
}

// AdminRevenueReport ответ для админки: сводка + по дням.
type AdminRevenueReport struct {
	Summary *AdminRevenueSummaryDTO           `json:"summary"`
	ByDay   []repositories.AdminRevenueDayRow `json:"by_day"`
}

func resolveDateRange(fromStr, toStr *string) (from, to string) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	toTime := today
	if toStr != nil && *toStr != "" {
		if t, err := time.Parse("2006-01-02", *toStr); err == nil {
			toTime = t
		}
	}

	fromTime := toTime.AddDate(0, 0, -29)
	if fromStr != nil && *fromStr != "" {
		if t, err := time.Parse("2006-01-02", *fromStr); err == nil {
			fromTime = t
		}
	}

	if fromTime.After(toTime) {
		fromTime, toTime = toTime, fromTime
	}

	return fromTime.Format("2006-01-02"), toTime.Format("2006-01-02")
}

func (s *AnalyticsService) AdminRevenueReport(ctx context.Context, storeID *uuid.UUID, fromStr, toStr *string) (*AdminRevenueReport, error) {
	from, to := resolveDateRange(fromStr, toStr)
	fromPtr, toPtr := from, to

	raw, err := s.orderRepo.AdminRevenueSummary(ctx, storeID, &fromPtr, &toPtr)
	if err != nil {
		return nil, err
	}

	avg := int64(0)
	if raw.OrdersCount > 0 {
		avg = raw.TotalRevenue / raw.OrdersCount
	}

	summary := &AdminRevenueSummaryDTO{
		TotalRevenue:      raw.TotalRevenue,
		OrdersCount:       raw.OrdersCount,
		AverageCheck:      avg,
		TotalDeliveryFees: raw.TotalDeliveryFee,
		DateFrom:          from,
		DateTo:            to,
	}

	byDay, err := s.orderRepo.AdminRevenueByDay(ctx, storeID, &fromPtr, &toPtr)
	if err != nil {
		return nil, err
	}

	return &AdminRevenueReport{Summary: summary, ByDay: byDay}, nil
}
