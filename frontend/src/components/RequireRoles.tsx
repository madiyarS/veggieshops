import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/useAuthStore';

export function RequireRoles({ roles, children }: { roles: string[]; children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);
  const navigate = useNavigate();
  const loc = useLocation();

  useEffect(() => {
    if (!accessToken) {
      navigate(`/login?next=${encodeURIComponent(loc.pathname + loc.search)}`, { replace: true });
      return;
    }
    if (user && !roles.includes(user.role)) {
      navigate('/', { replace: true });
    }
  }, [accessToken, user, roles, navigate, loc.pathname, loc.search]);

  if (!accessToken || !user || !roles.includes(user.role)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 text-gray-600">
        Проверка доступа...
      </div>
    );
  }

  return <>{children}</>;
}
