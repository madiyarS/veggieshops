import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { RefreshCw, AlertTriangle, ChevronDown, ChevronRight } from 'lucide-react';
import { adminAPI } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

type StockTab = 'balances' | 'moves' | 'count' | 'advanced';

interface StoreRow {
  id: string;
  name: string;
}

interface ZoneRow {
  id: string;
  name: string;
  code: string;
}

interface ProductStock {
  id: string;
  name: string;
  unit: string;
  inventory_unit?: 'piece' | 'weight_gram';
  stock_quantity: number;
  stock_reserved?: number;
  reorder_min_qty?: number;
}

interface MoveRow {
  id: string;
  product_id: string;
  product_name: string;
  delta: number;
  movement_type: string;
  reason?: string;
  created_at: string;
}

type ReceiptLine = {
  key: string;
  productId: string;
  zoneId: string;
  quantity: number;
  expiresAt: string;
};

type AuditLine = {
  key: string;
  productId: string;
  zoneId: string;
  countedQty: number;
};

function newReceiptLine(): ReceiptLine {
  return { key: crypto.randomUUID(), productId: '', zoneId: '', quantity: 1, expiresAt: '' };
}

function newAuditLine(): AuditLine {
  return { key: crypto.randomUUID(), productId: '', zoneId: '', countedQty: 0 };
}

function availableQty(p: ProductStock): number {
  const q = p.stock_quantity ?? 0;
  const r = p.stock_reserved ?? 0;
  return Math.max(0, q - r);
}

function formatStockLine(p: ProductStock): string {
  const a = availableQty(p);
  if (p.inventory_unit === 'weight_gram') {
    const kg = a / 1000;
    return `${kg.toFixed(kg < 10 && kg % 1 !== 0 ? 2 : 1)} кг`;
  }
  return `${a} ${p.unit || 'шт'}`;
}

function formatMinLine(p: ProductStock): string {
  const m = p.reorder_min_qty ?? 0;
  if (m <= 0) return '—';
  if (p.inventory_unit === 'weight_gram') {
    const kg = m / 1000;
    return `${kg.toFixed(kg < 10 && kg % 1 !== 0 ? 2 : 1)} кг`;
  }
  return `${m} ${p.unit || 'шт'}`;
}

function movementTypeLabel(t: string): { label: string; className: string } {
  const map: Record<string, { label: string; className: string }> = {
    receipt: { label: 'Приход', className: 'text-emerald-400' },
    sale: { label: 'Расход', className: 'text-slate-400' },
    write_off_damage: { label: 'Списание (порча)', className: 'text-red-400' },
    write_off_shrink: { label: 'Списание (усушка)', className: 'text-red-400' },
    write_off_resort: { label: 'Списание (пересорт)', className: 'text-red-400' },
    adjustment: { label: 'Корректировка', className: 'text-amber-400' },
    audit_adjustment: { label: 'Инвентаризация', className: 'text-amber-400' },
  };
  return map[t] || { label: t, className: 'text-slate-300' };
}

