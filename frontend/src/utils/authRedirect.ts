import type { NavigateFunction } from 'react-router-dom';

function safePath(p: string | null): string {
  if (!p || !p.startsWith('/') || p.startsWith('//')) return '/';
  return p;
}

/** Куда отправить пользователя после успешного входа по роли. */
export function redirectAfterLogin(role: string, next: string | null, navigate: NavigateFunction) {
  if (role === 'admin' || role === 'manager') {
    const target = next?.startsWith('/admin') ? safePath(next) : '/admin/stores';
    navigate(target, { replace: true });
    return;
  }
  if (role === 'courier') {
    const target = next?.startsWith('/courier') ? safePath(next) : '/courier';
    navigate(target, { replace: true });
    return;
  }
  const target =
    next && !next.startsWith('/admin') && !next.startsWith('/courier') ? safePath(next) : '/';
  navigate(target, { replace: true });
}
