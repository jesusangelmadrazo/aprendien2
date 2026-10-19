import { apiClient } from "./client";
import type { LessonRequest } from "../types";

export async function createLesson(input: {
  pickup_lat: number;
  pickup_lon: number;
  pickup_address?: string;
  duration_minutes?: number;
}) {
  const { data } = await apiClient.post<LessonRequest>("/api/lessons", input);
  return data;
}

export async function getLesson(id: string) {
  const { data } = await apiClient.get<LessonRequest>(`/api/lessons/${id}`);
  return data;
}

export async function listMyLessons(status?: string) {
  const { data } = await apiClient.get<LessonRequest[]>("/api/lessons", { params: { status } });
  return data ?? [];
}

export async function respondOfferRest(offerId: string, accept: boolean) {
  const { data } = await apiClient.post(`/api/lessons/offers/${offerId}/respond`, { accept });
  return data;
}

export async function startLesson(id: string) {
  const { data } = await apiClient.post(`/api/lessons/${id}/start`);
  return data;
}

export async function completeLesson(id: string) {
  const { data } = await apiClient.post(`/api/lessons/${id}/complete`);
  return data;
}

export async function cancelLesson(id: string, reason?: string) {
  const { data } = await apiClient.post(`/api/lessons/${id}/cancel`, { reason });
  return data;
}