function ReceiveModal({
  product,
  onClose,
  onDone,
}: {
  product: ProductStock;
  onClose: () => void;
  onDone: (qty: number, note: string) => void;
}) {
  const [qty, setQty] = useState('');
  const [note, setNote] = useState('');
  const unitHint = product.inventory_unit === 'weight_gram' ? 'граммы' : product.unit || 'шт';
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-2xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-6 shadow-xl">
        <h2 className="text-lg font-semibold text-[var(--admin-text-primary)]">Приход — {product.name}</h2>
        <p className="mt-1 text-xs text-[var(--admin-text-muted)]">На склад «Зал» (как в простом режиме zakazik)</p>
        <div className="mt-4 space-y-3">
          <label className="block text-xs text-[var(--admin-text-muted)]">
            Количество ({unitHint})
            <input
              type="number"
              min={1}
              step={product.inventory_unit === 'weight_gram' ? 250 : 1}
              value={qty}
              onChange={(e) => setQty(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
              autoFocus
            />
          </label>
          <label className="block text-xs text-[var(--admin-text-muted)]">
            Примечание (необязательно)
            <input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm text-[var(--admin-text-primary)]"
            />
          </label>
        </div>
        <div className="mt-6 flex gap-2">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 rounded-xl border border-[var(--admin-border)] py-2.5 text-sm text-[var(--admin-text-muted)]"
          >
            Отмена
          </button>
          <button
            type="button"
            onClick={() => {
              const n = parseInt(qty, 10);
              if (!n || n < 1) return;
              onDone(n, note.trim());
            }}
            className="flex-1 rounded-xl bg-[var(--admin-accent)] py-2.5 text-sm font-medium text-white"
          >
            Принять
          </button>
        </div>
      </div>
    </div>
  );
}

function WriteOffModal({
  product,
  onClose,
  onDone,
}: {
  product: ProductStock;
  onClose: () => void;
  onDone: (qty: number, type: 'damage' | 'shrink' | 'resort', reason: string) => void;
}) {
  const [qty, setQty] = useState('');
  const [type, setType] = useState<'damage' | 'shrink' | 'resort'>('damage');
  const [reason, setReason] = useState('');
  const unitHint = product.inventory_unit === 'weight_gram' ? 'граммы' : 'шт';
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-2xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-6 shadow-xl">
        <h2 className="text-lg font-semibold text-[var(--admin-text-primary)]">Списание — {product.name}</h2>
        <p className="mt-1 text-xs text-[var(--admin-text-muted)]">Списание с партий по FEFO</p>
        <div className="mt-4 space-y-3">
          <input
            type="number"
            min={1}
            placeholder={`Кол-во (${unitHint})`}
            value={qty}
            onChange={(e) => setQty(e.target.value)}
            className="w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
          />
          <select
            value={type}
            onChange={(e) => setType(e.target.value as typeof type)}
            className="w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm"
          >
            <option value="damage">Порча</option>
            <option value="shrink">Усушка</option>
            <option value="resort">Пересорт</option>
          </select>
          <input
            placeholder="Причина"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm"
          />
        </div>
        <div className="mt-6 flex gap-2">
          <button type="button" onClick={onClose} className="flex-1 rounded-xl border border-[var(--admin-border)] py-2.5 text-sm">
            Отмена
          </button>
          <button
            type="button"
            onClick={() => {
              const n = parseInt(qty, 10);
              if (!n || n < 1) return;
              onDone(n, type, reason.trim());
            }}
            className="flex-1 rounded-xl bg-red-600 py-2.5 text-sm font-medium text-white"
          >
            Списать
          </button>
        </div>
      </div>
    </div>
  );
}

