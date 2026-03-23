import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { ordersAPI, productsAPI, storesAPI } from '../services/api';
import { useCartStore, type InventoryUnitClient } from '../store/useCartStore';

const statusLabels: Record<string, string> = {
  pending: 'Приняли заказ',
  confirmed: 'Собираем ваш заказ',
  preparing: 'Собираем ваш заказ',
  in_delivery: 'Курьер уже в пути',
  delivered: 'Доставлен',
  cancelled: 'Отменён',
};

const STATUS_FLOW = ['pending', 'confirmed', 'preparing', 'in_delivery', 'delivered'] as const;
const NOTIFY_STORAGE_KEY = 'veggieshops_kz_order_notify';

type OrderItemRow = {
  product_id: string;
  quantity: number;
  price_at_order: number;
  subtotal: number;
};

type OrderRow = {
  id: string;
  order_number: string;
  store_id: string;
  status: string;
  total_amount: number;
  delivery_address: string;
  customer_name: string;
  delivery_code?: string;
  created_at?: string;
  items?: OrderItemRow[];
};

function formatDate(iso?: string) {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function MyOrdersList() {
  const [orders, setOrders] = useState<OrderRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(() => {
    setError('');
    return ordersAPI
      .listMine()
      .then((r) => setOrders((r.data as { data?: OrderRow[] })?.data || []))
      .catch(() => setError('Не удалось загрузить заказы'));
  }, []);

  useEffect(() => {
    load().finally(() => setLoading(false));
  }, [load]);

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-veggie-green mb-2">Мои заказы</h2>
      <p className="text-sm text-gray-500 mb-6">История и статус доставки</p>
      {error && <p className="text-red-600 mb-4">{error}</p>}
      {loading && <p className="text-gray-600">Загрузка...</p>}
      {!loading && !error && orders.length === 0 && (
        <div className="bg-white rounded-lg shadow p-8 text-center text-gray-600">
          <p>Пока нет заказов.</p>
          <Link to="/" className="inline-block mt-4 text-veggie-green font-medium hover:underline">
            На главную
          </Link>
        </div>
      )}
      <ul className="space-y-3">
        {orders.map((o) => (
          <li key={o.id}>
            <Link
              to={`/orders/${o.id}`}
              className="block bg-white rounded-lg shadow border border-gray-100 p-4 hover:border-veggie-green/40 transition-colors"
            >
              <div className="flex justify-between gap-2 flex-wrap">
                <span className="font-semibold text-gray-900">{o.order_number}</span>
                <span className="text-veggie-green">{statusLabels[o.status] || o.status}</span>
              </div>
              <p className="text-sm text-gray-500 mt-1">{formatDate(o.created_at)}</p>
              <p className="text-sm mt-2 text-gray-800">{o.total_amount} ₸</p>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function MyOrderDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const addItem = useCartStore((s) => s.addItem);
  const clear = useCartStore((s) => s.clear);
  const cartStoreId = useCartStore((s) => s.storeId);
  const cartItems = useCartStore((s) => s.items);

  const [order, setOrder] = useState<OrderRow | null>(null);
  const [names, setNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [repeatBusy, setRepeatBusy] = useState(false);
  const [repeatMsg, setRepeatMsg] = useState('');
  const [notifyOn, setNotifyOn] = useState(() => localStorage.getItem(NOTIFY_STORAGE_KEY) === '1');

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError('');
    ordersAPI
      .getMine(id)
      .then((r) => setOrder((r.data as { data?: OrderRow })?.data || null))
      .catch(() => setError('Заказ не найден'))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    if (!id || !order) return;
    if (order.status === 'delivered' || order.status === 'cancelled') return;
    let cancelled = false;
    const tick = async () => {
      if (cancelled) return;
      try {
        const r = await ordersAPI.getMine(id);
        const next = (r.data as { data?: OrderRow })?.data ?? null;
        if (!next || cancelled) return;
        setOrder((prev) => {
          if (!prev) return next;
          const merged = { ...prev, ...next, items: next.items?.length ? next.items : prev.items };
          if (prev.status !== next.status && localStorage.getItem(NOTIFY_STORAGE_KEY) === '1') {
            if (typeof Notification !== 'undefined' && Notification.permission === 'granted') {
              new Notification(`Заказ ${next.order_number}`, {
                body: statusLabels[next.status] || next.status,
              });
            }
          }
          return merged;
        });
      } catch {
        /* сеть */
      }
    };
    const t = setInterval(tick, 30000);
    return () => {
      cancelled = true;
      clearInterval(t);
    };
  }, [id, order?.id, order?.status]);

  useEffect(() => {
    const items = order?.items;
    if (!items?.length) return;
    const ids = [...new Set(items.map((i) => i.product_id))];
    let cancelled = false;
    Promise.all(
      ids.map((pid) =>
        productsAPI.getById(pid).then((r) => {
          const p = (r.data as { data?: { id: string; name?: string } })?.data;
          return p ? ([p.id, p.name || 'Товар'] as const) : null;
        })
      )
    ).then((pairs) => {
      if (cancelled) return;
      const m: Record<string, string> = {};
      for (const p of pairs) {
        if (p) m[p[0]] = p[1];
      }
      setNames(m);
    });
    return () => {
      cancelled = true;
    };
  }, [order]);

  const repeatOrder = async () => {
    if (!order?.items?.length || !order.store_id) return;
    setRepeatMsg('');
    setRepeatBusy(true);
    try {
      if (cartItems.length > 0 && cartStoreId && cartStoreId !== order.store_id) {
        if (
          !window.confirm(
            'В корзине товары другого магазина. Очистить корзину и собрать заказ заново из этого?'
          )
        ) {
          setRepeatBusy(false);
          return;
        }
        clear();
      } else if (cartItems.length > 0) {
        if (!window.confirm('Добавить позиции к текущей корзине?')) {
          setRepeatBusy(false);
          return;
        }
      }

      const storeRes = await storesAPI.getById(order.store_id);
      const sd = (storeRes.data as { data?: { name?: string; min_order_amount?: number } })?.data;
      const storeName = sd?.name || 'Магазин';
      const minOrderAmount = typeof sd?.min_order_amount === 'number' ? sd.min_order_amount : 0;

      let added = 0;
      let skipped = 0;

      for (const line of order.items) {
        try {
          const pr = await productsAPI.getById(line.product_id);
          const p = (pr.data as {
            data?: {
              id: string;
              name: string;
              price: number;
              unit: string;
              stock_quantity: number;
              inventory_unit?: InventoryUnitClient;
              cart_step_grams?: number;
              is_available?: boolean;
              is_active?: boolean;
              temporarily_unavailable?: boolean;
            };
          })?.data;
          if (
            !p ||
            p.is_active === false ||
            p.is_available === false ||
            p.temporarily_unavailable ||
            p.stock_quantity <= 0
          ) {
            skipped++;
            continue;
          }
          const inv: InventoryUnitClient = p.inventory_unit === 'weight_gram' ? 'weight_gram' : 'piece';
          let qty = line.quantity;
          if (inv === 'weight_gram') {
            qty = Math.min(qty, p.stock_quantity);
            if (qty < 1) {
              skipped++;
              continue;
            }
            qty = Math.round(qty);
          } else {
            qty = Math.min(qty, Math.floor(p.stock_quantity));
            if (qty < 1) {
              skipped++;
              continue;
            }
          }
          addItem({
            productId: p.id,
            name: p.name,
            price: p.price,
            unit: inv === 'weight_gram' ? 'кг' : p.unit,
            inventoryUnit: inv,
            cartStepGrams: p.cart_step_grams || 250,
            quantity: qty,
            storeId: order.store_id,
            storeName,
            minOrderAmount,
          });
          added++;
        } catch {
          skipped++;
        }
      }

      if (added === 0) {
        setRepeatMsg('Не удалось добавить позиции: товары недоступны или закончились.');
      } else {
        if (skipped > 0) {
          setRepeatMsg(`В корзину добавлено позиций: ${added}. Некоторые товары пропущены (${skipped}).`);
        }
        navigate('/cart');
      }
    } catch {
      setRepeatMsg('Ошибка при повторе заказа');
    } finally {
      setRepeatBusy(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-xl mx-auto">
        <p className="text-gray-600">Загрузка...</p>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="max-w-xl mx-auto">
        <p className="text-red-600">{error || 'Заказ не найден'}</p>
        <Link to="/orders" className="inline-block mt-4 text-veggie-green hover:underline">
          К списку заказов
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto">
      <Link to="/orders" className="text-sm text-veggie-green hover:underline mb-4 inline-block">
        ← Все заказы
      </Link>
      <h2 className="text-2xl font-bold text-veggie-green mb-2">Заказ {order.order_number}</h2>
      <p className="text-sm text-gray-500 mb-4">{formatDate(order.created_at)}</p>

      <div className="mb-6 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
        <label className="flex cursor-pointer items-start gap-3 text-sm text-gray-700">
          <input
            type="checkbox"
            checked={notifyOn}
            onChange={async (e) => {
              const on = e.target.checked;
              setNotifyOn(on);
              localStorage.setItem(NOTIFY_STORAGE_KEY, on ? '1' : '0');
              if (on && typeof Notification !== 'undefined' && Notification.permission === 'default') {
                await Notification.requestPermission();
              }
            }}
            className="mt-1"
          />
          <span>
            Уведомлять об изменении статуса в браузере (нужно разрешение на уведомления). Страница сама
            обновляет статус примерно раз в 30 сек.
          </span>
        </label>
      </div>

      {order.status === 'cancelled' ? (
        <p className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-800">
          Заказ отменён
        </p>
      ) : (
        <div className="mb-6 overflow-x-auto rounded-lg border border-gray-100 bg-white p-4 shadow-sm">
          <p className="mb-3 text-xs font-medium uppercase tracking-wide text-gray-500">Статус заказа</p>
          <ol className="flex min-w-[520px] items-start justify-between gap-1">
            {[
              { key: 'pending', short: 'Принят' },
              { key: 'confirmed', short: 'Подтв.' },
              { key: 'preparing', short: 'Сборка' },
              { key: 'in_delivery', short: 'В пути' },
              { key: 'delivered', short: 'У вас' },
            ].map((step, i) => {
              const idx = STATUS_FLOW.indexOf(order.status as (typeof STATUS_FLOW)[number]);
              const stepIdx = STATUS_FLOW.indexOf(step.key as (typeof STATUS_FLOW)[number]);
              const done = idx >= 0 && stepIdx <= idx;
              const current = idx === stepIdx;
              return (
                <li key={step.key} className="flex flex-1 flex-col items-center text-center">
                  <div
                    className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                      done
                        ? 'bg-veggie-green text-white'
                        : 'border-2 border-gray-200 bg-gray-50 text-gray-400'
                    } ${current ? 'ring-2 ring-veggie-green ring-offset-2' : ''}`}
                  >
                    {i + 1}
                  </div>
                  <span className={`mt-2 max-w-[4.5rem] text-[11px] leading-tight sm:text-xs ${done ? 'text-gray-900' : 'text-gray-400'}`}>
                    {step.short}
                  </span>
                </li>
              );
            })}
          </ol>
        </div>
      )}

      <div className="bg-white p-6 rounded shadow space-y-2 mb-6">
        <p className="text-veggie-green text-lg font-medium">{statusLabels[order.status] || order.status}</p>
        <p>Сумма: {order.total_amount} ₸</p>
        <p>Адрес: {order.delivery_address}</p>
        <p>Получатель: {order.customer_name}</p>
        {order.status !== 'delivered' && order.status !== 'cancelled' && order.delivery_code && (
          <div className="mt-4 p-4 rounded-lg bg-amber-50 border border-amber-200">
            <p className="text-sm font-medium text-amber-900">Код для курьера</p>
            <p className="text-2xl font-mono font-bold tracking-widest text-amber-950 mt-1">
              {order.delivery_code}
            </p>
            <p className="text-xs text-amber-800 mt-2">
              Назовите этот код курьеру при получении. Без кода заказ нельзя отметить доставленным.
            </p>
          </div>
        )}
      </div>

      {order.items && order.items.length > 0 && (
        <div className="bg-white p-6 rounded shadow mb-6">
          <h3 className="font-semibold mb-3">Состав заказа</h3>
          <ul className="space-y-2 text-sm">
            {order.items.map((it, idx) => (
              <li key={`${it.product_id}-${idx}`} className="flex justify-between gap-2 border-b border-gray-100 pb-2">
                <span>{names[it.product_id] || `Товар ${it.product_id.slice(0, 8)}…`}</span>
                <span className="text-gray-600 shrink-0">{it.subtotal} ₸</span>
              </li>
            ))}
          </ul>
          <button
            type="button"
            onClick={() => void repeatOrder()}
            disabled={repeatBusy || order.status === 'cancelled'}
            className="mt-4 w-full bg-veggie-green text-white py-2 rounded hover:bg-veggie-dark disabled:opacity-50"
          >
            {repeatBusy ? 'Добавляем…' : 'Повторить заказ'}
          </button>
          {repeatMsg && <p className="text-sm text-amber-800 mt-2">{repeatMsg}</p>}
        </div>
      )}
    </div>
  );
}
