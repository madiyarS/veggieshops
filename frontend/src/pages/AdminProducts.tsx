import { useCallback, useEffect, useState, type ChangeEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { adminAPI, type AdminProductPatch, type AdminProductPayload } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

interface StoreRow {
  id: string;
  name: string;
}

interface CategoryRow {
  id: string;
  name: string;
}

interface ProductRow {
  id: string;
  store_id: string;
  category_id: string;
  name: string;
  description: string;
  price: number;
  weight_gram: number;
  unit: string;
  stock_quantity: number;
  stock_reserved?: number;
  inventory_unit?: string;
  cart_step_grams?: number;
  reorder_min_qty?: number;
  image_url: string;
  origin: string;
  shelf_life_days: number | null;
  is_available: boolean;
  is_active: boolean;
  category?: { id: string; name: string };
}

type InventoryMode = 'piece' | 'weight_gram';

function productToForm(p: ProductRow | null) {
  if (!p) {
    return {
      category_id: '',
      name: '',
      description: '',
      price: '' as string | number,
      weight_gram: '' as string | number,
      unit: 'шт',
      stock_quantity: 0,
      image_url: '',
      origin: '',
      shelf_life_days: '' as string | number | '',
      is_available: true,
      is_active: true,
      inventory_mode: 'piece' as InventoryMode,
      cart_step_grams: 250,
    };
  }
  const mode: InventoryMode = p.inventory_unit === 'weight_gram' ? 'weight_gram' : 'piece';
  return {
    category_id: p.category_id,
    name: p.name,
    description: p.description || '',
    price: p.price,
    weight_gram: p.weight_gram,
    unit: p.unit || 'шт',
    stock_quantity: p.stock_quantity,
    image_url: p.image_url || '',
    origin: p.origin || '',
    shelf_life_days: p.shelf_life_days ?? '',
    is_available: p.is_available,
    is_active: p.is_active,
    inventory_mode: mode,
    cart_step_grams: p.cart_step_grams && p.cart_step_grams > 0 ? p.cart_step_grams : 250,
  };
}

type ProductFormState = ReturnType<typeof productToForm>;

export function AdminProducts() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { accessToken } = useAuthStore();

  const [stores, setStores] = useState<StoreRow[]>([]);
  const [categories, setCategories] = useState<CategoryRow[]>([]);
  const [products, setProducts] = useState<ProductRow[]>([]);
  const [filterCategoryId, setFilterCategoryId] = useState('');
  const [loadingMeta, setLoadingMeta] = useState(true);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [error, setError] = useState('');

  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ProductFormState>(() => productToForm(null));
  const [saving, setSaving] = useState(false);
  const [imageUploading, setImageUploading] = useState(false);

  const storeId = searchParams.get('store') || '';

  const setStoreId = (id: string) => {
    setSearchParams(id ? { store: id } : {}, { replace: true });
  };

  const loadProducts = useCallback(async () => {
    if (!accessToken || !storeId) return;
    setLoadingProducts(true);
    setError('');
    try {
      const r = await adminAPI.listProducts(storeId, filterCategoryId || undefined);
      setProducts((r.data as { data?: ProductRow[] })?.data || []);
    } catch {
      setError('Не удалось загрузить товары');
      setProducts([]);
    } finally {
      setLoadingProducts(false);
    }
  }, [accessToken, storeId, filterCategoryId]);

  useEffect(() => {
    if (!accessToken) {
      return;
    }
    let cancelled = false;
    setLoadingMeta(true);
    Promise.all([adminAPI.getStores(), adminAPI.listCategories()])
      .then(([st, cat]) => {
        if (cancelled) return;
        const stList = (st.data as { data?: StoreRow[] })?.data || [];
        const catList = (cat.data as { data?: CategoryRow[] })?.data || [];
        setStores(stList);
        setCategories(catList);
        setSearchParams(
          (prev) => {
            if (stList.length === 0) return prev;
            if (prev.get('store')) return prev;
            const next = new URLSearchParams(prev);
            next.set('store', stList[0].id);
            return next;
          },
          { replace: true }
        );
      })
      .catch(() => navigate('/login?next=' + encodeURIComponent('/admin/products')))
      .finally(() => {
        if (!cancelled) setLoadingMeta(false);
      });
    return () => {
      cancelled = true;
    };
  }, [accessToken, navigate, setSearchParams]);

  useEffect(() => {
    if (!storeId) return;
    loadProducts();
  }, [storeId, loadProducts]);

  if (!accessToken) return null;

  const startCreate = () => {
    const f = productToForm(null);
    if (filterCategoryId) f.category_id = filterCategoryId;
    setForm(f);
    setEditingId('new');
  };

  const startEdit = (p: ProductRow) => {
    setForm(productToForm(p));
    setEditingId(p.id);
  };

  const cancelForm = () => {
    setEditingId(null);
    setForm(productToForm(null));
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId) return;
    const price = Number(form.price);
    if (!form.category_id || !form.name.trim() || !Number.isFinite(price) || price <= 0) {
      setError('Укажите категорию, название и цену > 0');
      return;
    }
    setSaving(true);
    setError('');
    try {
      if (editingId === 'new') {
        const payload: AdminProductPayload = {
          category_id: form.category_id,
          name: form.name.trim(),
          description: form.description,
          price,
          weight_gram: Number(form.weight_gram) || 0,
          unit: form.unit || 'шт',
          stock_quantity: Number(form.stock_quantity) || 0,
          image_url: form.image_url,
          origin: form.origin,
          is_available: form.is_available,
          is_active: form.is_active,
          inventory_unit: form.inventory_mode,
          cart_step_grams:
            form.inventory_mode === 'weight_gram' ? Math.max(50, Number(form.cart_step_grams) || 250) : undefined,
        };
        const sl = form.shelf_life_days;
        if (sl !== '' && sl !== null && sl !== undefined && Number(sl) >= 0) {
          payload.shelf_life_days = Number(sl);
        }
        await adminAPI.createProduct(storeId, payload);
      } else if (editingId) {
        const orig = products.find((x) => x.id === editingId);
        const patch: AdminProductPatch = {
          category_id: form.category_id,
          name: form.name.trim(),
          description: form.description,
          price,
          weight_gram: Number(form.weight_gram) || 0,
          unit: form.unit || 'шт',
          stock_quantity: Number(form.stock_quantity) || 0,
          image_url: form.image_url,
          origin: form.origin,
          is_available: form.is_available,
          is_active: form.is_active,
          inventory_unit: form.inventory_mode,
          cart_step_grams:
            form.inventory_mode === 'weight_gram' ? Math.max(50, Number(form.cart_step_grams) || 250) : undefined,
        };
        const sl = form.shelf_life_days;
        const slEmpty = sl === '' || sl === null || (typeof sl === 'string' && sl.trim() === '');
        if (slEmpty) {
          if (orig != null && orig.shelf_life_days != null) {
            patch.clear_shelf_life = true;
          }
        } else if (Number(sl) >= 0) {
          patch.shelf_life_days = Number(sl);
        }
        await adminAPI.patchProduct(editingId, patch);
      }
      cancelForm();
      await loadProducts();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка сохранения');
    } finally {
      setSaving(false);
    }
  };

  const onProductImageSelected = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    setImageUploading(true);
    setError('');
    try {
      const res = await adminAPI.uploadProductImage(file);
      const url = (res.data as { data?: { url?: string } })?.data?.url;
      if (url) setForm((f) => ({ ...f, image_url: url }));
      else setError('Сервер не вернул ссылку на файл');
    } catch {
      setError('Не удалось загрузить изображение (проверьте UPLOAD_DIR на сервере)');
    } finally {
      setImageUploading(false);
    }
  };

  const deactivate = async (p: ProductRow) => {
    if (!window.confirm(`Скрыть товар «${p.name}» из каталога?`)) return;
    setError('');
    try {
      await adminAPI.deactivateProduct(p.id);
      await loadProducts();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка');
    }
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Товары</h1>
            <p className="text-slate-400 text-sm mt-1">Позиции по выбранному магазину</p>
          </div>
          <button
            type="button"
            onClick={startCreate}
            disabled={!storeId || categories.length === 0}
            className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90 disabled:opacity-40"
          >
            + Новый товар
          </button>
        </div>

        <div className="flex flex-wrap gap-4 items-end rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-4">
          <label className="block min-w-[200px]">
            <span className="text-xs text-[var(--admin-text-muted)]">Магазин</span>
            <select
              value={storeId}
              onChange={(e) => setStoreId(e.target.value)}
              disabled={loadingMeta || stores.length === 0}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            >
              {stores.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
          <label className="block min-w-[200px]">
            <span className="text-xs text-[var(--admin-text-muted)]">Фильтр по категории</span>
            <select
              value={filterCategoryId}
              onChange={(e) => setFilterCategoryId(e.target.value)}
              className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
            >
              <option value="">Все</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">{error}</div>
        )}

        {editingId && (
          <form
            onSubmit={submit}
            className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
          >
            <h2 className="font-semibold text-[var(--admin-text-primary)]">
              {editingId === 'new' ? 'Новый товар' : 'Редактирование'}
            </h2>
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Категория *</span>
                <select
                  value={form.category_id}
                  onChange={(e) => setForm((f) => ({ ...f, category_id: e.target.value }))}
                  required
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                >
                  <option value="">—</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </label>
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
                <span className="text-xs text-[var(--admin-text-muted)]">Описание</span>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  rows={2}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Как продаём *</span>
                <div className="mt-2 flex flex-wrap gap-4 text-sm text-[var(--admin-text-primary)]">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="inv_mode"
                      checked={form.inventory_mode === 'piece'}
                      onChange={() => setForm((f) => ({ ...f, inventory_mode: 'piece' }))}
                    />
                    Поштучно (шт, фикс. вес в карточке — опционально)
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="inv_mode"
                      checked={form.inventory_mode === 'weight_gram'}
                      onChange={() => setForm((f) => ({ ...f, inventory_mode: 'weight_gram' }))}
                    />
                    На развес (кг): цена за 1 кг, остаток и корзина в граммах
                  </label>
                </div>
              </label>
              {form.inventory_mode === 'weight_gram' && (
                <label className="block">
                  <span className="text-xs text-[var(--admin-text-muted)]">Шаг в корзине (г)</span>
                  <input
                    type="number"
                    min={50}
                    step={50}
                    value={form.cart_step_grams}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, cart_step_grams: Number(e.target.value) || 250 }))
                    }
                    className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                  />
                  <span className="text-xs text-slate-500">500 г — половина кг за шаг; 250 г — более мелкий шаг</span>
                </label>
              )}
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Цена (₸) *</span>
                <input
                  type="number"
                  min={1}
                  value={form.price}
                  onChange={(e) => setForm((f) => ({ ...f, price: e.target.value === '' ? '' : Number(e.target.value) }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                  required
                />
                {form.inventory_mode === 'weight_gram' && (
                  <span className="text-xs text-slate-500">за 1 кг</span>
                )}
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Вес (г)</span>
                <input
                  type="number"
                  min={0}
                  value={form.weight_gram}
                  onChange={(e) => setForm((f) => ({ ...f, weight_gram: e.target.value === '' ? '' : Number(e.target.value) }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Единица</span>
                <input
                  value={form.unit}
                  onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Остаток на складе</span>
                <input
                  type="number"
                  min={0}
                  value={form.stock_quantity}
                  onChange={(e) => setForm((f) => ({ ...f, stock_quantity: Number(e.target.value) }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <div className="block sm:col-span-2 space-y-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Фото товара</span>
                <div className="flex flex-wrap items-center gap-3">
                  <label className="inline-flex cursor-pointer items-center rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm text-[var(--admin-text-primary)] hover:border-[var(--admin-accent)]">
                    <input
                      type="file"
                      accept="image/jpeg,image/png,image/webp"
                      className="sr-only"
                      disabled={imageUploading}
                      onChange={(e) => void onProductImageSelected(e)}
                    />
                    {imageUploading ? 'Загрузка…' : 'Выбрать файл'}
                  </label>
                  {form.image_url && (
                    <img
                      src={form.image_url}
                      alt=""
                      className="h-16 w-16 rounded object-cover border border-[var(--admin-border)]"
                    />
                  )}
                </div>
                <label className="block">
                  <span className="text-xs text-[var(--admin-text-muted)]">Или URL картинки</span>
                  <input
                    value={form.image_url}
                    onChange={(e) => setForm((f) => ({ ...f, image_url: e.target.value }))}
                    placeholder="https://… или /uploads/…"
                    className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm text-[var(--admin-text-primary)]"
                  />
                </label>
              </div>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Происхождение</span>
                <input
                  value={form.origin}
                  onChange={(e) => setForm((f) => ({ ...f, origin: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Срок годности (дней)</span>
                <input
                  type="number"
                  min={0}
                  value={form.shelf_life_days === '' ? '' : form.shelf_life_days}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, shelf_life_days: e.target.value === '' ? '' : Number(e.target.value) }))
                  }
                  placeholder="пусто = не указано"
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={form.is_available}
                  onChange={(e) => setForm((f) => ({ ...f, is_available: e.target.checked }))}
                />
                <span className="text-sm text-[var(--admin-text-muted)]">Доступен к заказу</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={form.is_active}
                  onChange={(e) => setForm((f) => ({ ...f, is_active: e.target.checked }))}
                />
                <span className="text-sm text-[var(--admin-text-muted)]">Активен в каталоге</span>
              </label>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={saving}
                className="rounded-lg bg-[var(--admin-accent)] px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
              >
                {saving ? 'Сохранение...' : 'Сохранить'}
              </button>
              <button type="button" onClick={cancelForm} className="rounded-lg px-4 py-2 text-sm text-[var(--admin-text-muted)]">
                Отмена
              </button>
            </div>
          </form>
        )}

        {!storeId && !loadingMeta ? (
          <p className="text-[var(--admin-text-muted)]">Сначала создайте магазин во вкладке «Магазины».</p>
        ) : loadingProducts && products.length === 0 ? (
          <p className="text-[var(--admin-text-muted)]">Загрузка товаров...</p>
        ) : (
          <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[880px]">
                <thead>
                  <tr className="border-b border-[var(--admin-border)]">
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Название</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Учёт</th>
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Категория</th>
                    <th className="p-4 text-right text-sm font-medium text-[var(--admin-text-muted)]">Цена</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Склад (физ / рез)</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Витрина</th>
                    <th className="p-4 text-right text-sm font-medium text-[var(--admin-text-muted)]"></th>
                  </tr>
                </thead>
                <tbody>
                  {products.map((p) => (
                    <tr key={p.id} className="border-b border-[var(--admin-border)]/50 hover:bg-[var(--admin-bg-elevated)]/40">
                      <td className="p-4 font-medium text-[var(--admin-text-primary)]">{p.name}</td>
                      <td className="p-4 text-center text-sm text-[var(--admin-text-muted)]">
                        {p.inventory_unit === 'weight_gram' ? (
                          <span>
                            кг
                            {p.cart_step_grams ? (
                              <span className="block text-xs text-slate-500">шаг {p.cart_step_grams} г</span>
                            ) : null}
                          </span>
                        ) : (
                          'шт'
                        )}
                      </td>
                      <td className="p-4 text-[var(--admin-text-muted)]">{p.category?.name || '—'}</td>
                      <td className="p-4 text-right text-[var(--admin-accent)] font-medium">{p.price} ₸</td>
                      <td className="p-4 text-center text-sm text-[var(--admin-text-primary)]">
                        {p.stock_quantity}
                        {typeof p.stock_reserved === 'number' && p.stock_reserved > 0 ? (
                          <span className="text-amber-400"> / −{p.stock_reserved}</span>
                        ) : null}
                        {p.inventory_unit === 'weight_gram' ? <span className="block text-xs text-slate-500">вес, г</span> : null}
                      </td>
                      <td className="p-4 text-center text-sm text-[var(--admin-text-muted)]">
                        {p.is_active && p.is_available ? 'да' : 'нет'}
                      </td>
                      <td className="p-4 text-right space-x-2 whitespace-nowrap">
                        <button
                          type="button"
                          onClick={() => startEdit(p)}
                          className="text-sm text-[var(--admin-accent)] hover:opacity-90"
                        >
                          Изменить
                        </button>
                        <button type="button" onClick={() => deactivate(p)} className="text-sm text-red-400 hover:text-red-300">
                          Скрыть
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {storeId && !loadingProducts && products.length === 0 && (
              <p className="p-8 text-center text-[var(--admin-text-muted)]">В этом магазине пока нет товаров</p>
            )}
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
