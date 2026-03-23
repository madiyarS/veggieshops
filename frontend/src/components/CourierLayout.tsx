import { Link } from 'react-router-dom';
import { LogOut, Package } from 'lucide-react';
import { useAuthStore } from '../store/useAuthStore';

export function CourierLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuthStore();

  return (
    <div
      className="admin-root dark min-h-screen bg-[var(--admin-bg-base)] text-[var(--admin-text-primary)]"
      data-admin-theme="dark"
    >
      <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-4">
        <div className="flex items-center gap-2">
          <Package className="h-5 w-5 text-[var(--admin-accent)]" />
          <span className="font-semibold text-white">Курьер</span>
          <span className="text-sm text-[var(--admin-text-muted)] truncate max-w-[40vw]">
            {user?.first_name || user?.phone}
          </span>
        </div>
        <div className="flex items-center gap-4">
          <Link to="/" className="text-sm text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]">
            Сайт
          </Link>
          <button
            type="button"
            onClick={logout}
            className="flex items-center gap-1 text-sm text-red-400 hover:text-red-300"
          >
            <LogOut className="h-4 w-4" />
            Выйти
          </button>
        </div>
      </header>
      <div className="p-4 sm:p-6 max-w-4xl mx-auto">{children}</div>
    </div>
  );
}
