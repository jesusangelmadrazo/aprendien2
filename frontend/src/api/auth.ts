import { apiClient } from "./client";
import type { Role, User } from "../types";

interface AuthResponse {
  access_token: string;
  user: User;
}

export async function register(input: {
  email: string;
  phone: string;
  password: string;
  full_name: string;
  role: Role;
}) {
  const { data } = await apiClient.post<AuthResponse>("/api/auth/register", input);
  return data;
}

export async function login(email: string, password: string) {
  const { data } = await apiClient.post<AuthResponse>("/api/auth/login", { email, password });
  return data;
}
