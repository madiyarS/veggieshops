import axios, { type InternalAxiosRequestConfig } from 'axios';

const API_URL = import.meta.env.VITE_API_URL || '/api/v1';

/** /api/v1 → /api для запросов к /api/v2 */
function apiRootWithoutV1(): string {
  const u = import.meta.env.VITE_API_URL || '/api/v1';
  return u.replace(/\/?v1\/?$/, '') || '/api';
}

export const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
});

export const apiV2 = axios.create({
  baseURL: `${apiRootWithoutV1()}/v2`,
  headers: { 'Content-Type': 'application/json' },
});

function attachAuthBearer(config: InternalAxiosRequestConfig) {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  return config;
}

api.interceptors.request.use(attachAuthBearer);
apiV2.interceptors.request.use(attachAuthBearer);

api.interceptors.request.use((config) => {
  if (config.data instanceof FormData) {
    config.headers.delete('Content-Type');
  }
  return config;
});

export const authAPI = {
  register: (data: { phone: string; password: string; first_name?: string; last_name?: string }) =>
    api.post('/auth/register', data),
  login: (data: { phone: string; password: string }) => api.post('/auth/login', data),
  refresh: (refreshToken: string) => api.post('/auth/refresh', { refresh_token: refreshToken }),
};

export const courierAPI = {
  listOrders: () => api.get('/courier/orders'),
  accept: (orderId: string) => api.post(`/courier/orders/${orderId}/accept`),
  complete: (orderId: string, code: string) => api.post(`/courier/orders/${orderId}/complete`, { code }),
};

export const storesAPI = {
  getAll: () => api.get('/stores'),
  getById: (id: string) => api.get(`/stores/${id}`),
  getDistricts: (storeId: string) => api.get(`/stores/${storeId}/districts`),
  getTimeSlots: (storeId: string, date: string) => api.get(`/stores/${storeId}/time-slots?date=${date}`),
};

export type ProductListQuery = {
  categoryId?: string;
  q?: string;
  inStockOnly?: boolean;
  sort?: '' | 'name' | 'price_asc' | 'price_desc' | 'expiry_asc';
};

export const productsAPI = {
  getByStore: (storeId: string, opts?: ProductListQuery) => {
    const p = new URLSearchParams({ store_id: storeId });
    if (opts?.categoryId) p.set('category_id', opts.categoryId);
    const t = (opts?.q || '').trim();
    if (t) p.set('q', t);
    if (opts?.inStockOnly) p.set('in_stock_only', 'true');
    if (opts?.sort) p.set('sort', opts.sort);
    return api.get(`/products?${p}`);
  },
  /** Доступный остаток (шт или граммы) по списку товаров, до 50 id. */
  getAvailability: (storeId: string, productIds: string[]) => {
    const p = new URLSearchParams({ store_id: storeId, product_ids: productIds.join(',') });
    return api.get(`/products/availability?${p}`);
  },
  getById: (id: string) => api.get(`/products/${id}`),
};

export const categoriesAPI = {
  getAll: () => api.get('/categories'),
};

export type AdminCategoryPayload = {
  name: string;
  description?: string;
  icon_url?: string;
  order?: number;
  is_active?: boolean;
};

export type AdminCategoryPatch = Partial<AdminCategoryPayload>;

export type AdminProductPayload = {
  category_id: string;
  name: string;
  description?: string;
  price: number;
  weight_gram?: number;
  unit?: string;
  stock_quantity?: number;
  image_url?: string;
  origin?: string;
  shelf_life_days?: number | null;
  is_available?: boolean;
  is_active?: boolean;
  inventory_unit?: 'piece' | 'weight_gram';
  package_grams?: number | null;
  is_seasonal?: boolean;
  temporarily_unavailable?: boolean;
  substitute_product_id?: string;
  reorder_min_qty?: number;
  cart_step_grams?: number;
};

export type AdminProductPatch = Partial<{
  category_id: string;
  name: string;
  description: string;
  price: number;
  weight_gram: number;
  unit: string;
  stock_quantity: number;
  image_url: string;
  origin: string;
  shelf_life_days: number | null;
  clear_shelf_life: boolean;
  is_available: boolean;
  is_active: boolean;
  inventory_unit: 'piece' | 'weight_gram';
  package_grams: number | null;
  clear_package_grams: boolean;
  is_seasonal: boolean;
  temporarily_unavailable: boolean;
  substitute_product_id: string;
  clear_substitute: boolean;
  reorder_min_qty: number;
  cart_step_grams: number;
}>;

