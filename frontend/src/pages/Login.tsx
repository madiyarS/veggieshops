import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { User, Lock, Eye, EyeOff } from 'lucide-react';
import { authAPI } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { redirectAfterLogin } from '../utils/authRedirect';

type AuthPayload = {
  user?: { id: string; phone: string; first_name: string; last_name: string; role: string };
  access_token?: string;
  refresh_token?: string;
};

export function Login() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const next = searchParams.get('next');
  const setAuth = useAuthStore((s) => s.setAuth);
  const existingUser = useAuthStore((s) => s.user);
  const existingToken = useAuthStore((s) => s.accessToken);

  const [phone, setPhone] = useState('+7');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (existingToken && existingUser) {
      redirectAfterLogin(existingUser.role, next, navigate);
    }
  }, [existingToken, existingUser, next, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await authAPI.login({ phone, password });
      const data = (res.data as { data?: AuthPayload })?.data;
      if (!data?.user || !data?.access_token) {
        setError('Ошибка ответа сервера');
        return;
      }
      if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token);
      }
      setAuth(data.user, data.access_token);
      redirectAfterLogin(data.user.role, next, navigate);
    } catch {
      setError('Неверный телефон или пароль');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 bg-gray-50">
      <div className="w-full max-w-[400px] rounded-xl border border-gray-200 bg-white p-8 shadow-md">
        <h1 className="text-xl font-semibold text-veggie-green text-center">🥬 VeggieShops.kz</h1>
        <p className="mt-1 text-center text-sm text-gray-500">Вход в аккаунт</p>
        <form onSubmit={handleSubmit} className="mt-8 space-y-5">
          <div>
            <label className="sr-only" htmlFor="login-phone">
              Телефон
            </label>
            <div className="relative">
              <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                id="login-phone"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="w-full rounded-lg border border-gray-200 bg-white py-3 pl-10 pr-4 outline-none focus:border-veggie-green focus:ring-1 focus:ring-veggie-green"
                placeholder="+77001234567"
                required
              />
            </div>
          </div>
          <div>
            <label className="sr-only" htmlFor="login-password">
              Пароль
            </label>
            <div className="relative">
              <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                id="login-password"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-lg border border-gray-200 bg-white py-3 pl-10 pr-12 outline-none focus:border-veggie-green focus:ring-1 focus:ring-veggie-green"
                placeholder="Пароль"
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-700"
                aria-label={showPassword ? 'Скрыть' : 'Показать пароль'}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>
          {error && (
            <div className="rounded-lg bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-600">{error}</div>
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-veggie-green py-3 font-medium text-white hover:bg-veggie-dark disabled:opacity-50"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>
        <p className="mt-6 text-center text-sm text-gray-500">
          Нет аккаунта?{' '}
          <Link
            to={next ? `/register?next=${encodeURIComponent(next)}` : '/register'}
            className="text-veggie-green font-medium hover:underline"
          >
            Регистрация
          </Link>
        </p>
      </div>
      <Link to="/" className="mt-8 text-sm text-gray-500 hover:text-gray-800">
        ← На главную
      </Link>
    </div>
  );
}
