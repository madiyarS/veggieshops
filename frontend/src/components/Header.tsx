import { Link } from 'react-router-dom';
import { useCartStore } from '../store/useCartStore';
import { useAuthStore } from '../store/useAuthStore';

export function Header() {
  const items = useCartStore((s) => s.items);
  const count = items.reduce((sum, i) => sum + i.quantity, 0);
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const name = [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.phone;

  return (
    <header className="bg-veggie-green text-white shadow">
      <div className="container mx-auto px-4 py-4 flex flex-wrap justify-between items-center gap-3">
        <Link to="/" className="text-xl font-bold">
          🥬 VeggieShops.kz
        </Link>
        <nav className="flex flex-wrap gap-4 sm:gap-6 items-center text-sm sm:text-base">
          <Link to="/" className="hover:underline">
            Главная
          </Link>
          {(!user || user.role === 'customer') && (
            <Link
              to={user?.role === 'customer' ? '/orders' : '/login?next=%2Forders'}
              className="hover:underline"
            >
              Мои заказы
            </Link>
          )}
          {(!user || user.role === 'customer') && (
            <Link to="/cart" className="hover:underline flex items-center gap-1">
              Корзина {count > 0 && <span className="bg-veggie-light px-2 rounded">{count}</span>}
            </Link>
          )}
          {user ? (
            <>
              <span className="text-veggie-light truncate max-w-[140px]" title={name}>
                {name}
              </span>
              {(user.role === 'admin' || user.role === 'manager') && (
                <Link to="/admin/stores" className="hover:underline text-veggie-light">
                  Админка
                </Link>
              )}
              {user.role === 'courier' && (
                <Link to="/courier" className="hover:underline text-veggie-light">
                  Курьер
                </Link>
              )}
              <button type="button" onClick={logout} className="hover:underline text-veggie-light">
                Выйти
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="hover:underline text-veggie-light">
                Войти
              </Link>
              <Link to="/register" className="hover:underline text-veggie-light">
                Регистрация
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
