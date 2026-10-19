import { apiClient } from "./client";
import type { Payment } from "../types";

export async function getMyPayments() {
  const { data } = await apiClient.get<Payment[]>("/api/payments/me");
  return data ?? [];
}