export function AdminInventory() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { accessToken } = useAuthStore();
  const storeId = searchParams.get('store') || '';

  const [stores, setStores] = useState<StoreRow[]>([]);
  const [zones, setZones] = useState<ZoneRow[]>([]);
  const [products, setProducts] = useState<ProductStock[]>([]);
  const [moves, setMoves] = useState<MoveRow[]>([]);
  const [expiring, setExpiring] = useState<unknown[]>([]);
  const [reorder, setReorder] = useState<unknown[]>([]);
  const [suppliers, setSuppliers] = useState<{ id: string; name: string }[]>([]);
  const [tab, setTab] = useState<StockTab>('balances');
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [loadingMoves, setLoadingMoves] = useState(false);
  const [error, setError] = useState('');
  const [msg, setMsg] = useState('');

  const [receiveFor, setReceiveFor] = useState<ProductStock | null>(null);
  const [writeOffFor, setWriteOffFor] = useState<ProductStock | null>(null);

  const [rcpRows, setRcpRows] = useState<ReceiptLine[]>(() => [newReceiptLine()]);
  const [rcpNote, setRcpNote] = useState('');
  const [rcpSupplier, setRcpSupplier] = useState('');

  const [auditRows, setAuditRows] = useState<AuditLine[]>(() => [newAuditLine()]);
  const [auditNote, setAuditNote] = useState('');

  const [actuals, setActuals] = useState<Record<string, string>>({});
  const [countSaving, setCountSaving] = useState(false);

  const setStoreId = (id: string) => setSearchParams(id ? { store: id } : {}, { replace: true });

  const loadProducts = useCallback(async () => {
    if (!accessToken || !storeId) return;
    setLoadingProducts(true);
    setError('');
    try {
      const r = await adminAPI.listProducts(storeId);
      const list = (r.data as { data?: ProductStock[] })?.data || [];
      setProducts(list);
    } catch {
      setError('Не удалось загрузить товары');
      setProducts([]);
    } finally {
      setLoadingProducts(false);
    }
  }, [accessToken, storeId]);

  const loadMoves = useCallback(async () => {
    if (!accessToken || !storeId) return;
    setLoadingMoves(true);
    try {
      const r = await adminAPI.stockMovesJournal(storeId, 150);
      setMoves((r.data as { data?: MoveRow[] })?.data || []);
    } catch {
      setMoves([]);
    } finally {
      setLoadingMoves(false);
    }
  }, [accessToken, storeId]);

  const loadAdvanced = useCallback(async () => {
    if (!accessToken || !storeId) return;
    try {
      const [z, e, re, s] = await Promise.all([
        adminAPI.stockZones(storeId),
        adminAPI.stockExpiring(storeId, 3),
        adminAPI.stockReorderAlerts(storeId),
        adminAPI.listSuppliers(storeId),
      ]);
      setZones((z.data as { data?: ZoneRow[] })?.data || []);
      setExpiring((e.data as { data?: unknown[] })?.data || []);
      setReorder((re.data as { data?: unknown[] })?.data || []);
      setSuppliers((s.data as { data?: { id: string; name: string }[] })?.data || []);
    } catch {
      /* ignore */
    }
  }, [accessToken, storeId]);

  useEffect(() => {
    if (!accessToken) return;
    adminAPI
      .getStores()
      .then((r) => {
        const st = (r.data as { data?: StoreRow[] })?.data || [];
        setStores(st);
        setSearchParams(
          (prev) => {
            if (st.length === 0) return prev;
            if (prev.get('store')) return prev;
            const n = new URLSearchParams(prev);
            n.set('store', st[0].id);
            return n;
          },
          { replace: true }
        );
      })
      .catch(() => navigate('/login?next=' + encodeURIComponent('/admin/inventory')));
  }, [accessToken, navigate, setSearchParams]);

  useEffect(() => {
    loadProducts();
    loadAdvanced();
  }, [loadProducts, loadAdvanced]);

  useEffect(() => {
    if (tab === 'moves') loadMoves();
  }, [tab, loadMoves]);

  if (!accessToken) return null;

  const lowCount = products.filter((p) => {
    const min = p.reorder_min_qty ?? 0;
    return min > 0 && availableQty(p) < min;
  }).length;

  const handleReceive = async (qty: number, note: string) => {
    if (!storeId || !receiveFor) return;
    setMsg('');
    setError('');
    try {
      await adminAPI.stockReceiveSimple(storeId, { product_id: receiveFor.id, quantity: qty, note: note || undefined });
      setMsg('Приход проведён');
      setReceiveFor(null);
      await loadProducts();
      loadAdvanced();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка прихода');
    }
  };

  const handleWriteOff = async (qty: number, type: 'damage' | 'shrink' | 'resort', reason: string) => {
    if (!storeId || !writeOffFor) return;
    setMsg('');
    setError('');
    try {
      await adminAPI.stockWriteOff(storeId, { product_id: writeOffFor.id, quantity: qty, type, reason: reason || undefined });
      setMsg('Списание проведено');
      setWriteOffFor(null);
      await loadProducts();
      loadAdvanced();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка списания');
    }
  };

  const applyInventory = async () => {
    if (!storeId) return;
    const tasks = products.filter((p) => {
      const raw = actuals[p.id];
      if (raw === undefined || raw === '') return false;
      const n = parseInt(raw, 10);
      if (Number.isNaN(n) || n < 0) return false;
      return n !== availableQty(p);
    });
    if (tasks.length === 0) {
      setError('Введите факт, отличный от учётного, хотя бы по одной позиции');
      return;
    }
    setCountSaving(true);
    setError('');
    setMsg('');
    try {
      await Promise.all(
        tasks.map((p) =>
          adminAPI.stockSetActual(storeId, {
            product_id: p.id,
            actual: parseInt(actuals[p.id], 10),
            note: 'Инвентаризация',
          })
        )
      );
      setActuals({});
      setMsg('Инвентаризация проведена');
      await loadProducts();
      loadAdvanced();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка');
    } finally {
      setCountSaving(false);
    }
  };

  const doReceipt = async () => {
    if (!storeId) return;
    setMsg('');
    setError('');
    const lines = rcpRows
      .filter((r) => r.productId && r.zoneId && r.quantity >= 1)
      .map((r) => ({
        product_id: r.productId,
        zone_id: r.zoneId,
        quantity: r.quantity,
        ...(r.expiresAt.trim() ? { expires_at: r.expiresAt.trim() } : {}),
      }));
    if (lines.length === 0) {
      setError('Добавьте строку: товар, зона, количество');
      return;
    }
    try {
      await adminAPI.stockReceipt(storeId, {
        supplier_id: rcpSupplier || undefined,
        note: rcpNote || undefined,
        lines,
      });
      setMsg('Приход оформлен');
      setRcpRows([newReceiptLine()]);
      await loadProducts();
      loadAdvanced();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка прихода');
    }
  };

  const doAudit = async () => {
    if (!storeId) return;
    setMsg('');
    setError('');
    const lines = auditRows
      .filter((r) => r.productId && r.countedQty >= 0)
      .map((r) => ({
        product_id: r.productId,
        counted_qty: r.countedQty,
        ...(r.zoneId.trim() ? { zone_id: r.zoneId.trim() } : {}),
      }));
    if (lines.length === 0) {
      setError('Добавьте строки пересчёта');
      return;
    }
    try {
      await adminAPI.stockAuditComplete(storeId, { note: auditNote || undefined, lines });
      setMsg('Пересчёт применён');
      setAuditRows([newAuditLine()]);
      await loadProducts();
      loadAdvanced();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка');
    }
  };

  const tabs: { id: StockTab; label: string; hint: string }[] = [
    {
      id: 'balances',
      label: 'Остатки',
      hint: 'Простой приход попадает в зону торгового зала (sales_floor). Списание уменьшает остаток по FEFO.',
    },
    {
      id: 'moves',
      label: 'Движения',
      hint: 'Журнал операций с названием товара; для разборов с поддержкой пришлите время и тип строки.',
    },
    {
      id: 'count',
      label: 'Инвентаризация',
      hint: '«Факт» — сколько вы насчитали на полу/складе в выбранных единицах (шт или г). Учёт подстроится, дельта уйдёт в журнал.',
    },
  ];

  const changedInventoryCount = products.filter((p) => {
    const raw = actuals[p.id];
    if (raw === undefined || raw === '') return false;
    const n = parseInt(raw, 10);
    return !Number.isNaN(n) && n >= 0 && n !== availableQty(p);
  }).length;

  return (
    <AdminLayout>
      <div className="mx-auto max-w-6xl space-y-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Склад</h1>
            <p className="mt-1 max-w-xl text-sm text-[var(--admin-text-muted)]">
              Три вкладки в духе zakazik: остатки, журнал движений и инвентаризация по факту. Партии FEFO и зоны остаются под капотом;
              расширенный приход с зонами — ниже.
            </p>
          </div>
          <label className="block min-w-[220px]">
            <span className="text-xs text-[var(--admin-text-muted)]">Магазин</span>
            <select
              value={storeId}
              onChange={(e) => setStoreId(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            >
              {stores.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">{error}</div>
        )}
        {msg && (
          <div className="rounded-lg border border-green-500/30 bg-green-500/10 px-4 py-2 text-sm text-green-400">{msg}</div>
        )}

        <div className="flex flex-wrap gap-1 border-b border-[var(--admin-border)]">
          {tabs.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => {
                setTab(t.id);
                setError('');
              }}
              className={`border-b-2 px-4 py-3 text-sm font-medium transition ${
                tab === t.id
                  ? 'border-[var(--admin-accent)] text-[var(--admin-accent)]'
                  : 'border-transparent text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <p className="text-xs text-[var(--admin-text-muted)] mt-2 mb-1 max-w-3xl leading-relaxed">
          {tabs.find((x) => x.id === tab)?.hint}
        </p>

        {tab === 'balances' && (
          <div className="space-y-4">
            {lowCount > 0 && (
              <div className="flex items-center gap-2 rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
                <AlertTriangle className="h-4 w-4 shrink-0" />
                <span>
                  {lowCount} поз. ниже минимума — настройте порог в карточке товара (поле «заказать при»)
                </span>
              </div>
            )}
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => void loadProducts()}
                className="rounded-xl border border-[var(--admin-border)] p-2 text-[var(--admin-text-muted)] hover:bg-[var(--admin-bg-elevated)]"
              >
                <RefreshCw className="h-4 w-4" />
              </button>
            </div>
            {loadingProducts ? (
              <p className="text-[var(--admin-text-muted)]">Загрузка…</p>
            ) : (
              <div className="overflow-hidden rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)]">
                <table className="w-full min-w-[720px]">
                  <thead>
                    <tr className="border-b border-[var(--admin-border)] bg-[var(--admin-bg-elevated)]/50">
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Товар</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Доступно</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Мин.</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Статус</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Действия</th>
                    </tr>
                  </thead>
                  <tbody>
                    {products.map((p) => {
                      const a = availableQty(p);
                      const min = p.reorder_min_qty ?? 0;
                      const empty = a <= 0;
                      const low = !empty && min > 0 && a < min;
                      return (
                        <tr
                          key={p.id}
                          className={`border-b border-[var(--admin-border)]/40 ${empty ? 'border-l-4 border-l-red-500' : low ? 'border-l-4 border-l-amber-500' : ''}`}
                        >
                          <td className="px-4 py-3 font-medium text-[var(--admin-text-primary)]">{p.name}</td>
                          <td className="px-4 py-3 font-mono text-sm text-[var(--admin-text-primary)]">{formatStockLine(p)}</td>
                          <td className="px-4 py-3 text-sm text-[var(--admin-text-muted)]">{formatMinLine(p)}</td>
                          <td className="px-4 py-3 text-sm">
                            {empty ? (
                              <span className="rounded-full bg-red-500/20 px-2 py-0.5 text-xs text-red-300">Нет</span>
                            ) : low ? (
                              <span className="rounded-full bg-amber-500/20 px-2 py-0.5 text-xs text-amber-200">Мало</span>
                            ) : (
                              <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-xs text-emerald-300">Ок</span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-right">
                            <button
                              type="button"
                              onClick={() => setReceiveFor(p)}
                              className="mr-2 rounded-lg bg-emerald-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-emerald-500"
                            >
                              Приход
                            </button>
                            <button
                              type="button"
                              onClick={() => setWriteOffFor(p)}
                              className="rounded-lg bg-red-700/80 px-2.5 py-1 text-xs font-medium text-white hover:bg-red-600"
                            >
                              Списать
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                    {products.length === 0 && (
                      <tr>
                        <td colSpan={5} className="px-4 py-10 text-center text-[var(--admin-text-muted)]">
                          Нет товаров в этом магазине
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {tab === 'moves' && (
          <div>
            {loadingMoves ? (
              <p className="text-[var(--admin-text-muted)]">Загрузка…</p>
            ) : (
              <div className="max-h-[65vh] overflow-auto rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)]">
                <table className="w-full">
                  <thead className="sticky top-0 z-10 bg-[var(--admin-bg-elevated)]">
                    <tr className="border-b border-[var(--admin-border)]">
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Когда</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Товар</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Тип</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Δ</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-[var(--admin-text-muted)]">Примечание</th>
                    </tr>
                  </thead>
                  <tbody>
                    {moves.map((m) => {
                      const tl = movementTypeLabel(m.movement_type);
                      return (
                        <tr key={m.id} className="border-b border-[var(--admin-border)]/40">
                          <td className="px-4 py-2 text-sm text-[var(--admin-text-muted)]">
                            {new Date(m.created_at).toLocaleString('ru')}
                          </td>
                          <td className="px-4 py-2 text-sm text-[var(--admin-text-primary)]">{m.product_name}</td>
                          <td className={`px-4 py-2 text-sm font-medium ${tl.className}`}>{tl.label}</td>
                          <td className="px-4 py-2 font-mono text-sm text-[var(--admin-text-primary)]">
                            {m.delta > 0 ? '+' : ''}
                            {m.delta}
                          </td>
                          <td className="px-4 py-2 text-sm text-[var(--admin-text-muted)]">{m.reason || '—'}</td>
                        </tr>
                      );
                    })}
                    {moves.length === 0 && (
                      <tr>
                        <td colSpan={5} className="px-4 py-10 text-center text-[var(--admin-text-muted)]">
                          Нет движений
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {tab === 'count' && (
          <div className="space-y-4">
            <p className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-4 py-3 text-sm text-[var(--admin-text-muted)]">
              Введите <strong className="text-[var(--admin-text-primary)]">фактический остаток</strong> в тех же единицах, что и склад: для весовых товаров —{' '}
              <strong>граммы</strong>, для штучных — штуки. Расхождения оформляются одной кнопкой.
            </p>
            <div className="overflow-hidden rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)]">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-[var(--admin-border)] bg-[var(--admin-bg-elevated)]/50">
                    <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--admin-text-muted)]">Товар</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--admin-text-muted)]">Учёт (доступно)</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--admin-text-muted)]">Факт</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--admin-text-muted)]">Разница</th>
                  </tr>
                </thead>
                <tbody>
                  {products.map((p) => {
                    const sys = availableQty(p);
                    const raw = actuals[p.id];
                    let diff: number | null = null;
                    if (raw !== undefined && raw !== '') {
                      const n = parseInt(raw, 10);
                      if (!Number.isNaN(n)) diff = n - sys;
                    }
                    return (
                      <tr key={p.id} className="border-b border-[var(--admin-border)]/40">
                        <td className="px-4 py-2 font-medium text-[var(--admin-text-primary)]">{p.name}</td>
                        <td className="px-4 py-2 font-mono text-sm text-[var(--admin-text-muted)]">
                          {sys}
                          {p.inventory_unit === 'weight_gram' ? ' г' : ` ${p.unit || 'шт'}`}
                        </td>
                        <td className="px-4 py-2">
                          <input
                            type="number"
                            min={0}
                            value={actuals[p.id] ?? ''}
                            onChange={(e) => setActuals((prev) => ({ ...prev, [p.id]: e.target.value }))}
                            placeholder={String(sys)}
                            className="w-28 rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-2 py-1.5 font-mono text-sm text-[var(--admin-text-primary)]"
                          />
                        </td>
                        <td className="px-4 py-2 font-mono text-sm">
                          {diff === null ? (
                            <span className="text-slate-600">—</span>
                          ) : diff === 0 ? (
                            <span className="text-slate-500">0</span>
                          ) : diff > 0 ? (
                            <span className="text-emerald-400">+{diff}</span>
                          ) : (
                            <span className="text-red-400">{diff}</span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <span className="text-sm text-[var(--admin-text-muted)]">
                {changedInventoryCount > 0 ? `Изменено позиций: ${changedInventoryCount}` : 'Введите факт там, где пересчитали'}
              </span>
              <button
                type="button"
                disabled={countSaving || changedInventoryCount === 0}
                onClick={() => void applyInventory()}
                className="rounded-xl bg-[var(--admin-accent)] px-6 py-2.5 text-sm font-semibold text-white disabled:opacity-40"
              >
                {countSaving ? '…' : 'Провести инвентаризацию'}
              </button>
            </div>
          </div>
        )}

        <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)]">
          <button
            type="button"
            onClick={() => setAdvancedOpen((o) => !o)}
            className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium text-[var(--admin-text-primary)]"
          >
            <span className="flex items-center gap-2">
              {advancedOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              Расширенно: партии, зоны, многострочный приход
            </span>
            <span className="text-xs font-normal text-[var(--admin-text-muted)]">для опытных пользователей</span>
          </button>
          {advancedOpen && (
            <div className="space-y-8 border-t border-[var(--admin-border)] p-4">
              <section>
                <h2 className="font-semibold text-[var(--admin-text-primary)]">Скоро срок (3 дня)</h2>
                <ul className="mt-2 space-y-1 text-sm text-[var(--admin-text-muted)]">
                  {(expiring as { product_name?: string; quantity?: number; expires_at?: string }[]).map((row, i) => (
                    <li key={i}>
                      {row.product_name} — {row.quantity}, до {row.expires_at?.slice(0, 10) || '—'}
                    </li>
                  ))}
                  {expiring.length === 0 && <li>Нет партий с истекающим сроком в окне</li>}
                </ul>
              </section>
              <section>
                <h2 className="font-semibold text-[var(--admin-text-primary)]">Ниже минимума</h2>
                <ul className="mt-2 space-y-1 text-sm text-[var(--admin-text-muted)]">
                  {(reorder as { name?: string; available?: number; reorder_min_qty?: number }[]).map((row, i) => (
                    <li key={i}>
                      {row.name}: {row.available} / мин {row.reorder_min_qty}
                    </li>
                  ))}
                  {reorder.length === 0 && <li>В норме</li>}
                </ul>
              </section>
              <section className="space-y-3">
                <h2 className="font-semibold text-[var(--admin-text-primary)]">Приход с выбором зоны и срока</h2>
                <div className="space-y-3">
                  {rcpRows.map((row, idx) => (
                    <div
                      key={row.key}
                      className="grid gap-2 rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] p-3 sm:grid-cols-2 lg:grid-cols-6"
                    >
                      <select
                        value={row.productId}
                        onChange={(e) =>
                          setRcpRows((rows) => rows.map((r) => (r.key === row.key ? { ...r, productId: e.target.value } : r)))
                        }
                        className="rounded border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-2 py-1.5 text-sm lg:col-span-2"
                      >
                        <option value="">Товар</option>
                        {products.map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.name}
                          </option>
                        ))}
                      </select>
                      <select
                        value={row.zoneId}
                        onChange={(e) =>
                          setRcpRows((rows) => rows.map((r) => (r.key === row.key ? { ...r, zoneId: e.target.value } : r)))
                        }
                        className="rounded border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-2 py-1.5 text-sm lg:col-span-2"
                      >
                        <option value="">Зона</option>
                        {zones.map((z) => (
                          <option key={z.id} value={z.id}>
                            {z.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min={1}
                        value={row.quantity}
                        onChange={(e) =>
                          setRcpRows((rows) =>
                            rows.map((r) => (r.key === row.key ? { ...r, quantity: Number(e.target.value) } : r))
                          )
                        }
                        className="rounded border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-2 py-1.5 text-sm"
                      />
                      <input
                        type="date"
                        value={row.expiresAt}
                        onChange={(e) =>
                          setRcpRows((rows) => rows.map((r) => (r.key === row.key ? { ...r, expiresAt: e.target.value } : r)))
                        }
                        className="rounded border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-2 py-1.5 text-sm"
                      />
                      <div className="flex gap-2 lg:col-span-6">
                        <button
                          type="button"
                          onClick={() => setRcpRows((rows) => rows.filter((r) => r.key !== row.key))}
                          disabled={rcpRows.length <= 1}
                          className="text-xs text-red-400 disabled:opacity-40"
                        >
                          Удалить строку
                        </button>
                        {idx === rcpRows.length - 1 && (
                          <button
                            type="button"
                            onClick={() => setRcpRows((rows) => [...rows, newReceiptLine()])}
                            className="text-xs text-[var(--admin-accent)]"
                          >
                            + строка
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
                <input
                  placeholder="ID поставщика (опц.)"
                  value={rcpSupplier}
                  onChange={(e) => setRcpSupplier(e.target.value)}
                  className="mt-2 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm"
                />
                <input
                  placeholder="Примечание"
                  value={rcpNote}
                  onChange={(e) => setRcpNote(e.target.value)}
                  className="mt-2 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm"
                />
                <button type="button" onClick={() => void doReceipt()} className="mt-2 rounded-lg bg-[var(--admin-accent)] px-4 py-2 text-sm text-white">
                  Провести приход
                </button>
              </section>
              <section className="space-y-3">
                <h2 className="font-semibold text-[var(--admin-text-primary)]">Инвентаризация по зонам (полная)</h2>
                <p className="text-xs text-[var(--admin-text-muted)]">
                  Альтернатива вкладке «Инвентаризация» сверху: расхождения через партии и FEFO.
                </p>
                {auditRows.map((row, idx) => (
                  <div key={row.key} className="grid gap-2 rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] p-3 sm:grid-cols-2 lg:grid-cols-4">
                    <select
                      value={row.productId}
                      onChange={(e) =>
                        setAuditRows((rows) => rows.map((r) => (r.key === row.key ? { ...r, productId: e.target.value } : r)))
                      }
                      className="rounded border px-2 py-1.5 text-sm lg:col-span-2"
                    >
                      <option value="">Товар</option>
                      {products.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.name}
                        </option>
                      ))}
                    </select>
                    <select
                      value={row.zoneId}
                      onChange={(e) =>
                        setAuditRows((rows) => rows.map((r) => (r.key === row.key ? { ...r, zoneId: e.target.value } : r)))
                      }
                      className="rounded border px-2 py-1.5 text-sm"
                    >
                      <option value="">Зона (опц.)</option>
                      {zones.map((z) => (
                        <option key={z.id} value={z.id}>
                          {z.name}
                        </option>
                      ))}
                    </select>
                    <input
                      type="number"
                      min={0}
                      value={row.countedQty}
                      onChange={(e) =>
                        setAuditRows((rows) =>
                          rows.map((r) => (r.key === row.key ? { ...r, countedQty: Number(e.target.value) } : r))
                        )
                      }
                      className="rounded border px-2 py-1.5 text-sm"
                    />
                    <div className="flex gap-2 lg:col-span-4">
                      <button
                        type="button"
                        onClick={() => setAuditRows((rows) => rows.filter((r) => r.key !== row.key))}
                        disabled={auditRows.length <= 1}
                        className="text-xs text-red-400"
                      >
                        Удалить
                      </button>
                      {idx === auditRows.length - 1 && (
                        <button type="button" onClick={() => setAuditRows((rows) => [...rows, newAuditLine()])} className="text-xs text-[var(--admin-accent)]">
                          + строка
                        </button>
                      )}
                    </div>
                  </div>
                ))}
                <input
                  placeholder="Примечание"
                  value={auditNote}
                  onChange={(e) => setAuditNote(e.target.value)}
                  className="w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm"
                />
                <button type="button" onClick={() => void doAudit()} className="rounded-lg bg-[var(--admin-accent)] px-4 py-2 text-sm text-white">
                  Закрыть пересчёт
                </button>
              </section>
              <section>
                <h2 className="font-semibold text-[var(--admin-text-primary)]">Зоны и поставщики</h2>
                <ul className="mt-2 text-sm text-[var(--admin-text-muted)]">
                  {zones.map((z) => (
                    <li key={z.id}>
                      {z.name} ({z.code})
                    </li>
                  ))}
                </ul>
                <ul className="mt-2 text-sm text-[var(--admin-text-muted)]">
                  {suppliers.map((s) => (
                    <li key={s.id}>{s.name}</li>
                  ))}
                </ul>
              </section>
            </div>
          )}
        </div>
      </div>

      {receiveFor && (
        <ReceiveModal product={receiveFor} onClose={() => setReceiveFor(null)} onDone={(q, n) => void handleReceive(q, n)} />
      )}
      {writeOffFor && (
        <WriteOffModal
          product={writeOffFor}
          onClose={() => setWriteOffFor(null)}
          onDone={(q, t, r) => void handleWriteOff(q, t, r)}
        />
      )}
    </AdminLayout>
  );
}
