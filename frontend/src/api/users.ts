import { apiClient } from "./client";

export async function getUserName(userId: string) {
  const { data } = await apiClient.get<{ id: string; full_name: string }>(`/api/users/${userId}`);
  return data.full_name;
}
