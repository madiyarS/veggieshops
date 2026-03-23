package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "phone = ?", phone).Error
	if err == nil {
		return &user, nil
	}
	// Fallback: match by digits only (для +77000000000 и 77000000000)
	var sb strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			sb.WriteRune(c)
		}
	}
	digits := sb.String()
	if len(digits) >= 11 && digits[0] == '8' {
		digits = "7" + digits[1:]
	}
	if len(digits) == 10 {
		digits = "7" + digits
	}
	err = r.db.WithContext(ctx).Raw(
		"SELECT * FROM users WHERE REGEXP_REPLACE(phone, '[^0-9]', '', 'g') = ? LIMIT 1",
		digits,
	).Scan(&user).Error
	if err != nil || user.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}
