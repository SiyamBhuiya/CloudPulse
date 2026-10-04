import { api } from "./api";
import type { AuthResponse, LoginPayload, RegisterPayload } from "../types/user";

export const authService = {
  login: (payload: LoginPayload) =>
    api<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  register: (payload: RegisterPayload) =>
    api<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};