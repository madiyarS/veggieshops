import { create } from 'zustand';

export type InventoryUnitClient = 'piece' | 'weight_gram';

export interface CartItem {
  productId: string;
  name: string;
  price: number;
  quantity: number;
  unit: string;
  /** По умолчанию штуки; для weight_gram quantity — граммы, price — за 1 кг */
  inventoryUnit?: InventoryUnitClient;
  cartStepGrams?: number;
}

export interface CartItemInput extends Omit<CartItem, 'quantity'> {
  quantity?: number;
  /** Обязательно при непустой корзине; при смене магазина вызывайте clear() или подтверждение в UI */
  storeId: string;
  storeName: string;
  minOrderAmount: number;
}

function lineTotal(i: CartItem): number {
  if (i.inventoryUnit === 'weight_gram') {
    const t = Math.round((i.price * i.quantity) / 1000);
    return t > 0 || i.quantity <= 0 ? t : 1;
  }
  return i.price * i.quantity;
}

interface CartStore {
  items: CartItem[];
  storeId: string | null;
  storeName: string;
  minOrderAmount: number;
  /** Доступно к продаже (граммы для weight_gram, штуки для piece) — с сервера */
  availabilityByProductId: Record<string, number>;
  setAvailability: (m: Record<string, number>) => void;
  addItem: (item: CartItemInput) => void;
  removeItem: (productId: string) => void;
  updateQuantity: (productId: string, quantity: number) => void;
  clear: () => void;
  total: () => number;
  /** Сумма товаров достигла минимума для текущего магазина */
  meetsMinOrder: () => boolean;
}

export const useCartStore = create<CartStore>((set, get) => ({
  items: [],
  storeId: null,
  storeName: '',
  minOrderAmount: 0,
  availabilityByProductId: {},

  setAvailability: (m) =>
    set((state) => ({
      availabilityByProductId: { ...state.availabilityByProductId, ...m },
    })),

  addItem: (item) => {
    const step = item.inventoryUnit === 'weight_gram' ? item.cartStepGrams || 250 : 1;
    const qty = item.quantity ?? step;
    set((state) => {
      const maxA = state.availabilityByProductId[item.productId];
      const existing = state.items.find((i) => i.productId === item.productId);
      const nextMeta = {
        storeId: item.storeId,
        storeName: item.storeName,
        minOrderAmount: item.minOrderAmount,
      };
      const base: CartItem = {
        productId: item.productId,
        name: item.name,
        price: item.price,
        unit: item.unit,
        inventoryUnit: item.inventoryUnit,
        cartStepGrams: item.cartStepGrams,
        quantity: qty,
      };
      if (existing) {
        const addQ = item.inventoryUnit === 'weight_gram' ? step : item.quantity || 1;
        let newQ = existing.quantity + addQ;
        if (maxA !== undefined && maxA >= 0) {
          newQ = Math.min(newQ, maxA);
        }
        if (newQ <= 0) {
          return state;
        }
        return {
          ...nextMeta,
          items: state.items.map((i) =>
            i.productId === item.productId ? { ...i, quantity: newQ } : i
          ),
        };
      }
      let firstQ = qty;
      if (maxA !== undefined && maxA >= 0) {
        firstQ = Math.min(firstQ, maxA);
      }
      if (firstQ <= 0) {
        return state;
      }
      return {
        ...nextMeta,
        items: [...state.items, { ...base, quantity: firstQ }],
      };
    });
  },

  removeItem: (productId) =>
    set((state) => {
      const next = state.items.filter((i) => i.productId !== productId);
      if (next.length === 0) {
        return {
          items: [],
          storeId: null,
          storeName: '',
          minOrderAmount: 0,
          availabilityByProductId: {},
        };
      }
      return { items: next };
    }),

  updateQuantity: (productId, quantity) => {
    if (quantity <= 0) {
      get().removeItem(productId);
      return;
    }
    set((state) => ({
      items: state.items.map((i) => {
        if (i.productId !== productId) return i;
        const maxA = state.availabilityByProductId[i.productId];
        if (i.inventoryUnit === 'weight_gram') {
          let q = Math.round(quantity);
          if (q < 1) q = 1;
          if (maxA !== undefined && maxA >= 0) q = Math.min(q, maxA);
          return { ...i, quantity: q };
        }
        let q = Math.floor(quantity);
        if (q < 1) q = 1;
        if (maxA !== undefined && maxA >= 0) q = Math.min(q, maxA);
        return { ...i, quantity: q };
      }),
    }));
  },

  clear: () =>
    set({
      items: [],
      storeId: null,
      storeName: '',
      minOrderAmount: 0,
      availabilityByProductId: {},
    }),

  total: () => get().items.reduce((sum, i) => sum + lineTotal(i), 0),

  meetsMinOrder: () => {
    const { items, minOrderAmount } = get();
    if (items.length === 0) return true;
    if (minOrderAmount <= 0) return true;
    return get().total() >= minOrderAmount;
  },
}));

export function cartLineTotal(i: CartItem): number {
  return lineTotal(i);
}
