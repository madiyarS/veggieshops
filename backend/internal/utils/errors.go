package utils

import "errors"

var (
	ErrNotFound           = errors.New("не найдено")
	ErrUnauthorized       = errors.New("требуется авторизация")
	ErrForbidden          = errors.New("доступ запрещен")
	ErrInvalidInput       = errors.New("неверные данные")
	ErrDeliveryUnavailable = errors.New("доставка в ваш район недоступна")
	ErrMinOrderAmount     = errors.New("минимальная сумма заказа не достигнута")
	ErrTimeSlotFull       = errors.New("временное окно недоступно")
	ErrInsufficientStock   = errors.New("недостаточно товара на складе")
	ErrWrongDeliveryCode   = errors.New("неверный код доставки")
	ErrOrderNotInDelivery  = errors.New("заказ не в статусе «в доставке»")
	ErrCourierProfile        = errors.New("профиль курьера не найден")
	ErrDeliveredOnlyViaCode  = errors.New("статус «доставлен» только через код у курьера")
)
