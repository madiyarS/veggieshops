import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCartStore } from '../store/useCartStore';
import { useAuthStore } from '../store/useAuthStore';
import { ordersAPI, storesAPI } from '../services/api';
import { formatApiErrorForUi } from '../utils/apiError';

type Step = 1 | 2 | 3 | 4 | 5;

/** Дата YYYY-MM-DD в локальном календаре браузера (без сдвига UTC). */
function formatLocalYMD(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function formatSlotTime(t: string) {
  if (!t) return '';
  return t.length >= 5 ? t.slice(0, 5) : t;
}

export function Checkout() {
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const items = useCartStore((s) => s.items);
  const cartTotal = useCartStore((s) => s.total());
  const clear = useCartStore((s) => s.clear);
  const cartStoreId = useCartStore((s) => s.storeId);
  const cartStoreName = useCartStore((s) => s.storeName);
  const cartMinOrder = useCartStore((s) => s.minOrderAmount);
  const meetsMinOrder = useCartStore((s) => s.meetsMinOrder);
  const [step, setStep] = useState<Step>(1);
  const [storeId, setStoreId] = useState('');
  const [deliveryType, setDeliveryType] = useState<'regular' | 'express'>('regular');
  const [districtId, setDistrictId] = useState('');
  const [districts, setDistricts] = useState<
    { id: string; name: string; delivery_fee_regular: number; delivery_fee_express: number }[]
  >([]);
  const [address, setAddress] = useState('');
  const [timeSlotId, setTimeSlotId] = useState('');
  const [timeSlots, setTimeSlots] = useState<
    { id: string; start_time: string; end_time: string; available_slots: number }[]
  >([]);
  const [date, setDate] = useState('');
  const [slotsLoading, setSlotsLoading] = useState(false);
  const [slotsError, setSlotsError] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<'kaspi' | 'halyk' | 'cash'>('cash');
  const [customerName, setCustomerName] = useState('');
  const [customerPhone, setCustomerPhone] = useState('');
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const defaultDeliveryDate = formatLocalYMD(tomorrow);
  const todayStr = formatLocalYMD(new Date());

  useEffect(() => {
    if (items.length === 0 && step === 1) navigate('/cart');
  }, [items, step, navigate]);

  useEffect(() => {
    if (items.length > 0 && cartStoreId) {
      setStoreId(cartStoreId);
    }
  }, [items.length, cartStoreId]);

  useEffect(() => {
    if (storeId) storesAPI.getDistricts(storeId).then((r) => setDistricts((r.data as { data?: typeof districts })?.data || []));
  }, [storeId]);

  useEffect(() => {
    if (!user) return;
    const name = [user.first_name, user.last_name].filter(Boolean).join(' ').trim();
    setCustomerName((prev) => (prev.trim() ? prev : name));
    setCustomerPhone((prev) => (prev.trim() ? prev : user.phone || ''));
  }, [user]);

  useEffect(() => {
    if (step !== 4 || deliveryType !== 'regular' || !storeId || !date) return;
    let cancelled = false;
    setSlotsLoading(true);
    setSlotsError('');
    storesAPI
      .getTimeSlots(storeId, date)
      .then((r) => {
        if (cancelled) return;
        const list = (r.data as { data?: typeof timeSlots })?.data || [];
        setTimeSlots(list);
        setTimeSlotId((prev) => (list.some((s) => s.id === prev) ? prev : ''));
      })
      .catch(() => {
        if (!cancelled) setSlotsError('Не удалось загрузить окна доставки. Проверьте дату и попробуйте снова.');
      })
      .finally(() => {
        if (!cancelled) setSlotsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [step, storeId, date, deliveryType]);

  const deliveryFee = districtId
    ? (deliveryType === 'express'
        ? districts.find((d) => d.id === districtId)?.delivery_fee_express
        : districts.find((d) => d.id === districtId)?.delivery_fee_regular) || 0
    : 0;
  const orderTotal = cartTotal + deliveryFee;
  const minShort = cartMinOrder > 0 ? Math.max(0, cartMinOrder - cartTotal) : 0;

  const selectedSlotLabel = timeSlots.find((s) => s.id === timeSlotId);

  const changeStoreAndRestart = () => {
    if (
      window.confirm(
        'Сменить магазин? Корзина будет очищена — нужно снова выбрать магазин на главной и добавить товары.'
      )
    ) {
      clear();
      navigate('/');
    }
  };

  const handleSubmit = async () => {
    if (!meetsMinOrder()) {
      setError(`Минимальная сумма товаров для этого магазина — ${cartMinOrder} ₸ (сейчас ${cartTotal} ₸)`);
      return;
    }
    if (!storeId || !districtId || !address || !customerName.trim() || !customerPhone.trim()) {
      setError('Заполните все поля');
      return;
    }
    if (deliveryType === 'regular' && !timeSlotId) {
      setError('Выберите дату и время доставки');
      return;
    }
    setLoading(true);
    setError('');
    try {
      let slotId = timeSlotId;
      if (deliveryType === 'express' && !slotId) {
        const r = await storesAPI.getTimeSlots(storeId, todayStr);
        const slots = (r.data as { data?: { id: string }[] })?.data || [];
        slotId = slots[0]?.id || '';
      }
      if (!slotId) {
        setError('Нет доступных слотов на выбранный день. Выберите другую дату или тип доставки.');
        setLoading(false);
        return;
      }
      const res = await ordersAPI.create({
        store_id: storeId,
        district_id: districtId,
        delivery_type: deliveryType,
        delivery_time_slot_id: slotId,
        delivery_address: address,
        customer_phone: customerPhone.trim(),
        customer_name: customerName.trim(),
        payment_method: paymentMethod,
        items: items.map((i) => ({ product_id: i.productId, quantity: i.quantity })),
        notes,
      });
      const data = (res.data as { data?: { id: string } })?.data;
      clear();
      if (data?.id) navigate(`/orders/${data.id}`);
      else navigate('/orders');
    } catch (e: unknown) {
      setError(formatApiErrorForUi(e));
    } finally {
      setLoading(false);
    }
  };

  const goBackFromPayment = () => {
    if (deliveryType === 'express') setStep(3);
    else setStep(4);
  };

  const stepProgress =
    deliveryType === 'express'
      ? step <= 3
        ? `Шаг ${step} из 4`
        : 'Шаг 4 из 4'
      : `Шаг ${step} из 5`;

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-veggie-green mb-2">Оформление заказа</h2>
      <p className="text-sm text-gray-500 mb-4">{stepProgress}</p>
      {error && <p className="text-red-600 mb-4">{error}</p>}

      {step === 1 && (
        <div className="bg-white p-6 rounded shadow">
          <h3 className="font-semibold mb-4">1. Магазин и тип доставки</h3>
          <div className="rounded-lg border border-gray-200 bg-gray-50 p-4 mb-4">
            <p className="text-sm text-gray-600">Заказ из корзины</p>
            <p className="text-lg font-semibold text-gray-900 mt-1">{cartStoreName || 'Магазин'}</p>
            {cartMinOrder > 0 && (
              <div className="mt-2 text-sm text-gray-700">
                <p>
                  Мин. сумма по товарам (без доставки): <span className="font-medium">{cartMinOrder} ₸</span>
                </p>
                <p className="mt-1">
                  В корзине сейчас: <span className="font-medium">{cartTotal} ₸</span>
                  {minShort > 0 ? (
                    <span className="text-amber-700 ml-2">— не хватает {minShort} ₸</span>
                  ) : (
                    <span className="text-green-700 ml-2">✓</span>
                  )}
                </p>
              </div>
            )}
            <button
              type="button"
              onClick={changeStoreAndRestart}
              className="mt-3 text-sm text-veggie-green font-medium hover:underline"
            >
              Сменить магазин (очистить корзину и вернуться на главную)
            </button>
          </div>
          <p className="text-sm text-gray-600 mb-3">Тип доставки</p>
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                checked={deliveryType === 'regular'}
                onChange={() => setDeliveryType('regular')}
              />
              <span>Обычная (окно 2–3 ч — выберите дату и время на следующем шаге)</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                checked={deliveryType === 'express'}
                onChange={() => setDeliveryType('express')}
              />
              <span>Экспресс (30–60 мин, ближайшее окно)</span>
            </label>
          </div>
          <button
            onClick={() => setStep(2)}
            className="bg-veggie-green text-white px-6 py-2 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!storeId || !meetsMinOrder()}
          >
            Далее
          </button>
          {!meetsMinOrder() && cartMinOrder > 0 && (
            <p className="text-sm text-amber-700 mt-2">Добавьте товаров в корзине до минимума, затем вернитесь к оформлению.</p>
          )}
        </div>
      )}

      {step === 2 && (
        <div className="bg-white p-6 rounded shadow">
          <h3 className="font-semibold mb-4">2. Район доставки</h3>
          <select value={districtId} onChange={(e) => setDistrictId(e.target.value)} className="w-full border p-2 rounded mb-6">
            <option value="">-- Район --</option>
            {districts.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name} ({deliveryType === 'express' ? d.delivery_fee_express : d.delivery_fee_regular} ₸)
              </option>
            ))}
          </select>
          <button type="button" onClick={() => setStep(1)} className="mr-2 text-gray-600">
            Назад
          </button>
          <button onClick={() => setStep(3)} className="bg-veggie-green text-white px-6 py-2 rounded" disabled={!districtId}>
            Далее
          </button>
        </div>
      )}

      {step === 3 && (
        <div className="bg-white p-6 rounded shadow">
          <h3 className="font-semibold mb-4">3. Адрес и контакты</h3>
          {!user?.phone && (
            <p className="text-sm text-gray-600 mb-3">
              <Link to="/login?next=/checkout" className="text-veggie-green font-medium hover:underline">
                Войдите
              </Link>
              {' · '}
              <Link to="/register?next=/checkout" className="text-veggie-green font-medium hover:underline">
                Регистрация
              </Link>
              <span className="text-gray-500"> — подставим имя и телефон автоматически.</span>
            </p>
          )}
          <input
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="Улица, дом, квартира"
            className="w-full border p-2 rounded mb-4"
          />
          <input
            value={customerName}
            onChange={(e) => setCustomerName(e.target.value)}
            placeholder="Ваше имя"
            className="w-full border p-2 rounded mb-4"
          />
          <input
            value={customerPhone}
            onChange={(e) => setCustomerPhone(e.target.value)}
            placeholder="+7 XXX XXX XX XX"
            className="w-full border p-2 rounded mb-6"
          />
          <button type="button" onClick={() => setStep(2)} className="mr-2 text-gray-600">
            Назад
          </button>
          <button
            onClick={async () => {
              if (deliveryType === 'express') {
                setDate(todayStr);
                const res = await storesAPI.getTimeSlots(storeId, todayStr);
                const slots = (res.data as { data?: typeof timeSlots })?.data || [];
                if (slots.length > 0) {
                  setTimeSlotId(slots[0].id);
                  setTimeSlots(slots);
                } else {
                  setTimeSlotId('');
                  setTimeSlots([]);
                }
                setStep(5);
              } else {
                setDate(defaultDeliveryDate);
                setTimeSlotId('');
                setTimeSlots([]);
                setStep(4);
              }
            }}
            className="bg-veggie-green text-white px-6 py-2 rounded"
            disabled={!address.trim() || !customerName.trim() || !customerPhone.trim()}
          >
            Далее
          </button>
        </div>
      )}

      {step === 4 && deliveryType === 'regular' && (
        <div className="bg-white p-6 rounded shadow">
          <h3 className="font-semibold mb-2">4. Дата и время доставки</h3>
          <p className="text-sm text-gray-600 mb-4">Для обычной доставки нужно выбрать день и слот.</p>
          <label className="block text-sm text-gray-700 mb-1">Дата</label>
          <input
            type="date"
            min={todayStr}
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="w-full border p-2 rounded mb-4"
          />
          {slotsLoading && <p className="text-sm text-gray-500 mb-2">Загружаем доступные окна…</p>}
          {slotsError && <p className="text-red-600 text-sm mb-2">{slotsError}</p>}
          {!slotsLoading && !slotsError && timeSlots.length === 0 && date && (
            <p className="text-amber-700 text-sm mb-2">На этот день нет окон — выберите другую дату.</p>
          )}
          <label className="block text-sm text-gray-700 mb-1">Время</label>
          <select
            value={timeSlotId}
            onChange={(e) => setTimeSlotId(e.target.value)}
            className="w-full border p-2 rounded mb-6"
            disabled={slotsLoading || timeSlots.length === 0}
          >
            <option value="">-- Выберите окно --</option>
            {timeSlots.map((s) => (
              <option key={s.id} value={s.id}>
                {formatSlotTime(s.start_time)} – {formatSlotTime(s.end_time)} (свободно: {s.available_slots})
              </option>
            ))}
          </select>
          <button type="button" onClick={() => setStep(3)} className="mr-2 text-gray-600">
            Назад
          </button>
          <button
            onClick={() => setStep(5)}
            className="bg-veggie-green text-white px-6 py-2 rounded"
            disabled={!timeSlotId || slotsLoading}
          >
            Далее
          </button>
        </div>
      )}

      {step === 5 && (
        <div className="bg-white p-6 rounded shadow">
          <h3 className="font-semibold mb-4">{deliveryType === 'express' ? '4' : '5'}. Способ оплаты и проверка</h3>
          {deliveryType === 'regular' && (
            <div className="rounded-lg bg-gray-50 border border-gray-100 p-3 mb-4 text-sm">
              <p className="font-medium text-gray-800">Доставка</p>
              <p className="text-gray-600">
                {date} ·{' '}
                {selectedSlotLabel
                  ? `${formatSlotTime(selectedSlotLabel.start_time)} – ${formatSlotTime(selectedSlotLabel.end_time)}`
                  : 'время не выбрано'}
              </p>
            </div>
          )}
          <div className="space-y-2 mb-4">
            <label className="block">
              <input type="radio" checked={paymentMethod === 'kaspi'} onChange={() => setPaymentMethod('kaspi')} /> Kaspi.kz
            </label>
            <label className="block">
              <input type="radio" checked={paymentMethod === 'halyk'} onChange={() => setPaymentMethod('halyk')} /> Halyk Bank
            </label>
            <label className="block">
              <input type="radio" checked={paymentMethod === 'cash'} onChange={() => setPaymentMethod('cash')} /> Наличные
            </label>
          </div>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Комментарий"
            className="w-full border p-2 rounded mb-4"
          />
          <div className="border-t pt-4 mb-6">
            <p>Товары: {cartTotal} ₸</p>
            <p>Доставка: {deliveryFee} ₸</p>
            <p className="font-bold text-lg">Итого: {orderTotal} ₸</p>
          </div>
          <button type="button" onClick={goBackFromPayment} className="mr-2 text-gray-600">
            Назад
          </button>
          <button onClick={handleSubmit} disabled={loading} className="bg-veggie-green text-white px-6 py-2 rounded">
            {loading ? 'Оформляем...' : 'Оформить заказ'}
          </button>
        </div>
      )}
    </div>
  );
}
