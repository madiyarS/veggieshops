import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { productsAPI, categoriesAPI, storesAPI } from '../services/api';
import { useCartStore, type InventoryUnitClient } from '../store/useCartStore';
import { useAuthStore } from '../store/useAuthStore';

const STOCK_WATCH_KEY = 'veggieshops_kz_stock_watch_v1';

interface Product {
  id: string;
  name: string;
  price: number;
  unit: string;
  stock_quantity: number;
  image_url?: string;
  inventory_unit?: InventoryUnitClient;
  temporarily_unavailable?: boolean;
  is_seasonal?: boolean;
  cart_step_grams?: number;
  substitute?: { id: string; name: string };
  nearest_batch_expires_at?: string;
  catalog_low_stock?: boolean;
}

interface Category {
  id: string;
  name: string;
}

type StockWatchEntry = { storeId: string; productId: string; name: string };

function readStockWatch(): StockWatchEntry[] {
  try {
    const raw = localStorage.getItem(STOCK_WATCH_KEY);
    if (!raw) return [];
    const a = JSON.parse(raw) as unknown;
    if (!Array.isArray(a)) return [];
    return a.filter(
      (x): x is StockWatchEntry =>
        x && typeof x === 'object' && 'storeId' in x && 'productId' in x && 'name' in x
    );
  } catch {
    return [];
  }
}

function writeStockWatch(entries: StockWatchEntry[]) {
  localStorage.setItem(STOCK_WATCH_KEY, JSON.stringify(entries));
}

function formatStock(p: Product): string {
  if (p.inventory_unit === 'weight_gram') {
    const kg = p.stock_quantity / 1000;
    const decimals = kg < 10 && kg % 1 !== 0 ? 2 : 1;
    return `${kg.toFixed(decimals)} кг`;
  }
  return `${p.stock_quantity} ${p.unit}`;
}

