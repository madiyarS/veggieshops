import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { adminAPI, type AdminStorePatch } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

interface Store {
  id: string;
  name: string;
  description?: string;
  address: string;
  latitude: number;
  longitude: number;
  phone?: string;
  email?: string;
  delivery_radius_km: number;
  min_order_amount: number;
  max_order_weight_kg?: number | null;
  is_active?: boolean;
}

const DEFAULT_LAT = 51.1694;
const DEFAULT_LON = 71.4491;

export function AdminStores() {
  const navigate = useNavigate();
  const { accessToken } = useAuthStore();
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');
  const [formSuccess, setFormSuccess] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editError, setEditError] = useState('');
  const [editSaving, setEditSaving] = useState(false);
  const [editForm, setEditForm] = useState({
    name: '',
    description: '',
    address: '',
    latitude: '',
    longitude: '',
    phone: '',
    email: '',
    delivery_radius_km: '',
    min_order_amount: '',
    max_order_weight_kg: '',
    is_active: true,
  });
  const [form, setForm] = useState({
    name: '',
    description: '',
    address: '',
    latitude: String(DEFAULT_LAT),
    longitude: String(DEFAULT_LON),
    phone: '',
    email: '',
    delivery_radius_km: '3',
    min_order_amount: '2500',
    max_order_weight_kg: '',
    copy_catalog_from_store_id: '',
  });

  const load = useCallback(
    () =>
      adminAPI
        .getStores()
        .then((r) => setStores((r.data as { data?: Store[] })?.data || []))
        .catch(() => navigate('/login?next=' + encodeURIComponent('/admin/stores'))),
    [navigate]
  );

  useEffect(() => {
    load().finally(() => setLoading(false));
  }, [load]);

  if (!accessToken) return null;

  const submitStore = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    const lat = Number(form.latitude);
    const lon = Number(form.longitude);
    if (!form.name.trim() || !form.address.trim()) {
      setFormError('Укажите название и адрес');
      return;
    }
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
      setFormError('Некорректные координаты');
      return;
    }
    setSaving(true);
    try {
      const res = await adminAPI.createStore({
        name: form.name.trim(),
        description: form.description,
        address: form.address.trim(),
        latitude: lat,
        longitude: lon,
        phone: form.phone || undefined,
        email: form.email || undefined,
        delivery_radius_km: Number(form.delivery_radius_km) || 3,
        min_order_amount: Number(form.min_order_amount) || 2500,
        max_order_weight_kg:
          form.max_order_weight_kg === '' ? undefined : Number(form.max_order_weight_kg) || undefined,
        copy_catalog_from_store_id: form.copy_catalog_from_store_id || undefined,
      });
      const msg = (res.data as { message?: string })?.message?.trim();
      if (msg) setFormSuccess(msg);
      setShowForm(false);
      setForm({
        name: '',
        description: '',
        address: '',
        latitude: String(DEFAULT_LAT),
        longitude: String(DEFAULT_LON),
        phone: '',
        email: '',
        delivery_radius_km: '3',
        min_order_amount: '2500',
        max_order_weight_kg: '',
        copy_catalog_from_store_id: '',
      });
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setFormError(ax.response?.data?.error || 'Не удалось создать магазин');
    } finally {
      setSaving(false);
    }
  };

  const openEdit = (s: Store) => {
    setEditingId(s.id);
    setEditError('');
    setEditForm({
      name: s.name,
      description: s.description || '',
      address: s.address,
      latitude: String(s.latitude ?? ''),
      longitude: String(s.longitude ?? ''),
      phone: s.phone || '',
      email: s.email || '',
      delivery_radius_km: String(s.delivery_radius_km ?? 3),
      min_order_amount: String(s.min_order_amount ?? 2500),
      max_order_weight_kg: s.max_order_weight_kg != null ? String(s.max_order_weight_kg) : '',
      is_active: s.is_active !== false,
    });
  };

  const submitEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingId) return;
    setEditError('');
    const lat = Number(editForm.latitude);
    const lon = Number(editForm.longitude);
    if (!editForm.name.trim() || !editForm.address.trim()) {
      setEditError('Укажите название и адрес');
      return;
    }
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
      setEditError('Некорректные координаты');
      return;
    }
    setEditSaving(true);
    try {
      const patch: AdminStorePatch = {
        name: editForm.name.trim(),
        description: editForm.description,
        address: editForm.address.trim(),
        latitude: lat,
        longitude: lon,
        phone: editForm.phone || undefined,
        email: editForm.email || undefined,
        delivery_radius_km: Number(editForm.delivery_radius_km) || 3,
        min_order_amount: Number(editForm.min_order_amount) || 2500,
        is_active: editForm.is_active,
      };
      if (editForm.max_order_weight_kg.trim() === '') {
        patch.clear_max_weight = true;
      } else {
        const w = Number(editForm.max_order_weight_kg);
        patch.max_order_weight_kg = Number.isFinite(w) ? w : undefined;
      }
      await adminAPI.patchStore(editingId, patch);
      setEditingId(null);
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setEditError(ax.response?.data?.error || 'Не удалось сохранить');
    } finally {
      setEditSaving(false);
    }
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Магазины</h1>
            <p className="text-slate-400 text-sm mt-1">
              Новые активные магазины сразу появляются на главной и в выборе каталога
            </p>
            {formSuccess && (
              <div className="mt-3 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-2 text-sm text-emerald-400 flex flex-wrap items-center justify-between gap-2">
                <span>{formSuccess}</span>
                <button
                  type="button"
                  onClick={() => setFormSuccess('')}
                  className="text-emerald-300 hover:text-white text-xs font-medium"
                >
                  Закрыть
                </button>
              </div>
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              setShowForm((v) => !v);
              setFormError('');
              setFormSuccess('');
            }}
            className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90"
          >
            {showForm ? 'Закрыть форму' : '+ Добавить магазин'}
          </button>
        </div>

        {showForm && (
          <form
            onSubmit={submitStore}
            className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
          >
            <h2 className="font-semibold text-[var(--admin-text-primary)]">Новый магазин</h2>
            {formError && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">{formError}</div>
            )}
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Название *</span>
                <input
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                  required
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Адрес *</span>
                <input
                  value={form.address}
                  onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                  required
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Описание</span>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  rows={2}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Широта *</span>
                <input
                  value={form.latitude}
                  onChange={(e) => setForm((f) => ({ ...f, latitude: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Долгота *</span>
                <input
                  value={form.longitude}
                  onChange={(e) => setForm((f) => ({ ...f, longitude: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <p className="text-xs text-[var(--admin-text-muted)] sm:col-span-2">
                Координаты нужны для проверки доставки (по умолчанию центр Астаны). Уточните под точку на карте.
              </p>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Телефон</span>
                <input
                  value={form.phone}
                  onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Email</span>
                <input
                  type="email"
                  value={form.email}
                  onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Радиус доставки (км)</span>
                <input
                  type="number"
                  min={0.5}
                  step={0.1}
                  value={form.delivery_radius_km}
                  onChange={(e) => setForm((f) => ({ ...f, delivery_radius_km: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Мин. заказ (₸)</span>
                <input
                  type="number"
                  min={0}
                  value={form.min_order_amount}
                  onChange={(e) => setForm((f) => ({ ...f, min_order_amount: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Макс. вес заказа (кг), опционально</span>
                <input
                  type="number"
                  min={0}
                  step={0.1}
                  value={form.max_order_weight_kg}
                  onChange={(e) => setForm((f) => ({ ...f, max_order_weight_kg: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">
                  Скопировать товары и остатки склада из другого магазина (новые карточки и id)
                </span>
                <select
                  value={form.copy_catalog_from_store_id}
                  onChange={(e) => setForm((f) => ({ ...f, copy_catalog_from_store_id: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                >
                  <option value="">— Пустой каталог —</option>
                  {stores.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <button
              type="submit"
              disabled={saving}
              className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white disabled:opacity-50"
            >
              {saving ? 'Создание...' : 'Создать магазин'}
            </button>
          </form>
        )}

        {editingId && (
          <form
            onSubmit={submitEdit}
            className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="font-semibold text-[var(--admin-text-primary)]">Редактирование магазина</h2>
              <button
                type="button"
                onClick={() => setEditingId(null)}
                className="text-sm text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]"
              >
                Закрыть
              </button>
            </div>
            {editError && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">{editError}</div>
            )}
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Название *</span>
                <input
                  value={editForm.name}
                  onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Адрес *</span>
                <input
                  value={editForm.address}
                  onChange={(e) => setEditForm((f) => ({ ...f, address: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Описание</span>
                <textarea
                  value={editForm.description}
                  onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))}
                  rows={2}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Широта *</span>
                <input
                  value={editForm.latitude}
                  onChange={(e) => setEditForm((f) => ({ ...f, latitude: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Долгота *</span>
                <input
                  value={editForm.longitude}
                  onChange={(e) => setEditForm((f) => ({ ...f, longitude: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Телефон</span>
                <input
                  value={editForm.phone}
                  onChange={(e) => setEditForm((f) => ({ ...f, phone: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Email</span>
                <input
                  type="email"
                  value={editForm.email}
                  onChange={(e) => setEditForm((f) => ({ ...f, email: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Радиус доставки (км)</span>
                <input
                  type="number"
                  min={0.5}
                  step={0.1}
                  value={editForm.delivery_radius_km}
                  onChange={(e) => setEditForm((f) => ({ ...f, delivery_radius_km: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Мин. заказ (₸)</span>
                <input
                  type="number"
                  min={0}
                  value={editForm.min_order_amount}
                  onChange={(e) => setEditForm((f) => ({ ...f, min_order_amount: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Макс. вес (кг), пусто = снять лимит</span>
                <input
                  type="number"
                  min={0}
                  step={0.1}
                  value={editForm.max_order_weight_kg}
                  onChange={(e) => setEditForm((f) => ({ ...f, max_order_weight_kg: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="flex items-center gap-2 sm:col-span-2">
                <input
                  type="checkbox"
                  checked={editForm.is_active}
                  onChange={(e) => setEditForm((f) => ({ ...f, is_active: e.target.checked }))}
                />
                <span className="text-sm text-[var(--admin-text-primary)]">Магазин активен (виден на сайте)</span>
              </label>
            </div>
            <button
              type="submit"
              disabled={editSaving}
              className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white disabled:opacity-50"
            >
              {editSaving ? 'Сохранение...' : 'Сохранить изменения'}
            </button>
          </form>
        )}

        {loading && (
          <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-8 text-center text-[var(--admin-text-muted)]">
            Загрузка...
          </div>
        )}

        {!loading && (
          <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[640px]">
                <thead>
                  <tr className="border-b border-[var(--admin-border)]">
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Название</th>
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Адрес</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Радиус (км)</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Мин. заказ (₸)</th>
                    <th className="p-4 text-right text-sm font-medium text-[var(--admin-text-muted)]">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {stores.map((s) => (
                    <tr key={s.id} className="border-b border-[var(--admin-border)]/50 last:border-0 hover:bg-[var(--admin-bg-elevated)]/50 transition-colors">
                      <td className="p-4 font-medium text-[var(--admin-text-primary)]">{s.name}</td>
                      <td className="p-4 text-[var(--admin-text-muted)]">{s.address}</td>
                      <td className="p-4 text-center text-[var(--admin-text-primary)]">{s.delivery_radius_km}</td>
                      <td className="p-4 text-center text-[var(--admin-text-primary)]">{s.min_order_amount}</td>
                      <td className="p-4 text-right">
                        <div className="flex flex-wrap justify-end gap-x-3 gap-y-1">
                          <button
                            type="button"
                            onClick={() => openEdit(s)}
                            className="text-sm font-medium text-[var(--admin-accent)] hover:opacity-90"
                          >
                            Изменить
                          </button>
                          <Link
                            to={`/admin/delivery?store=${s.id}`}
                            className="text-sm font-medium text-slate-300 hover:text-white"
                          >
                            Доставка
                          </Link>
                          <Link
                            to={`/catalog?store=${s.id}`}
                            className="text-sm font-medium text-[var(--admin-accent)] hover:opacity-90"
                          >
                            Каталог →
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {stores.length === 0 && (
              <p className="p-8 text-center text-[var(--admin-text-muted)]">
                Магазинов нет — добавьте первый через форму выше.
              </p>
            )}
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
