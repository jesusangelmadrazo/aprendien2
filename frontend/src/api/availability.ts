import { apiClient } from "./client";
import type { AvailabilityBlock, AvailabilitySlot, BookingResult, ScheduleOverride, SlotWithInstructor } from "../types";

export async function createBlock(input: { day_of_week: number; start_time: string; end_time: string }) {
  const { data } = await apiClient.post<AvailabilityBlock>("/api/instructors/me/availability", input);
  return data;
}

export async function listMyBlocks() {
  const { data } = await apiClient.get<AvailabilityBlock[]>("/api/instructors/me/availability");
  return data ?? [];
}

export async function deleteBlock(id: string) {
  await apiClient.delete(`/api/instructors/me/availability/${id}`);
}

export async function getInstructorSlots(instructorId: string, days = 14) {
  const { data } = await apiClient.get<AvailabilitySlot[]>(`/api/instructors/${instructorId}/slots`, {
    params: { days },
  });
  return data ?? [];
}

// Cross-instructor search: every slot open on one date, optionally narrowed
// to a single instructor.
export async function getSlotsForDate(date: string, instructorId?: string) {
  const { data } = await apiClient.get<SlotWithInstructor[]>("/api/instructors/slots", {
    params: { date, instructor_id: instructorId || undefined },
  });
  return data ?? [];
}

export async function setOverride(input: { date: string; is_day_off: boolean; start_time?: string; end_time?: string }) {
  const { data } = await apiClient.post<ScheduleOverride>("/api/instructors/me/schedule-overrides", input);
  return data;
}

export async function listMyOverrides() {
  const { data } = await apiClient.get<ScheduleOverride[]>("/api/instructors/me/schedule-overrides");
  return data ?? [];
}

export async function deleteOverride(date: string) {
  await apiClient.delete(`/api/instructors/me/schedule-overrides/${date}`);
}

export async function bookSlot(
  instructorId: string,
  input: { scheduled_at: string; pickup_lat: number; pickup_lon: number; pickup_address?: string; duration_minutes?: number }
) {
  const { data } = await apiClient.post<BookingResult>(`/api/instructors/${instructorId}/book`, input);
  return data;
}
