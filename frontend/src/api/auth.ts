import apiClient from './client';
import { AuthResponse, User } from './types';

export async function register(email: string, displayName: string, password: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/register', { email, displayName, password });
  return data;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/login', { email, password });
  return data;
}

export async function getMe(): Promise<User> {
  const { data } = await apiClient.get<User>('/auth/me');
  return data;
}

export async function searchUsers(query: string, limit = 10): Promise<User[]> {
  const { data } = await apiClient.get<User[]>('/users/search', { params: { q: query, limit } });
  return data;
}
