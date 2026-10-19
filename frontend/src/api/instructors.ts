import { apiClient } from "./client";
import type { InstructorProfile, InstructorSummary, NearbyInstructor } from "../types";

export async function createMyProfile(input: {
  vehicle_make: string;
  vehicle_model: string;
  vehicle_plate: string;
  vehicle_year?: number;
  hourly_rate_uyu: number;
  bio?: string;
}) {
  const { data } = await apiClient.post<InstructorProfile>("/api/instructors/me/profile", input);
  return data;
}

export async function getMyProfile() {
  const { data } = await apiClient.get<InstructorProfile>("/api/instructors/me/profile");
  return data;
}

export async function updateMyStatus(status: "online" | "offline" | "busy", lat?: number, lon?: number) {
  const { data } = await apiClient.patch("/api/instructors/me/status", { status, lat, lon });
  return data;
}

export async function getNearbyInstructors(lat: number, lon: number, radiusKm = 5) {
  const { data } = await apiClient.get<NearbyInstructor[]>("/api/instructors/nearby", {
    params: { lat, lon, radius_km: radiusKm },
  });
  return data ?? [];
}

// All instructors regardless of current online status — for booking a future
// scheduled slot, which shouldn't require them to be live right now.
export async function getAllInstructors() {
  const { data } = await apiClient.get<InstructorSummary[]>("/api/instructors");
  return data ?? [];
}
