import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCartStore, cartLineTotal } from '../store/useCartStore';
import { productsAPI } from '../services/api';

function formatAvailLabel(
  inventoryUnit: 'piece' | 'weight_gram' | undefined,
  qty: number,
  unit: string
): string {
  if (inventoryUnit === 'weight_gram') {
    const kg = qty / 1000;
    const d = kg < 10 && kg % 1 !== 0 ? 2 : 1;
    return `${kg.toFixed(d)} кг`;
  }
  return `${qty} ${unit}`;
}

export function Cart() {
  const navigate = useNavigate();
  const {
    items,
    removeItem,
    updateQuantity,
    total,
    clear,
    storeName,
    minOrderAmount,
    meetsMinOrder,
    storeId,
    setAvailability,
    availabilityByProductId,
  } = useCartStore();
  const [availErr, setAvailErr] = useState('');

  useEffect(() => {
    if (!storeId || items.length === 0) return;
    const ids = [...new Set(items.map((i) => i.productId))];
    if (ids.length === 0) return;
    let cancelled = false;
    productsAPI
      .getAvailability(storeId, ids)
      .then((r) => {
        if (cancelled) return;
        const data = (r.data as { data?: Record<string, number> })?.data || {};
        setAvailability(data);
        setAvailErr('');
      })
      .catch(() => {
        if (!cancelled) setAvailErr('Не удалось обновить остатки — цифры «доступно» могут быть устаревшими.');
      });
    return () => {
      cancelled = true;
    };
  }, [storeId, [...new Set(items.map((x) => x.productId))].sort().join(','), setAvailability]);

  const sum = total();
  const short = minOrderAmount > 0 ? Math.max(0, minOrderAmount - sum) : 0;
  const canCheckout = items.length > 0 && meetsMinOrder();

  const changeStore = () => {
    if (items.length === 0) {
      navigate('/');
      return;
    }
    if (
      window.confirm(
        'Чтобы выбрать другой магазин, корзина будет очищена. Перейти на главную и выбрать магазин заново?'
      )
    ) {
      clear();
      navigate('/');
    }
  };

  return (
    <div>
      <h2 className="text-2xl font-bold text-veggie-green mb-6">Корзина</h2>
      {items.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-600">Корзина пуста</p>
          <Link to="/" className="text-veggie-green underline mt-2 inline-block">
            Выбрать магазин на главной
          </Link>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="mb-6 pb-4 border-b border-gray-100">
            <p className="text-gray-800">
              <span className="text-gray-500">Магазин:</span>{' '}
              <span className="font-semibold text-veggie-green">{storeName || '—'}</span>
            </p>
            {minOrderAmount > 0 && (
              <div className="mt-2 text-sm">
                <p className="text-gray-600">
                  Минимальная сумма заказа по товарам (без доставки):{' '}
                  <span className="font-medium text-gray-900">{minOrderAmount} ₸</span>
                </p>
                <p className="mt-1">
                  Сейчас в корзине: <span className="font-medium">{sum} ₸</span>
                  {short > 0 ? (
                    <span className="text-amber-700 ml-2">— не хватает ещё {short} ₸ до минимума</span>
                  ) : (
                    <span className="text-green-700 ml-2">— минимум достигнут ✓</span>
                  )}
                </p>
              </div>
            )}
            {availErr && <p className="mt-2 text-sm text-amber-700">{availErr}</p>}
            <button
              type="button"
              onClick={changeStore}
              className="mt-3 text-sm text-gray-600 underline hover:text-veggie-green"
            >
              Сменить магазин (очистит корзину)
            </button>
          </div>

          <ul className="divide-y">
            {items.map((i) => {
              const step = i.inventoryUnit === 'weight_gram' ? i.cartStepGrams || 250 : 1;
              const avail = availabilityByProductId[i.productId];
              const qtyLabel =
                i.inventoryUnit === 'weight_gram'
                  ? `${(i.quantity / 1000).toFixed(i.quantity % 1000 === 0 ? 0 : 2)} кг`
                  : String(i.quantity);
              const priceLabel =
                i.inventoryUnit === 'weight_gram' ? `${i.price} ₸/кг` : `${i.price} ₸/${i.unit}`;
              const over = avail !== undefined && avail >= 0 && i.quantity > avail;
              return (
                <li
                  key={i.productId}
                  className="py-4 flex flex-col sm:flex-row sm:justify-between sm:items-start gap-2"
                >
                  <div className="min-w-0 flex-1">
                    <span className="font-medium">{i.name}</span>
                    <span className="text-gray-600 ml-2">{priceLabel}</span>
                    <p className="text-sm text-gray-500 mt-1">
                      {qtyLabel} · {cartLineTotal(i)} ₸
                    </p>
                    {avail !== undefined && avail >= 0 && (
                      <p className={`text-sm mt-1 ${over ? 'text-amber-800 font-medium' : 'text-gray-600'}`}>
                        Доступно: {formatAvailLabel(i.inventoryUnit, avail, i.unit)}
                        {over && ' — уменьшите количество'}
                      </p>
                    )}
                  </div>
                  <div className="flex flex-col sm:items-end gap-2 shrink-0">
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => updateQuantity(i.productId, i.quantity - step)}
                        className="w-8 h-8 rounded bg-gray-200"
                      >
                        -
                      </button>
                      <span className="min-w-[4rem] text-center text-sm">{qtyLabel}</span>
                      <button
                        type="button"
                        onClick={() => updateQuantity(i.productId, i.quantity + step)}
                        className="w-8 h-8 rounded bg-gray-200"
                      >
                        +
                      </button>
                      <button type="button" onClick={() => removeItem(i.productId)} className="text-red-600 ml-2">
                        ✕
                      </button>
                    </div>
                    {i.inventoryUnit === 'weight_gram' && (
                      <label className="flex items-center gap-2 text-xs text-gray-600 w-full sm:w-auto">
                        <span className="shrink-0">кг</span>
                        <input
                          type="number"
                          inputMode="decimal"
                          min={0.001}
                          step={0.05}
                          max={avail !== undefined && avail >= 0 ? avail / 1000 : undefined}
                          key={`${i.productId}-${i.quantity}`}
                          defaultValue={String(i.quantity / 1000)}
                          onBlur={(e) => {
                            const kg = parseFloat(e.target.value.replace(',', '.'));
                            if (!Number.isFinite(kg) || kg <= 0) return;
                            updateQuantity(i.productId, Math.round(kg * 1000));
                          }}
                          className="w-28 border border-gray-200 rounded px-2 py-1 text-gray-900"
                        />
                      </label>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
          <div className="mt-6 pt-4 border-t flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
            <span className="text-xl font-bold">Итого: {sum} ₸</span>
            <div className="flex flex-col items-stretch sm:items-end gap-2">
              {!canCheckout && minOrderAmount > 0 && (
                <p className="text-sm text-amber-700 text-right">Добавьте товаров на {short} ₸, чтобы оформить заказ</p>
              )}
              <Link
                to="/checkout"
                className={`text-center px-6 py-2 rounded ${
                  canCheckout
                    ? 'bg-veggie-green text-white hover:bg-veggie-dark'
                    : 'bg-gray-200 text-gray-500 pointer-events-none cursor-not-allowed'
                }`}
                aria-disabled={!canCheckout}
                onClick={(e) => {
                  if (!canCheckout) e.preventDefault();
                }}
              >
                Оформить заказ
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