function formatExpiryLabel(iso?: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

export function Catalog() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);
  const storeId = searchParams.get('store') || '';
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryId, setCategoryId] = useState<string>('');
  const [nameSearch, setNameSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [inStockOnly, setInStockOnly] = useState(false);
  const [sort, setSort] = useState<'' | 'name' | 'price_asc' | 'price_desc' | 'expiry_asc'>('');
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [storeName, setStoreName] = useState('');
  const [minOrderAmount, setMinOrderAmount] = useState(0);
  const [restockBanner, setRestockBanner] = useState<string[]>([]);
  const [kgDraft, setKgDraft] = useState<Record<string, string>>({});
  const addItem = useCartStore((s) => s.addItem);

  useEffect(() => {
    if (!storeId) {
      setStoreName('');
      setMinOrderAmount(0);
      return;
    }
    storesAPI
      .getById(storeId)
      .then((r) => {
        const sd = (r.data as { data?: { name?: string; min_order_amount?: number } })?.data;
        setStoreName(sd?.name || '');
        setMinOrderAmount(typeof sd?.min_order_amount === 'number' ? sd.min_order_amount : 0);
      })
      .catch(() => {
        setStoreName('');
        setMinOrderAmount(0);
      });
  }, [storeId]);

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(nameSearch.trim()), 320);
    return () => clearTimeout(t);
  }, [nameSearch]);

  useEffect(() => {
    if (!storeId) return;
    categoriesAPI
      .getAll()
      .then((cat) => setCategories((cat.data as { data?: Category[] })?.data || []))
      .catch(() => setCategories([]));
  }, [storeId]);

  useEffect(() => {
    if (!storeId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setLoadError('');
    productsAPI
      .getByStore(storeId, {
        categoryId: categoryId || undefined,
        q: debouncedSearch || undefined,
        inStockOnly,
        sort: sort || undefined,
      })
      .then((pr) => {
        const list = (pr.data as { data?: Product[] })?.data || [];
        setProducts(list);
        const allWatch = readStockWatch();
        const forStore = allWatch.filter((w) => w.storeId === storeId);
        const otherStores = allWatch.filter((w) => w.storeId !== storeId);
        const stay: StockWatchEntry[] = [];
        const back: string[] = [];
        for (const w of forStore) {
          const prod = list.find((x) => x.id === w.productId);
          if (prod && prod.stock_quantity > 0 && !prod.temporarily_unavailable) {
            back.push(w.name);
          } else {
            stay.push(w);
          }
        }
        writeStockWatch([...otherStores, ...stay]);
        if (back.length) setRestockBanner(back);
      })
      .catch(() => {
        setLoadError('Не удалось загрузить каталог. Обновите страницу или попробуйте позже.');
        setProducts([]);
      })
      .finally(() => setLoading(false));
  }, [storeId, categoryId, debouncedSearch, inStockOnly, sort]);

  const defaultKg = (p: Product) => {
    const step = p.cart_step_grams || 250;
    return (step / 1000).toFixed(2).replace(/\.?0+$/, '') || String(step / 1000);
  };

  const getKgValue = (p: Product) => {
    const raw = kgDraft[p.id];
    if (raw !== undefined && raw !== '') return raw;
    return defaultKg(p);
  };

  const handleAddToCart = (p: Product, kgOverride?: number) => {
    if (!storeId || !storeName) return;
    if (!accessToken) {
      navigate(`/login?next=${encodeURIComponent(`/catalog?store=${storeId}`)}`);
      return;
    }
    if (!user || user.role !== 'customer') {
      window.alert(
        'Оформление заказа доступно только покупателям. Выйдите из аккаунта сотрудника или зарегистрируйте отдельный номер.'
      );
      return;
    }
    if (p.temporarily_unavailable || p.stock_quantity <= 0) return;
    const cart = useCartStore.getState();
    if (cart.items.length > 0 && cart.storeId && cart.storeId !== storeId) {
      if (
        !window.confirm('В корзине товары другого магазина. Очистить корзину и добавить этот товар?')
      ) {
        return;
      }
      cart.clear();
    }
    const inv: InventoryUnitClient = p.inventory_unit === 'weight_gram' ? 'weight_gram' : 'piece';
    const step = inv === 'weight_gram' ? p.cart_step_grams || 250 : 1;
    let qty = step;
    if (inv === 'weight_gram') {
      const kgStr = kgOverride !== undefined ? String(kgOverride) : getKgValue(p).replace(',', '.');
      const kg = parseFloat(kgStr);
      if (!Number.isFinite(kg) || kg <= 0) {
        window.alert('Укажите вес в килограммах (например 2.5)');
        return;
      }
      qty = Math.round(kg * 1000);
      qty = Math.max(1, Math.min(qty, p.stock_quantity));
    }
    if (inv === 'weight_gram' && p.stock_quantity < qty) return;
    addItem({
      productId: p.id,
      name: p.name,
      price: p.price,
      unit: inv === 'weight_gram' ? 'кг' : p.unit,
      inventoryUnit: inv,
      cartStepGrams: p.cart_step_grams || 250,
      quantity: qty,
      storeId,
      storeName,
      minOrderAmount,
    });
  };

  const toggleStockWatch = (p: Product) => {
    if (p.stock_quantity > 0) return;
    const all = readStockWatch();
    const exists = all.some((w) => w.storeId === storeId && w.productId === p.id);
    if (exists) {
      writeStockWatch(all.filter((w) => !(w.storeId === storeId && w.productId === p.id)));
      return;
    }
    writeStockWatch([...all, { storeId, productId: p.id, name: p.name }]);
  };

  const isWatching = (p: Product) =>
    readStockWatch().some((w) => w.storeId === storeId && w.productId === p.id);

  const sortOptions = [
    { v: '' as const, label: 'Как в магазине' },
    { v: 'name' as const, label: 'По названию' },
    { v: 'price_asc' as const, label: 'Цена ↑' },
    { v: 'price_desc' as const, label: 'Цена ↓' },
    { v: 'expiry_asc' as const, label: 'Срок годности' },
  ] as const;

  if (!storeId) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-600">Выберите магазин на главной странице</p>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-veggie-green mb-2">Каталог товаров</h2>
      {storeName && (
        <p className="text-gray-600 mb-1">
          Магазин: <span className="font-medium text-gray-800">{storeName}</span>
          {minOrderAmount > 0 && (
            <span className="ml-2">
              · Мин. сумма заказа (товары): <span className="font-medium">{minOrderAmount} ₸</span>
            </span>
          )}
        </p>
      )}
      <p className="text-sm text-gray-500 mb-4">
        Цены весовых товаров — за 1 кг; в корзину вес указывается в килограммах (можно дробно, например 2.5).
      </p>
      {restockBanner.length > 0 && (
        <div
          className="mb-4 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-900"
          role="status"
        >
          <p className="font-medium">Снова в наличии</p>
          <p>{restockBanner.join(', ')}</p>
          <button
            type="button"
            className="mt-2 text-green-800 underline text-xs"
            onClick={() => setRestockBanner([])}
          >
            Скрыть
          </button>
        </div>
      )}
      <label className="block max-w-md mb-4">
        <span className="text-xs text-gray-500">Поиск по названию</span>
        <input
          type="search"
          value={nameSearch}
          onChange={(e) => setNameSearch(e.target.value)}
          placeholder="Например: помидор"
          className="mt-1 w-full border border-gray-200 rounded-lg px-3 py-2 text-gray-900 bg-white"
        />
      </label>
      <div className="flex flex-wrap gap-3 items-center mb-6">
        <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
          <input
            type="checkbox"
            checked={inStockOnly}
            onChange={(e) => setInStockOnly(e.target.checked)}
          />
          Только в наличии
        </label>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <span className="text-gray-500">Сортировка</span>
          <select
            value={sort}
            onChange={(e) => setSort(e.target.value as typeof sort)}
            className="border border-gray-200 rounded-lg px-2 py-1.5 bg-white text-gray-900"
          >
            {sortOptions.map((o) => (
              <option key={o.v || 'default'} value={o.v}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
      </div>
      {categories.length > 0 && (
        <div className="flex gap-2 mb-6 flex-wrap">
          <button
            onClick={() => setCategoryId('')}
            className={`px-4 py-2 rounded ${!categoryId ? 'bg-veggie-green text-white' : 'bg-gray-200'}`}
          >
            Все
          </button>
          {categories.map((c) => (
            <button
              key={c.id}
              onClick={() => setCategoryId(c.id)}
              className={`px-4 py-2 rounded ${categoryId === c.id ? 'bg-veggie-green text-white' : 'bg-gray-200'}`}
            >
              {c.name}
            </button>
          ))}
        </div>
      )}
      {loadError && (
        <p className="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{loadError}</p>
      )}
      {loading && <p>Загрузка...</p>}
      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {products.map((p) => {
          const isWeight = p.inventory_unit === 'weight_gram';
          const step = isWeight ? p.cart_step_grams || 250 : 1;
          const canAdd = !p.temporarily_unavailable && p.stock_quantity >= (isWeight ? 1 : step);
          const priceLabel = isWeight ? `${p.price} ₸ / кг` : `${p.price} ₸ / ${p.unit}`;
          const exp = formatExpiryLabel(p.nearest_batch_expires_at);
          return (
            <div key={p.id} className="bg-white rounded-lg shadow p-4 border relative">
              {p.catalog_low_stock && p.stock_quantity > 0 && (
                <span className="absolute top-2 right-2 text-[10px] font-semibold uppercase tracking-wide bg-amber-100 text-amber-900 px-2 py-0.5 rounded">
                  Мало
                </span>
              )}
              {p.image_url && <img src={p.image_url} alt={p.name} className="w-full h-32 object-cover rounded" />}
              <h3 className="font-semibold mt-2">{p.name}</h3>
              {p.is_seasonal && <p className="text-xs text-amber-700 mt-0.5">Сезонный товар</p>}
              <p className="text-veggie-green font-bold">{priceLabel}</p>
              <p className="text-sm text-gray-600 mt-1">
                {p.temporarily_unavailable ? (
                  <span className="text-amber-800">
                    Временно нет в продаже
                    {p.substitute && (
                      <>
                        {' · '}
                        <a href={`#product-${p.substitute.id}`} className="underline text-veggie-green">
                          Замена: {p.substitute.name}
                        </a>
                      </>
                    )}
                  </span>
                ) : p.stock_quantity > 0 ? (
                  <>
                    В наличии: <span className="font-medium text-gray-800">{formatStock(p)}</span>
                    {exp && (
                      <span className="block text-xs text-gray-500 mt-0.5">Ближайший срок: {exp}</span>
                    )}
                  </>
                ) : (
                  <span className="text-amber-700">Нет в наличии</span>
                )}
              </p>
              {isWeight && canAdd && (
                <label className="block mt-2 text-xs text-gray-500">
                  Вес, кг
                  <input
                    type="number"
                    inputMode="decimal"
                    min={0.001}
                    step={0.05}
                    max={p.stock_quantity / 1000}
                    value={getKgValue(p)}
                    onChange={(e) => setKgDraft((d) => ({ ...d, [p.id]: e.target.value }))}
                    className="mt-0.5 w-full border border-gray-200 rounded px-2 py-1.5 text-gray-900"
                  />
                </label>
              )}
              <button
                id={`product-${p.id}`}
                onClick={() => handleAddToCart(p)}
                disabled={!canAdd}
                className="mt-2 w-full bg-veggie-green text-white py-2 rounded hover:bg-veggie-dark disabled:opacity-50"
              >
                {isWeight ? 'В корзину' : 'В корзину'}
              </button>
              {!isWeight && step > 1 && (
                <p className="text-[10px] text-gray-400 mt-1">Шаг: {step} {p.unit}</p>
              )}
              {p.stock_quantity <= 0 && !p.temporarily_unavailable && (
                <button
                  type="button"
                  onClick={() => {
                    toggleStockWatch(p);
                    setProducts((prev) => [...prev]);
                  }}
                  className="mt-2 w-full text-sm text-veggie-green underline"
                >
                  {isWatching(p) ? 'Убрать из напоминаний' : 'Напомнить при появлении'}
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
