import { useCallback, useEffect, useState } from 'react';
import { courierAPI } from '../services/api';
import { CourierLayout } from '../components/CourierLayout';

const statusLabels: Record<string, string> = {
  pending: 'В сборке',
  confirmed: 'В сборке',
  preparing: 'Собирают заказ',
  in_delivery: 'Везём клиенту',
  delivered: 'Доставлен',
  cancelled: 'Отменён',
};

function canCourierAccept(status: string) {
  return status === 'pending' || status === 'preparing' || status === 'confirmed';
}

interface OrderRow {
  id: string;
  order_number: string;
  status: string;
  total_amount: number;
  delivery_address: string;
  customer_name: string;
  customer_phone: string;
  courier_id?: string | null;
}

export function CourierPage() {
  const [orders, setOrders] = useState<OrderRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<string | null>(null);
  const [codeByOrder, setCodeByOrder] = useState<Record<string, string>>({});

  const load = useCallback(() => {
    setError('');
    return courierAPI
      .listOrders()
      .then((r) => setOrders((r.data as { data?: OrderRow[] })?.data || []))
      .catch(() => setError('Не удалось загрузить заказы'));
  }, []);

  useEffect(() => {
    load().finally(() => setLoading(false));
  }, [load]);

  const accept = async (id: string) => {
    setBusyId(id);
    setError('');
    try {
      await courierAPI.accept(id);
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка');
    } finally {
      setBusyId(null);
    }
  };

  const complete = async (id: string) => {
    const code = (codeByOrder[id] || '').trim();
    if (!code) {
      setError('Введите код от клиента');
      return;
    }
    setBusyId(id);
    setError('');
    try {
      await courierAPI.complete(id, code);
      setCodeByOrder((m) => ({ ...m, [id]: '' }));
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Неверный код или статус');
    } finally {
      setBusyId(null);
    }
  };

  return (
    <CourierLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Заказы</h1>
          <p className="text-slate-400 text-sm mt-1">
            «Принять в доставку» — заказ у вас, со склада спишется после кода от клиента. «Доставлен» — введите код с
            экрана заказа у клиента.
          </p>
        </div>
        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">{error}</div>
        )}
        {loading ? (
          <p className="text-[var(--admin-text-muted)]">Загрузка...</p>
        ) : orders.length === 0 ? (
          <p className="text-[var(--admin-text-muted)]">Нет активных заказов</p>
        ) : (
          <ul className="space-y-4">
            {orders.map((o) => (
              <li
                key={o.id}
                className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-4 space-y-3"
              >
                <div className="flex flex-wrap justify-between gap-2">
                  <span className="font-semibold text-white">{o.order_number}</span>
                  <span className="text-[var(--admin-accent)]">{o.total_amount} ₸</span>
                </div>
                <p className="text-sm text-[var(--admin-text-muted)]">{statusLabels[o.status] || o.status}</p>
                <p className="text-sm">{o.delivery_address}</p>
                <p className="text-sm text-[var(--admin-text-muted)]">
                  {o.customer_name} · {o.customer_phone}
                </p>
                <div className="flex flex-wrap gap-2 pt-2">
                  {canCourierAccept(o.status) && (
                    <button
                      type="button"
                      disabled={busyId === o.id}
                      onClick={() => accept(o.id)}
                      className="rounded-lg bg-[var(--admin-accent)] px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                    >
                      Принять в доставку
                    </button>
                  )}
                  {o.status === 'in_delivery' && (
                    <div className="flex flex-wrap items-center gap-2 w-full">
                      <input
                        value={codeByOrder[o.id] || ''}
                        onChange={(e) => setCodeByOrder((m) => ({ ...m, [o.id]: e.target.value }))}
                        placeholder="Код от клиента"
                        className="flex-1 min-w-[140px] rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        maxLength={8}
                      />
                      <button
                        type="button"
                        disabled={busyId === o.id}
                        onClick={() => complete(o.id)}
                        className="rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
                      >
                        Доставлен
                      </button>
                    </div>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </CourierLayout>
  );
}
