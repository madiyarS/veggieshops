import axios from 'axios';

export type ParsedApiError = { message: string; code?: string };

/** Разбор ответа API (error + code для поддержки). */
export function parseApiError(e: unknown): ParsedApiError {
  if (axios.isAxiosError(e)) {
    const d = e.response?.data as { error?: string; code?: string } | undefined;
    if (d?.error) {
      return { message: d.error, code: d.code };
    }
    if (e.message) {
      return { message: e.message };
    }
  }
  if (e instanceof Error) {
    return { message: e.message };
  }
  return { message: 'Произошла ошибка. Попробуйте ещё раз.' };
}

/** Текст для пользователя + код в скобках, если есть. */
export function formatApiErrorForUi(e: unknown): string {
  const { message, code } = parseApiError(e);
  if (code) {
    return `${message} (код: ${code})`;
  }
  return message;
}