export type AdminStorePatch = Partial<{
  name: string;
  description: string;
  address: string;
  latitude: number;
  longitude: number;
  phone: string;
  email: string;
  delivery_radius_km: number;
  min_order_amount: number;
  max_order_weight_kg: number | null;
  clear_max_weight: boolean;
  is_active: boolean;
  working_hours_start: string;
  working_hours_end: string;
  clear_working_hours_start: boolean;
  clear_working_hours_end: boolean;
}>;

export type AdminDistrictPayload = {
  name: string;
  distance_km: number;
  delivery_fee_regular: number;
  delivery_fee_express: number;
  is_active?: boolean;
  streets?: string[];
};

export type AdminDistrictPatch = Partial<AdminDistrictPayload> & {
  streets?: string[];
};

export type AdminTimeSlotPayload = {
  day_of_week: number;
  start_time: string;
  end_time: string;
  max_orders?: number;
  is_active?: boolean;
};

export type AdminTimeSlotPatch = Partial<{
  day_of_week: number;
  start_time: string;
  end_time: string;
  max_orders: number;
  is_active: boolean;
}>;

export const adminAPI = {
  getStores: () => api.get('/admin/stores'),
  patchStore: (id: string, data: AdminStorePatch) => api.patch(`/admin/stores/${id}`, data),
  createStore: (data: {
    name: string;
    address: string;
    latitude: number;
    longitude: number;
    description?: string;
    phone?: string;
    email?: string;
    delivery_radius_km?: number;
    min_order_amount?: number;
    max_order_weight_kg?: number | null;
    /** Скопировать товары и остатки склада из другого магазина (новые id). */
    copy_catalog_from_store_id?: string;
  }) => api.post('/admin/stores', data),
  getRevenueReport: (params?: { store_id?: string; from?: string; to?: string }) => {
    const q = new URLSearchParams();
    if (params?.store_id) q.set('store_id', params.store_id);
    if (params?.from) q.set('from', params.from);
    if (params?.to) q.set('to', params.to);
    const s = q.toString();
    return api.get(`/admin/analytics/revenue${s ? `?${s}` : ''}`);
  },
  listCategories: () => api.get('/admin/categories'),
  createCategory: (data: AdminCategoryPayload) => api.post('/admin/categories', data),
  patchCategory: (id: string, data: AdminCategoryPatch) => api.patch(`/admin/categories/${id}`, data),
  deleteCategory: (id: string) => api.delete(`/admin/categories/${id}`),
  listProducts: (storeId: string, categoryId?: string) =>
    api.get(`/admin/stores/${storeId}/products${categoryId ? `?category_id=${categoryId}` : ''}`),
  uploadProductImage: (file: File) => {
    const fd = new FormData();
    fd.append('file', file);
    return api.post('/admin/upload/product-image', fd);
  },
  createProduct: (storeId: string, data: AdminProductPayload) => api.post(`/admin/stores/${storeId}/products`, data),
  patchProduct: (id: string, data: AdminProductPatch) => api.patch(`/admin/products/${id}`, data),
  deactivateProduct: (id: string) => api.delete(`/admin/products/${id}`),
  listDistricts: (storeId: string) => api.get(`/admin/stores/${storeId}/districts`),
  createDistrict: (storeId: string, data: AdminDistrictPayload) =>
    api.post(`/admin/stores/${storeId}/districts`, data),
  patchDistrict: (id: string, data: AdminDistrictPatch) => api.patch(`/admin/districts/${id}`, data),
  deleteDistrict: (id: string) => api.delete(`/admin/districts/${id}`),
  listTimeSlotsAdmin: (storeId: string) => api.get(`/admin/stores/${storeId}/time-slots`),
  createTimeSlot: (storeId: string, data: AdminTimeSlotPayload) =>
    api.post(`/admin/stores/${storeId}/time-slots`, data),
  patchTimeSlot: (id: string, data: AdminTimeSlotPatch) => api.patch(`/admin/time-slots/${id}`, data),
  deleteTimeSlot: (id: string) => api.delete(`/admin/time-slots/${id}`),
  patchOrder: (
    id: string,
    data: { action: 'cancel_pending' | 'commit_stock' | 'cancel_delivery' }
  ) => api.patch(`/admin/orders/${id}`, data),
  stockZones: (storeId: string) => api.get(`/admin/stores/${storeId}/stock/zones`),
  stockExpiring: (storeId: string, days?: number) =>
    api.get(`/admin/stores/${storeId}/stock/expiring${days ? `?days=${days}` : ''}`),
  stockReorderAlerts: (storeId: string) => api.get(`/admin/stores/${storeId}/stock/reorder-alerts`),
  stockMovesJournal: (storeId: string, limit?: number) =>
    api.get(`/admin/stores/${storeId}/stock/moves-journal${limit != null ? `?limit=${limit}` : ''}`),
  stockReceiveSimple: (storeId: string, data: { product_id: string; quantity: number; note?: string }) =>
    api.post(`/admin/stores/${storeId}/stock/receive-simple`, data),
  stockSetActual: (storeId: string, data: { product_id: string; actual: number; note?: string }) =>
    api.post(`/admin/stores/${storeId}/stock/set-actual`, data),
  stockWriteOff: (
    storeId: string,
    data: { product_id: string; quantity: number; type: 'damage' | 'shrink' | 'resort'; reason?: string }
  ) => api.post(`/admin/stores/${storeId}/stock/write-off`, data),
  stockReceipt: (
    storeId: string,
    data: {
      supplier_id?: string;
      note?: string;
      lines: { product_id: string; zone_id: string; quantity: number; expires_at?: string }[];
    }
  ) => api.post(`/admin/stores/${storeId}/stock/receipt`, data),
  stockAuditComplete: (
    storeId: string,
    data: { note?: string; lines: { product_id: string; zone_id?: string; counted_qty: number }[] }
  ) => api.post(`/admin/stores/${storeId}/stock/audit/complete`, data),
  listSuppliers: (storeId: string) => api.get(`/admin/stores/${storeId}/suppliers`),
  createSupplier: (storeId: string, data: { name: string; phone?: string }) =>
    api.post(`/admin/stores/${storeId}/suppliers`, data),
  listProductBatches: (storeId: string, productId: string) =>
    api.get(`/admin/stores/${storeId}/products/${productId}/batches`),
};

