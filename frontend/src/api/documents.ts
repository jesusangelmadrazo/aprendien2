import { apiClient } from "./client";
import type { DocumentStatusOverview, DocumentType, InstructorDocument, PendingDocument } from "../types";

export async function getMyDocumentStatus() {
  const { data } = await apiClient.get<DocumentStatusOverview>("/api/instructors/me/documents");
  return data;
}

export async function submitDocument(docType: DocumentType, photo: File, expiryDate?: string) {
  const form = new FormData();
  form.append("doc_type", docType);
  if (expiryDate) form.append("expiry_date", expiryDate);
  form.append("photo", photo);
  const { data } = await apiClient.post<InstructorDocument>("/api/instructors/me/documents", form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return data;
}

export async function getPendingDocuments() {
  const { data } = await apiClient.get<PendingDocument[]>("/api/admin/documents/pending");
  return data ?? [];
}

export async function reviewDocument(id: string, approve: boolean, reason?: string) {
  await apiClient.post(`/api/admin/documents/${id}/review`, { approve, reason });
}
