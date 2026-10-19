import { apiClient } from "./client";
import type { ExamStatusOverview, ExamSubmission, ExamType, PendingExam } from "../types";

export async function getMyExamStatus() {
  const { data } = await apiClient.get<ExamStatusOverview>("/api/exams/me");
  return data;
}

export async function submitExam(type: ExamType, photo: File) {
  const form = new FormData();
  form.append("type", type);
  form.append("photo", photo);
  const { data } = await apiClient.post<ExamSubmission>("/api/exams/submit", form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return data;
}

export async function getPendingExams() {
  const { data } = await apiClient.get<PendingExam[]>("/api/admin/exams/pending");
  return data ?? [];
}

export async function reviewExam(id: string, approve: boolean, reason?: string) {
  await apiClient.post(`/api/admin/exams/${id}/review`, { approve, reason });
}
