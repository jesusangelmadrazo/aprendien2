import { apiClient } from "./client";
import type { AccountType, InstructorBalance, Payout, PayoutMethod, WalletSummary } from "../types";

export async function getMyWallet() {
  const { data } = await apiClient.get<WalletSummary>("/api/instructors/me/wallet");
  return data;
}

export async function getMyPayoutMethod() {
  const { data } = await apiClient.get<PayoutMethod>("/api/instructors/me/payout-method");
  return data;
}

export async function setMyPayoutMethod(input: {
  bank_name: string;
  account_type: AccountType;
  account_number: string;
  account_holder_name: string;
  account_holder_document: string;
}) {
  const { data } = await apiClient.put<PayoutMethod>("/api/instructors/me/payout-method", input);
  return data;
}

export async function getPendingBalances() {
  const { data } = await apiClient.get<InstructorBalance[]>("/api/admin/payouts/pending");
  return data ?? [];
}

export async function createPayout(instructorId: string) {
  const { data } = await apiClient.post<Payout>(`/api/admin/payouts/${instructorId}`);
  return data;
}