/** Пагинация и v2 (курсор в meta). */
export const adminV2API = {
  listProductsPaged: (params: {
    store_id: string;
    limit?: number;
    after_id?: string;
    category_id?: string;
    q?: string;
  }) => {
    const q = new URLSearchParams({ store_id: params.store_id });
    if (params.limit != null) q.set('limit', String(params.limit));
    if (params.after_id) q.set('after_id', params.after_id);
    if (params.category_id) q.set('category_id', params.category_id);
    const search = (params.q || '').trim();
    if (search) q.set('q', search);
    return apiV2.get(`/products?${q}`);
  },
  listOrdersPaged: (
    storeId: string,
    params?: {
      limit?: number;
      after_created_at?: string;
      after_id?: string;
      status?: string;
      date_from?: string;
      date_to?: string;
    }
  ) => {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.after_created_at) q.set('after_created_at', params.after_created_at);
    if (params?.after_id) q.set('after_id', params.after_id);
    if (params?.status) q.set('status', params.status);
    if (params?.date_from) q.set('date_from', params.date_from);
    if (params?.date_to) q.set('date_to', params.date_to);
    const s = q.toString();
    return apiV2.get(`/admin/stores/${storeId}/orders${s ? `?${s}` : ''}`);
  },
  listStockMovementsPaged: (
    storeId: string,
    params?: { limit?: number; after_created_at?: string; after_id?: string }
  ) => {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.after_created_at) q.set('after_created_at', params.after_created_at);
    if (params?.after_id) q.set('after_id', params.after_id);
    const s = q.toString();
    return apiV2.get(`/admin/stores/${storeId}/stock/movements${s ? `?${s}` : ''}`);
  },
};

export const ordersAPI = {
  checkDelivery: (data: { store_id: string; latitude: number; longitude: number; address?: string }) =>
    api.post('/orders/check-delivery', data),
  create: (data: {
    store_id: string;
    district_id: string;
    delivery_type: string;
    delivery_time_slot_id: string;
    delivery_address: string;
    customer_phone: string;
    customer_name: string;
    payment_method: string;
    items: { product_id: string; quantity: number }[];
    notes?: string;
  }) => api.post('/orders', data),
  listMine: (params?: { limit?: number }) => {
    const q = params?.limit != null ? `?limit=${params.limit}` : '';
    return api.get(`/orders/mine${q}`);
  },
  getMine: (id: string) => api.get(`/orders/${id}`),
};
