import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  adminAPI,
  type AdminDistrictPatch,
  type AdminDistrictPayload,
  type AdminTimeSlotPatch,
  type AdminTimeSlotPayload,
} from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

interface StoreRow {
  id: string;
  name: string;
}

interface DistrictRow {
  id: string;
  store_id: string;
  name: string;
  distance_km: number;
  delivery_fee_regular: number;
  delivery_fee_express: number;
  is_active: boolean;
  streets?: { street_name: string }[];
}

interface SlotRow {
  id: string;
  store_id: string;
  day_of_week: number;
  start_time: string;
  end_time: string;
  max_orders: number;
  is_active: boolean;
}

const DAYS = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

function timeShort(t: string) {
  if (!t) return '';
  return t.length >= 5 ? t.slice(0, 5) : t;
}

const emptyDistrict = {
  name: '',
  distance_km: '1',
  delivery_fee_regular: '500',
  delivery_fee_express: '800',
  is_active: true,
  streetsText: '',
};

const emptySlot = {
  day_of_week: '1',
  start_time: '09:00',
  end_time: '11:00',
  max_orders: '10',
  is_active: true,
};

export function AdminDelivery() {
  const navigate = useNavigate();
  const { accessToken } = useAuthStore();
  const [searchParams, setSearchParams] = useSearchParams();
  const storeIdFromUrl = searchParams.get('store') || '';

  const [stores, setStores] = useState<StoreRow[]>([]);
  const [storeId, setStoreId] = useState(storeIdFromUrl);
  const [tab, setTab] = useState<'districts' | 'slots'>('districts');
  const [districts, setDistricts] = useState<DistrictRow[]>([]);
  const [slots, setSlots] = useState<SlotRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const [saving, setSaving] = useState(false);

  const [dEditing, setDEditing] = useState<string | null>(null);
  const [dForm, setDForm] = useState(emptyDistrict);

  const [sEditing, setSEditing] = useState<string | null>(null);
  const [sForm, setSForm] = useState(emptySlot);

  const loadStores = useCallback((): Promise<void> => {
    if (!accessToken) return Promise.resolve();
    return adminAPI
      .getStores()
      .then((r) => setStores((r.data as { data?: StoreRow[] })?.data || []))
      .catch(() => navigate('/login?next=' + encodeURIComponent('/admin/delivery')));
  }, [accessToken, navigate]);

  const loadData = useCallback(() => {
    if (!accessToken || !storeId) {
      setDistricts([]);
      setSlots([]);
      return;
    }
    setError('');
    setInfo('');
    const d = adminAPI.listDistricts(storeId).then((r) => setDistricts((r.data as { data?: DistrictRow[] })?.data || []));
    const s = adminAPI.listTimeSlotsAdmin(storeId).then((r) => setSlots((r.data as { data?: SlotRow[] })?.data || []));
    return Promise.all([d, s]).catch(() => {
      setError('Не удалось загрузить данные доставки');
      navigate('/login?next=' + encodeURIComponent('/admin/delivery'));
    });
  }, [accessToken, storeId, navigate]);

  useEffect(() => {
    loadStores().finally(() => setLoading(false));
  }, [loadStores]);

  useEffect(() => {
    if (!storeIdFromUrl && stores.length === 1) {
      const only = stores[0].id;
      setStoreId(only);
      setSearchParams({ store: only }, { replace: true });
    } else if (storeIdFromUrl && storeIdFromUrl !== storeId) {
      setStoreId(storeIdFromUrl);
    }
  }, [storeIdFromUrl, stores, setSearchParams, storeId]);

  useEffect(() => {
    if (!storeId) return;
    setLoading(true);
    loadData()?.finally(() => setLoading(false));
  }, [storeId, loadData]);

  const selectedStore = useMemo(() => stores.find((s) => s.id === storeId), [stores, storeId]);

  const onPickStore = (id: string) => {
    setStoreId(id);
    if (id) setSearchParams({ store: id });
    else setSearchParams({});
    setDEditing(null);
    setSEditing(null);
  };

  if (!accessToken) return null;

  const startNewDistrict = () => {
    setDEditing('new');
    setDForm(emptyDistrict);
  };

  const startEditDistrict = (row: DistrictRow) => {
    setDEditing(row.id);
    const lines = (row.streets || []).map((x) => x.street_name).join('\n');
    setDForm({
      name: row.name,
      distance_km: String(row.distance_km),
      delivery_fee_regular: String(row.delivery_fee_regular),
      delivery_fee_express: String(row.delivery_fee_express),
      is_active: row.is_active,
      streetsText: lines,
    });
  };

  const submitDistrict = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId) return;
    if (!dForm.name.trim()) {
      setError('Укажите название района');
      return;
    }
    setSaving(true);
    setError('');
    setInfo('');
    const streets = dForm.streetsText
      .split('\n')
      .map((x) => x.trim())
      .filter(Boolean);
    try {
      if (dEditing === 'new') {
        const body: AdminDistrictPayload = {
          name: dForm.name.trim(),
          distance_km: Number(dForm.distance_km) || 0,
          delivery_fee_regular: Number(dForm.delivery_fee_regular) || 0,
          delivery_fee_express: Number(dForm.delivery_fee_express) || 0,
          is_active: dForm.is_active,
          streets,
        };
        await adminAPI.createDistrict(storeId, body);
      } else if (dEditing) {
        const patch: AdminDistrictPatch = {
          name: dForm.name.trim(),
          distance_km: Number(dForm.distance_km) || 0,
          delivery_fee_regular: Number(dForm.delivery_fee_regular) || 0,
          delivery_fee_express: Number(dForm.delivery_fee_express) || 0,
          is_active: dForm.is_active,
          streets,
        };
        await adminAPI.patchDistrict(dEditing, patch);
      }
      setDEditing(null);
      await loadData();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка сохранения района');
    } finally {
      setSaving(false);
    }
  };

  const removeDistrict = async (row: DistrictRow) => {
    if (!window.confirm(`Удалить или отключить район «${row.name}»?`)) return;
    setError('');
    setInfo('');
    try {
      const r = await adminAPI.deleteDistrict(row.id);
      const msg = (r.data as { message?: string })?.message;
      if (msg) setInfo(msg);
      await loadData();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Не удалось удалить');
    }
  };

  const startNewSlot = () => {
    setSEditing('new');
    setSForm(emptySlot);
  };

  const startEditSlot = (row: SlotRow) => {
    setSEditing(row.id);
    setSForm({
      day_of_week: String(row.day_of_week),
      start_time: timeShort(row.start_time),
      end_time: timeShort(row.end_time),
      max_orders: String(row.max_orders || 10),
      is_active: row.is_active,
    });
  };

  const submitSlot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId) return;
    const dow = Number(sForm.day_of_week);
    if (!Number.isFinite(dow) || dow < 0 || dow > 6) {
      setError('День недели: 0–6');
      return;
    }
    setSaving(true);
    setError('');
    setInfo('');
    try {
      if (sEditing === 'new') {
        const body: AdminTimeSlotPayload = {
          day_of_week: dow,
          start_time: sForm.start_time.trim(),
          end_time: sForm.end_time.trim(),
          max_orders: Number(sForm.max_orders) || 10,
          is_active: sForm.is_active,
        };
        await adminAPI.createTimeSlot(storeId, body);
      } else if (sEditing) {
        const patch: AdminTimeSlotPatch = {
          day_of_week: dow,
          start_time: sForm.start_time.trim(),
          end_time: sForm.end_time.trim(),
          max_orders: Number(sForm.max_orders) || 10,
          is_active: sForm.is_active,
        };
        await adminAPI.patchTimeSlot(sEditing, patch);
      }
      setSEditing(null);
      await loadData();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка сохранения слота');
    } finally {
      setSaving(false);
    }
  };

  const removeSlot = async (row: SlotRow) => {
    if (!window.confirm(`Удалить или отключить слот ${DAYS[row.day_of_week]} ${timeShort(row.start_time)}–${timeShort(row.end_time)}?`))
      return;
    setError('');
    setInfo('');
    try {
      const r = await adminAPI.deleteTimeSlot(row.id);
      const msg = (r.data as { message?: string })?.message;
      if (msg) setInfo(msg);
      await loadData();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Не удалось удалить');
    }
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Доставка</h1>
          <p className="text-slate-400 text-sm mt-1">
            Районы (цены доставки, улицы) и окна времени для выбранного магазина. Цены товаров — в разделе «Товары».
          </p>
        </div>

        <div className="flex flex-wrap items-end gap-4">
          <label className="block min-w-[200px]">
            <span className="text-xs text-[var(--admin-text-muted)]">Магазин</span>
            <select
              value={storeId}
              onChange={(e) => onPickStore(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            >
              <option value="">— выберите —</option>
              {stores.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          {selectedStore && (
            <span className="text-sm text-[var(--admin-text-muted)] pb-2">Выбрано: {selectedStore.name}</span>
          )}
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">{error}</div>
        )}
        {info && (
          <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-2 text-sm text-emerald-400">
            {info}
          </div>
        )}

        {!storeId && !loading && (
          <p className="text-[var(--admin-text-muted)]">Выберите магазин, чтобы редактировать доставку.</p>
        )}

        {storeId && (
          <>
            <div className="flex gap-2 border-b border-[var(--admin-border)]">
              <button
                type="button"
                onClick={() => setTab('districts')}
                className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  tab === 'districts'
                    ? 'border-[var(--admin-accent)] text-[var(--admin-text-primary)]'
                    : 'border-transparent text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]'
                }`}
              >
                Районы и тарифы
              </button>
              <button
                type="button"
                onClick={() => setTab('slots')}
                className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  tab === 'slots'
                    ? 'border-[var(--admin-accent)] text-[var(--admin-text-primary)]'
                    : 'border-transparent text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]'
                }`}
              >
                Окна доставки
              </button>
            </div>

            {loading && (
              <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-8 text-center text-[var(--admin-text-muted)]">
                Загрузка...
              </div>
            )}

            {!loading && tab === 'districts' && (
              <div className="space-y-4">
                <div className="flex justify-end">
                  <button
                    type="button"
                    onClick={() => (dEditing ? setDEditing(null) : startNewDistrict())}
                    className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90"
                  >
                    {dEditing ? 'Отмена' : '+ Район'}
                  </button>
                </div>

                {dEditing && (
                  <form
                    onSubmit={submitDistrict}
                    className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
                  >
                    <h2 className="font-semibold text-[var(--admin-text-primary)]">
                      {dEditing === 'new' ? 'Новый район' : 'Редактирование района'}
                    </h2>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <label className="block sm:col-span-2">
                        <span className="text-xs text-[var(--admin-text-muted)]">Название *</span>
                        <input
                          value={dForm.name}
                          onChange={(e) => setDForm((f) => ({ ...f, name: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">Расстояние (км)</span>
                        <input
                          type="number"
                          step="0.1"
                          min={0}
                          value={dForm.distance_km}
                          onChange={(e) => setDForm((f) => ({ ...f, distance_km: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="flex items-center gap-2 pt-6">
                        <input
                          type="checkbox"
                          checked={dForm.is_active}
                          onChange={(e) => setDForm((f) => ({ ...f, is_active: e.target.checked }))}
                        />
                        <span className="text-sm text-[var(--admin-text-primary)]">Активен</span>
                      </label>
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">Доставка обычная (₸)</span>
                        <input
                          type="number"
                          min={0}
                          value={dForm.delivery_fee_regular}
                          onChange={(e) => setDForm((f) => ({ ...f, delivery_fee_regular: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">Доставка экспресс (₸)</span>
                        <input
                          type="number"
                          min={0}
                          value={dForm.delivery_fee_express}
                          onChange={(e) => setDForm((f) => ({ ...f, delivery_fee_express: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="block sm:col-span-2">
                        <span className="text-xs text-[var(--admin-text-muted)]">Улицы (по одной на строку)</span>
                        <textarea
                          rows={4}
                          value={dForm.streetsText}
                          onChange={(e) => setDForm((f) => ({ ...f, streetsText: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)] font-mono text-sm"
                        />
                      </label>
                    </div>
                    <button
                      type="submit"
                      disabled={saving}
                      className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white disabled:opacity-50"
                    >
                      {saving ? 'Сохранение...' : 'Сохранить'}
                    </button>
                  </form>
                )}

                <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] overflow-hidden">
                  <div className="overflow-x-auto">
                    <table className="w-full min-w-[720px]">
                      <thead>
                        <tr className="border-b border-[var(--admin-border)]">
                          <th className="p-3 text-left text-xs font-medium text-[var(--admin-text-muted)]">Название</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">км</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">Обычн.</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">Экспр.</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">Акт.</th>
                          <th className="p-3 text-right text-xs font-medium text-[var(--admin-text-muted)]"></th>
                        </tr>
                      </thead>
                      <tbody>
                        {districts.map((d) => (
                          <tr
                            key={d.id}
                            className="border-b border-[var(--admin-border)]/50 last:border-0 hover:bg-[var(--admin-bg-elevated)]/50"
                          >
                            <td className="p-3 text-sm text-[var(--admin-text-primary)]">{d.name}</td>
                            <td className="p-3 text-center text-sm text-[var(--admin-text-muted)]">{d.distance_km}</td>
                            <td className="p-3 text-center text-sm">{d.delivery_fee_regular}</td>
                            <td className="p-3 text-center text-sm">{d.delivery_fee_express}</td>
                            <td className="p-3 text-center text-sm">{d.is_active ? 'да' : 'нет'}</td>
                            <td className="p-3 text-right space-x-2 whitespace-nowrap">
                              <button
                                type="button"
                                onClick={() => startEditDistrict(d)}
                                className="text-sm font-medium text-[var(--admin-accent)] hover:opacity-90"
                              >
                                Изменить
                              </button>
                              <button
                                type="button"
                                onClick={() => removeDistrict(d)}
                                className="text-sm font-medium text-red-400 hover:text-red-300"
                              >
                                Удалить
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {districts.length === 0 && !dEditing && (
                    <p className="p-6 text-center text-sm text-[var(--admin-text-muted)]">Районов пока нет</p>
                  )}
                </div>
              </div>
            )}

            {!loading && tab === 'slots' && (
              <div className="space-y-4">
                <div className="flex justify-end">
                  <button
                    type="button"
                    onClick={() => (sEditing ? setSEditing(null) : startNewSlot())}
                    className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90"
                  >
                    {sEditing ? 'Отмена' : '+ Окно'}
                  </button>
                </div>

                {sEditing && (
                  <form
                    onSubmit={submitSlot}
                    className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
                  >
                    <h2 className="font-semibold text-[var(--admin-text-primary)]">
                      {sEditing === 'new' ? 'Новое окно' : 'Редактирование окна'}
                    </h2>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">День (0=Вс … 6=Сб)</span>
                        <input
                          type="number"
                          min={0}
                          max={6}
                          value={sForm.day_of_week}
                          onChange={(e) => setSForm((f) => ({ ...f, day_of_week: e.target.value }))}
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="flex items-center gap-2 pt-6">
                        <input
                          type="checkbox"
                          checked={sForm.is_active}
                          onChange={(e) => setSForm((f) => ({ ...f, is_active: e.target.checked }))}
                        />
                        <span className="text-sm text-[var(--admin-text-primary)]">Активно</span>
                      </label>
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">Начало (ЧЧ:ММ)</span>
                        <input
                          value={sForm.start_time}
                          onChange={(e) => setSForm((f) => ({ ...f, start_time: e.target.value }))}
                          placeholder="09:00"
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="block">
                        <span className="text-xs text-[var(--admin-text-muted)]">Конец (ЧЧ:ММ)</span>
                        <input
                          value={sForm.end_time}
                          onChange={(e) => setSForm((f) => ({ ...f, end_time: e.target.value }))}
                          placeholder="11:00"
                          className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                      <label className="block sm:col-span-2">
                        <span className="text-xs text-[var(--admin-text-muted)]">Макс. заказов в окне</span>
                        <input
                          type="number"
                          min={1}
                          value={sForm.max_orders}
                          onChange={(e) => setSForm((f) => ({ ...f, max_orders: e.target.value }))}
                          className="mt-1 w-full max-w-xs rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                        />
                      </label>
                    </div>
                    <button
                      type="submit"
                      disabled={saving}
                      className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white disabled:opacity-50"
                    >
                      {saving ? 'Сохранение...' : 'Сохранить'}
                    </button>
                  </form>
                )}

                <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] overflow-hidden">
                  <div className="overflow-x-auto">
                    <table className="w-full min-w-[560px]">
                      <thead>
                        <tr className="border-b border-[var(--admin-border)]">
                          <th className="p-3 text-left text-xs font-medium text-[var(--admin-text-muted)]">День</th>
                          <th className="p-3 text-left text-xs font-medium text-[var(--admin-text-muted)]">Время</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">Лимит</th>
                          <th className="p-3 text-center text-xs font-medium text-[var(--admin-text-muted)]">Акт.</th>
                          <th className="p-3 text-right text-xs font-medium text-[var(--admin-text-muted)]"></th>
                        </tr>
                      </thead>
                      <tbody>
                        {slots.map((sl) => (
                          <tr
                            key={sl.id}
                            className="border-b border-[var(--admin-border)]/50 last:border-0 hover:bg-[var(--admin-bg-elevated)]/50"
                          >
                            <td className="p-3 text-sm text-[var(--admin-text-primary)]">
                              {DAYS[sl.day_of_week] ?? sl.day_of_week}
                            </td>
                            <td className="p-3 text-sm text-[var(--admin-text-muted)]">
                              {timeShort(sl.start_time)} — {timeShort(sl.end_time)}
                            </td>
                            <td className="p-3 text-center text-sm">{sl.max_orders}</td>
                            <td className="p-3 text-center text-sm">{sl.is_active ? 'да' : 'нет'}</td>
                            <td className="p-3 text-right space-x-2 whitespace-nowrap">
                              <button
                                type="button"
                                onClick={() => startEditSlot(sl)}
                                className="text-sm font-medium text-[var(--admin-accent)] hover:opacity-90"
                              >
                                Изменить
                              </button>
                              <button
                                type="button"
                                onClick={() => removeSlot(sl)}
                                className="text-sm font-medium text-red-400 hover:text-red-300"
                              >
                                Удалить
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {slots.length === 0 && !sEditing && (
                    <p className="p-6 text-center text-sm text-[var(--admin-text-muted)]">Окон пока нет</p>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </AdminLayout>
  );
}
