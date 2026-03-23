import { useCallback, useEffect, useState } from 'react';
import { adminAPI } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

interface StoreRow {
  id: string;
  name: string;
}

interface Summary {
  total_revenue: number;
  orders_count: number;
  average_check: number;
  total_delivery_fees: number;
  date_from: string;
  date_to: string;
}

interface DayRow {
  date: string;
  revenue: number;
  orders: number;
}

export function AdminReports() {
  const { accessToken } = useAuthStore();
  const [stores, setStores] = useState<StoreRow[]>([]);
  const [storeId, setStoreId] = useState('');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [summary, setSummary] = useState<Summary | null>(null);
  const [byDay, setByDay] = useState<DayRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    if (!accessToken) return;
    setLoading(true);
    setError('');
    try {
      const r = await adminAPI.getRevenueReport({
        store_id: storeId || undefined,
        from: from || undefined,
        to: to || undefined,
      });
      const data = (r.data as { data?: { summary?: Summary; by_day?: DayRow[] } })?.data;
      setSummary(data?.summary ?? null);
      setByDay(data?.by_day ?? []);
    } catch {
      setError('Не удалось загрузить отчёт');
    } finally {
      setLoading(false);
    }
  }, [accessToken, storeId, from, to]);

  useEffect(() => {
    adminAPI
      .getStores()
      .then((r) => setStores((r.data as { data?: StoreRow[] })?.data || []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!accessToken) return;
    load();
  }, [accessToken, load]);

  if (!accessToken) return null;

  const maxDayRev = byDay.reduce((m, d) => Math.max(m, d.revenue), 0) || 1;

  return (
    <AdminLayout>
      <div className="space-y-8">
        <div>
          <h1 className="text-2xl font-bold text-white">Отчёт по выручке</h1>
          <p className="text-slate-400 text-sm mt-1">
            Сумма заказов и доставки за период (без отменённых заказов). По умолчанию — последние 30 дней.
          </p>
        </div>

        <div className="flex flex-wrap gap-4 items-end rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-4">
          <label className="block min-w-[200px]">
            <span className="text-xs text-[var(--admin-text-muted)]">Магазин</span>
            <select
              value={storeId}
              onChange={(e) => setStoreId(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            >
              <option value="">Все магазины</option>
              {stores.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="text-xs text-[var(--admin-text-muted)]">С даты</span>
            <input
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className="mt-1 rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            />
          </label>
          <label className="block">
            <span className="text-xs text-[var(--admin-text-muted)]">По дату</span>
            <input
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className="mt-1 rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            />
          </label>
          <button
            type="button"
            onClick={() => load()}
            disabled={loading}
            className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
          >
            Обновить
          </button>
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">{error}</div>
        )}

        {loading && !summary ? (
          <p className="text-[var(--admin-text-muted)]">Загрузка...</p>
        ) : summary ? (
          <>
            <p className="text-sm text-[var(--admin-text-muted)]">
              Период: <span className="text-[var(--admin-text-primary)]">{summary.date_from}</span> —{' '}
              <span className="text-[var(--admin-text-primary)]">{summary.date_to}</span>
            </p>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5">
                <p className="text-xs text-[var(--admin-text-muted)] uppercase tracking-wide">Выручка</p>
                <p className="mt-2 text-2xl font-bold text-white">{summary.total_revenue.toLocaleString('ru-RU')} ₸</p>
              </div>
              <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5">
                <p className="text-xs text-[var(--admin-text-muted)] uppercase tracking-wide">Заказов</p>
                <p className="mt-2 text-2xl font-bold text-white">{summary.orders_count}</p>
              </div>
              <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5">
                <p className="text-xs text-[var(--admin-text-muted)] uppercase tracking-wide">Средний чек</p>
                <p className="mt-2 text-2xl font-bold text-white">{summary.average_check.toLocaleString('ru-RU')} ₸</p>
              </div>
              <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5">
                <p className="text-xs text-[var(--admin-text-muted)] uppercase tracking-wide">Доставка (сумма)</p>
                <p className="mt-2 text-2xl font-bold text-[var(--admin-accent)]">
                  {summary.total_delivery_fees.toLocaleString('ru-RU')} ₸
                </p>
              </div>
            </div>

            <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5">
              <h2 className="font-semibold text-white mb-4">По дням</h2>
              {byDay.length === 0 ? (
                <p className="text-sm text-[var(--admin-text-muted)]">Нет заказов в выбранном периоде</p>
              ) : (
                <div className="space-y-3">
                  {byDay.map((d) => (
                    <div key={d.date} className="flex items-center gap-3">
                      <span className="w-28 shrink-0 text-sm text-[var(--admin-text-muted)]">{d.date}</span>
                      <div className="flex-1 h-8 rounded-lg bg-[var(--admin-bg-elevated)] overflow-hidden">
                        <div
                          className="h-full bg-[var(--admin-accent)]/80 rounded-lg transition-all"
                          style={{ width: `${Math.max(4, (d.revenue / maxDayRev) * 100)}%` }}
                        />
                      </div>
                      <span className="w-32 text-right text-sm font-medium text-white tabular-nums">
                        {d.revenue.toLocaleString('ru-RU')} ₸
                      </span>
                      <span className="w-16 text-right text-xs text-[var(--admin-text-muted)]">{d.orders} зак.</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        ) : null}
      </div>
    </AdminLayout>
  );
}
