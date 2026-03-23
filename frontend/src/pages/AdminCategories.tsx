import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminAPI, type AdminCategoryPatch } from '../services/api';
import { useAuthStore } from '../store/useAuthStore';
import { AdminLayout } from '../components/AdminLayout';

interface Category {
  id: string;
  name: string;
  description: string;
  icon_url: string;
  order: number;
  is_active: boolean;
}

const emptyForm = {
  name: '',
  description: '',
  icon_url: '',
  order: 0,
  is_active: true,
};

export function AdminCategories() {
  const navigate = useNavigate();
  const { accessToken } = useAuthStore();
  const [list, setList] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);

  const load = useCallback(() => {
    if (!accessToken) return;
    setError('');
    return adminAPI
      .listCategories()
      .then((r) => setList((r.data as { data?: Category[] })?.data || []))
      .catch(() => {
        setError('Не удалось загрузить категории');
        navigate('/login?next=' + encodeURIComponent('/admin/categories'));
      })
      .finally(() => setLoading(false));
  }, [accessToken, navigate]);

  useEffect(() => {
    setLoading(true);
    load();
  }, [load]);

  if (!accessToken) return null;

  const startCreate = () => {
    setEditingId('new');
    setForm(emptyForm);
  };

  const startEdit = (c: Category) => {
    setEditingId(c.id);
    setForm({
      name: c.name,
      description: c.description || '',
      icon_url: c.icon_url || '',
      order: c.order,
      is_active: c.is_active,
    });
  };

  const cancelForm = () => {
    setEditingId(null);
    setForm(emptyForm);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      setError('Укажите название');
      return;
    }
    setSaving(true);
    setError('');
    try {
      if (editingId === 'new') {
        await adminAPI.createCategory({
          name: form.name.trim(),
          description: form.description,
          icon_url: form.icon_url,
          order: Number(form.order) || 0,
          is_active: form.is_active,
        });
      } else if (editingId) {
        const patch: AdminCategoryPatch = {
          name: form.name.trim(),
          description: form.description,
          icon_url: form.icon_url,
          order: Number(form.order) || 0,
          is_active: form.is_active,
        };
        await adminAPI.patchCategory(editingId, patch);
      }
      cancelForm();
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Ошибка сохранения');
    } finally {
      setSaving(false);
    }
  };

  const remove = async (c: Category) => {
    if (!window.confirm(`Удалить категорию «${c.name}»? Товары в ней должны быть перенесены.`)) return;
    setError('');
    try {
      await adminAPI.deleteCategory(c.id);
      await load();
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: string } } };
      setError(ax.response?.data?.error || 'Не удалось удалить');
    }
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Категории</h1>
            <p className="text-slate-400 text-sm mt-1">Группы товаров в каталоге</p>
          </div>
          <button
            type="button"
            onClick={startCreate}
            className="rounded-lg bg-[var(--admin-accent)] px-4 py-2.5 text-sm font-medium text-white hover:opacity-90"
          >
            + Новая категория
          </button>
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">{error}</div>
        )}

        {editingId && (
          <form
            onSubmit={submit}
            className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] p-5 space-y-4"
          >
            <h2 className="font-semibold text-[var(--admin-text-primary)]">
              {editingId === 'new' ? 'Новая категория' : 'Редактирование'}
            </h2>
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Название *</span>
                <input
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                  required
                />
              </label>
              <label className="block sm:col-span-2">
                <span className="text-xs text-[var(--admin-text-muted)]">Описание</span>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  rows={2}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">URL иконки</span>
                <input
                  value={form.icon_url}
                  onChange={(e) => setForm((f) => ({ ...f, icon_url: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-sm text-[var(--admin-text-primary)]"
                  placeholder="https://..."
                />
              </label>
              <label className="block">
                <span className="text-xs text-[var(--admin-text-muted)]">Порядок сортировки</span>
                <input
                  type="number"
                  value={form.order}
                  onChange={(e) => setForm((f) => ({ ...f, order: Number(e.target.value) }))}
                  className="mt-1 w-full rounded-lg border border-[var(--admin-border)] bg-[var(--admin-bg-base)] px-3 py-2 text-[var(--admin-text-primary)]"
                />
              </label>
              <label className="flex items-center gap-2 sm:col-span-2">
                <input
                  type="checkbox"
                  checked={form.is_active}
                  onChange={(e) => setForm((f) => ({ ...f, is_active: e.target.checked }))}
                />
                <span className="text-sm text-[var(--admin-text-muted)]">Активна в каталоге</span>
              </label>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={saving}
                className="rounded-lg bg-[var(--admin-accent)] px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
              >
                {saving ? 'Сохранение...' : 'Сохранить'}
              </button>
              <button type="button" onClick={cancelForm} className="rounded-lg px-4 py-2 text-sm text-[var(--admin-text-muted)]">
                Отмена
              </button>
            </div>
          </form>
        )}

        {loading ? (
          <p className="text-[var(--admin-text-muted)]">Загрузка...</p>
        ) : (
          <div className="rounded-xl border border-[var(--admin-border)] bg-[var(--admin-bg-surface)] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[640px]">
                <thead>
                  <tr className="border-b border-[var(--admin-border)]">
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Название</th>
                    <th className="p-4 text-left text-sm font-medium text-[var(--admin-text-muted)]">Порядок</th>
                    <th className="p-4 text-center text-sm font-medium text-[var(--admin-text-muted)]">Активна</th>
                    <th className="p-4 text-right text-sm font-medium text-[var(--admin-text-muted)]"></th>
                  </tr>
                </thead>
                <tbody>
                  {list.map((c) => (
                    <tr key={c.id} className="border-b border-[var(--admin-border)]/50 hover:bg-[var(--admin-bg-elevated)]/40">
                      <td className="p-4">
                        <p className="font-medium text-[var(--admin-text-primary)]">{c.name}</p>
                        {c.description && (
                          <p className="text-xs text-[var(--admin-text-muted)] mt-0.5 line-clamp-2">{c.description}</p>
                        )}
                      </td>
                      <td className="p-4 text-[var(--admin-text-primary)]">{c.order}</td>
                      <td className="p-4 text-center">{c.is_active ? '✓' : '—'}</td>
                      <td className="p-4 text-right space-x-2 whitespace-nowrap">
                        <button
                          type="button"
                          onClick={() => startEdit(c)}
                          className="text-sm text-[var(--admin-accent)] hover:opacity-90"
                        >
                          Изменить
                        </button>
                        <button type="button" onClick={() => remove(c)} className="text-sm text-red-400 hover:text-red-300">
                          Удалить
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {list.length === 0 && <p className="p-8 text-center text-[var(--admin-text-muted)]">Категорий пока нет</p>}
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
