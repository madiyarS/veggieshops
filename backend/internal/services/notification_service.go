package services

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
)

type NotificationService struct {
	notifRepo repositories.NotificationRepository
	whatsappKey string
}

func NewNotificationService(nr repositories.NotificationRepository, whatsappKey string) *NotificationService {
	return &NotificationService{
		notifRepo:   nr,
		whatsappKey: whatsappKey,
	}
}

func (s *NotificationService) SendWhatsAppNotification(ctx context.Context, phone, message string) error {
	// TODO: Integrate with Twilio/WhatsApp Business API
	msgPreview := message
	if len(message) > 50 {
		msgPreview = message[:50] + "..."
	}
	slog.Info("WhatsApp notification", "phone", phone, "message", msgPreview)
	if s.whatsappKey == "" {
		return nil
	}
	return nil
}

func (s *NotificationService) SendOrderNotification(ctx context.Context, orderID uuid.UUID, userID *uuid.UUID, message string) error {
	notif := &models.Notification{
		OrderID: orderID,
		UserID:  userID,
		Channel: models.ChannelWhatsApp,
		Status:  models.NotifPending,
		Message: message,
	}
	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return err
	}
	phone := "" // Get from order
	if phone != "" {
		_ = s.SendWhatsAppNotification(ctx, phone, message)
	}
	return nil
}

func (s *NotificationService) GetNotificationHistory(ctx context.Context, orderID uuid.UUID) ([]*models.Notification, error) {
	return s.notifRepo.GetByOrderID(ctx, orderID)
}
