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

export function Register() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const next = searchParams.get('next');
  const setAuth = useAuthStore((s) => s.setAuth);
  const existingUser = useAuthStore((s) => s.user);
  const existingToken = useAuthStore((s) => s.accessToken);

  const [phone, setPhone] = useState('+7');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
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
      const res = await authAPI.register({
        phone,
        password,
        first_name: firstName || undefined,
        last_name: lastName || undefined,
      });
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
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Не удалось зарегистрироваться');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 bg-gray-50">
      <div className="w-full max-w-[400px] rounded-xl border border-gray-200 bg-white p-8 shadow-md">
        <p className="text-center text-sm font-medium text-veggie-green">VeggieShops.kz</p>
        <h1 className="text-xl font-semibold text-gray-900 text-center mt-1">Регистрация</h1>
        <p className="mt-1 text-center text-sm text-gray-500">Покупатель — заказы на сайте</p>
        <form onSubmit={handleSubmit} className="mt-8 space-y-4">
          <input
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            placeholder="Имя"
            className="w-full rounded-lg border border-gray-200 px-4 py-3 outline-none focus:border-veggie-green"
          />
          <input
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            placeholder="Фамилия"
            className="w-full rounded-lg border border-gray-200 px-4 py-3 outline-none focus:border-veggie-green"
          />
          <div className="relative">
            <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full rounded-lg border border-gray-200 py-3 pl-10 pr-4 outline-none focus:border-veggie-green"
              placeholder="+77001234567"
              required
            />
          </div>
          <div className="relative">
            <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-lg border border-gray-200 py-3 pl-10 pr-12 outline-none focus:border-veggie-green"
              placeholder="Пароль"
              required
              minLength={6}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400"
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
          {error && <div className="rounded-lg bg-red-50 border border-red-100 px-4 py-2 text-sm text-red-600">{error}</div>}
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-veggie-green py-3 font-medium text-white disabled:opacity-50"
          >
            {loading ? 'Регистрация...' : 'Создать аккаунт'}
          </button>
        </form>
        <p className="mt-6 text-center text-sm text-gray-500">
          Уже есть аккаунт?{' '}
          <Link
            to={next ? `/login?next=${encodeURIComponent(next)}` : '/login'}
            className="text-veggie-green font-medium hover:underline"
          >
            Войти
          </Link>
        </p>
      </div>
      <Link to="/" className="mt-8 text-sm text-gray-500 hover:text-gray-800">
        ← На главную
      </Link>
    </div>
  );
}
