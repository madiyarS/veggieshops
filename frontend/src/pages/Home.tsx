import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { storesAPI } from '../services/api';

interface Store {
  id: string;
  name: string;
  address: string;
  delivery_radius_km: number;
  min_order_amount: number;
}

export function Home() {
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    storesAPI.getAll()
      .then((r) => setStores((r.data as { data?: Store[] })?.data || []))
      .catch(() => setError('Магазины недоступны. Запустите backend с БД.'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <section className="mb-12 text-center">
        <h1 className="text-4xl font-bold text-veggie-green mb-4">Свежие овощи и фрукты с доставкой</h1>
        <p className="text-xl text-gray-600">Выберите магазин и закажите доставку на дом</p>
      </section>

      {loading && <p className="text-center">Загрузка...</p>}
      {error && <p className="text-center text-red-600">{error}</p>}

      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
        {stores.map((store) => (
          <div key={store.id} className="bg-white rounded-lg shadow p-6 border">
            <h3 className="text-lg font-semibold text-veggie-green">{store.name}</h3>
            <p className="text-gray-600 text-sm mt-1">{store.address}</p>
            <p className="text-sm mt-2">Доставка: в радиусе {store.delivery_radius_km} км</p>
            <p className="text-sm">Мин. заказ: {store.min_order_amount} ₸</p>
            <Link
              to={`/catalog?store=${store.id}`}
              className="mt-4 inline-block bg-veggie-green text-white px-4 py-2 rounded hover:bg-veggie-dark"
            >
              Перейти в каталог
            </Link>
          </div>
        ))}
      </div>

      {!loading && !error && stores.length === 0 && (
        <div className="text-center text-gray-500 py-12">
          <p>Магазины пока не добавлены.</p>
          <p className="mt-2">Запустите миграции БД и создайте магазин через API.</p>
        </div>
      )}
    </div>
  );
}
