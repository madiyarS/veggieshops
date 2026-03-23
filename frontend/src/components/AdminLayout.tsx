import { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  Store,
  ChevronLeft,
  ChevronRight,
  LogOut,
  LayoutDashboard,
  Tags,
  Package,
  TrendingUp,
  Truck,
  Warehouse,
} from 'lucide-react';
import { useAuthStore } from '../store/useAuthStore';

const NAV = [
  { to: '/admin/stores', label: 'Магазины', icon: Store },
  { to: '/admin/delivery', label: 'Доставка', icon: Truck },
  { to: '/admin/categories', label: 'Категории', icon: Tags },
  { to: '/admin/products', label: 'Товары', icon: Package },
  { to: '/admin/inventory', label: 'Склад', icon: Warehouse },
  { to: '/admin/reports', label: 'Выручка', icon: TrendingUp },
] as const;

const ADMIN_SECTION_LABEL: Record<string, string> = {
  '/admin/stores': 'Магазины',
  '/admin/delivery': 'Доставка',
  '/admin/categories': 'Категории',
  '/admin/products': 'Товары',
  '/admin/inventory': 'Склад',
  '/admin/reports': 'Выручка',
};

function adminSectionTitle(pathname: string): string {
  if (ADMIN_SECTION_LABEL[pathname]) return ADMIN_SECTION_LABEL[pathname];
  for (const [path, label] of Object.entries(ADMIN_SECTION_LABEL)) {
    if (path !== '/admin/stores' && pathname.startsWith(path)) return label;
  }
  return 'Панель';
}

function getInitials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '—';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function AdminLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const { user, logout } = useAuthStore();
  const sidebarW = collapsed ? 60 : 240;

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key === 'b') {
        e.preventDefault();
        setCollapsed((c) => !c);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  const displayName = [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.phone || 'Админ';

  return (
    <div
      className="admin-root dark min-h-screen bg-[var(--admin-bg-base)] text-[var(--admin-text-primary)]"
      data-admin-theme="dark"
      style={{ ['--sidebar-w' as string]: `${sidebarW}px` }}
    >
      <aside
        className={`fixed left-0 top-0 bottom-0 z-40 flex flex-col border-r border-[var(--admin-border)] bg-[var(--admin-bg-surface)] transition-all duration-300 hidden sm:flex ${
          collapsed ? 'w-[60px]' : 'w-[240px]'
        }`}
      >
        <div className="flex h-14 items-center justify-between border-b border-[var(--admin-border)] px-4">
          {!collapsed && (
            <Link
              to="/"
              className="text-sm font-medium text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]"
            >
              🥬 VeggieShops.kz
            </Link>
          )}
        </div>
        <nav className="flex-1 space-y-1 overflow-y-auto p-2">
          {NAV.map((item) => {
            const Icon = item.icon;
            const isActive =
              location.pathname === item.to ||
              (item.to !== '/admin/stores' && location.pathname.startsWith(item.to));
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'border-l-[3px] border-[var(--admin-accent)] bg-[var(--admin-bg-elevated)] text-[var(--admin-text-primary)]'
                    : 'border-l-[3px] border-transparent text-[var(--admin-text-muted)] hover:bg-[var(--admin-bg-elevated)] hover:text-[var(--admin-text-primary)]'
                } ${collapsed ? 'justify-center px-2' : ''}`}
              >
                <Icon className="h-[18px] w-[18px] shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            );
          })}
        </nav>
        <div className="border-t border-[var(--admin-border)] p-2">
          <button
            type="button"
            onClick={() => setCollapsed((c) => !c)}
            className="flex w-full items-center justify-center gap-2 rounded-lg py-2 text-[var(--admin-text-muted)] hover:bg-[var(--admin-bg-elevated)] hover:text-[var(--admin-text-primary)]"
          >
            {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
            {!collapsed && <span className="text-xs">Свернуть</span>}
          </button>
          {!collapsed && user && (
            <div className="flex items-center gap-2 px-3 py-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--admin-accent)]/20 text-xs font-medium text-[var(--admin-accent)]">
                {getInitials(displayName)}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-[var(--admin-text-primary)]">{displayName}</p>
                <p className="truncate text-xs text-[var(--admin-text-muted)]">{user.role}</p>
              </div>
            </div>
          )}
          <button
            type="button"
            onClick={logout}
            className="flex w-full items-center gap-2 rounded-lg py-2 text-red-400 hover:bg-red-500/10 hover:text-red-300"
          >
            <LogOut className="h-4 w-4 shrink-0" />
            {!collapsed && <span className="text-sm">Выход</span>}
          </button>
        </div>
      </aside>

      <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-[var(--admin-border)] bg-[var(--admin-bg-surface)] pb-[env(safe-area-inset-bottom)] sm:hidden">
        <nav className="flex gap-2 overflow-x-auto p-3">
          {NAV.map((item) => {
            const Icon = item.icon;
            const isActive =
              location.pathname === item.to ||
              (item.to !== '/admin/stores' && location.pathname.startsWith(item.to));
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex items-center gap-2 whitespace-nowrap rounded-lg px-4 py-2.5 text-xs font-semibold ${
                  isActive
                    ? 'bg-[var(--admin-accent)] text-white'
                    : 'bg-[var(--admin-bg-elevated)] text-[var(--admin-text-muted)]'
                }`}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>
      </div>

      <main className="sm:ml-[var(--sidebar-w)] pb-24 sm:pb-0">
        <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-[var(--admin-border)] bg-[var(--admin-bg-surface)] px-4">
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setCollapsed((c) => !c)}
              className="rounded-lg p-2 text-[var(--admin-text-muted)] hover:bg-[var(--admin-bg-elevated)] sm:hidden"
              aria-label="Меню"
            >
              <LayoutDashboard className="h-5 w-5" />
            </button>
            <span className="text-sm text-[var(--admin-text-muted)] truncate max-w-[60vw] sm:max-w-none">
              {displayName}
            </span>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <button
              type="button"
              onClick={logout}
              className="text-sm text-red-400 hover:text-red-300 sm:hidden"
            >
              Выход
            </button>
            <Link
              to="/"
              className="text-sm text-[var(--admin-text-muted)] hover:text-[var(--admin-text-primary)]"
            >
              На сайт
            </Link>
          </div>
        </header>
        <div className="p-4 sm:p-6">
          <nav
            className="mb-4 flex flex-wrap items-center gap-1.5 text-xs text-[var(--admin-text-muted)]"
            aria-label="Навигация"
          >
            <Link
              to="/admin/stores"
              className="hover:text-[var(--admin-text-primary)] transition-colors"
            >
              Админ
            </Link>
            <span className="text-[var(--admin-border)]" aria-hidden>
              /
            </span>
            <span className="font-medium text-[var(--admin-text-primary)]">{adminSectionTitle(location.pathname)}</span>
          </nav>
          {children}
        </div>
      </main>
    </div>
  );
}
