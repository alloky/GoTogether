import apiClient from './client';
import { Meeting, TimeSlot, CreateMeetingInput, VoteInput } from './types';

export async function listMeetings(): Promise<Meeting[]> {
  const { data } = await apiClient.get<Meeting[]>('/meetings/');
  return data;
}

export async function listAllMeetings(): Promise<Meeting[]> {
  const { data } = await apiClient.get<Meeting[]>('/meetings/all');
  return data;
}

export async function getMeeting(id: string): Promise<Meeting> {
  const { data } = await apiClient.get<Meeting>(`/meetings/${id}/`);
  return data;
}

export async function createMeeting(input: CreateMeetingInput): Promise<Meeting> {
  const { data } = await apiClient.post<Meeting>('/meetings/', input);
  return data;
}

export async function updateMeeting(id: string, input: { title?: string; description?: string }): Promise<Meeting> {
  const { data } = await apiClient.put<Meeting>(`/meetings/${id}/`, input);
  return data;
}

export async function deleteMeeting(id: string): Promise<void> {
  await apiClient.delete(`/meetings/${id}/`);
}

export async function confirmMeeting(id: string, timeSlotId?: string): Promise<Meeting> {
  const { data } = await apiClient.post<Meeting>(`/meetings/${id}/confirm`, timeSlotId ? { timeSlotId } : {});
  return data;
}

export async function addParticipants(id: string, emails: string[]): Promise<void> {
  await apiClient.post(`/meetings/${id}/participants`, { emails });
}

export async function updateRSVP(id: string, status: 'accepted' | 'declined'): Promise<void> {
  await apiClient.put(`/meetings/${id}/participants/rsvp`, { status });
}

export async function vote(id: string, input: VoteInput): Promise<void> {
  await apiClient.post(`/meetings/${id}/votes`, input);
}

export async function getVotes(id: string): Promise<TimeSlot[]> {
  const { data } = await apiClient.get<TimeSlot[]>(`/meetings/${id}/votes`);
  return data;
}

export async function setTags(id: string, tags: string[]): Promise<void> {
  await apiClient.put(`/meetings/${id}/tags`, { tags });
}

export async function getAllTags(): Promise<string[]> {
  const { data } = await apiClient.get<string[]>('/meetings/tags/all');
  return data;
}
