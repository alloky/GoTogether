import apiClient from './client';

export async function linkTelegram(telegramUsername: string): Promise<void> {
  await apiClient.post('/auth/link/telegram', { telegramUsername });
}
