import { api, ApiError } from "./api";
import type { AuthResponse, LoginPayload, RegisterPayload, User } from "../types/user";

const MOCK = import.meta.env.VITE_MOCK === "true";

// Demo account for testing without a backend
export const DEMO_EMAIL = "demo@cloudpulse.dev";
export const DEMO_PASSWORD = "demo1234";

const delay = <T,>(v: T) => new Promise<T>((r) => setTimeout(() => r(v), 400));
const fail = (status: number, msg: string) =>
  new Promise<never>((_, rej) => setTimeout(() => rej(new ApiError(status, msg)), 400));

// Accounts created through the Register page (kept in memory, lost on refresh)
const mockUsers: { user: User; password: string }[] = [
  {
    user: { id: "demo", name: "Demo User", email: DEMO_EMAIL },
    password: DEMO_PASSWORD,
  },
];

const mockToken = (u: User) => `mock.${u.id}.${Date.now()}`;

export const authService = {
  login: (payload: LoginPayload): Promise<AuthResponse> => {
    if (MOCK) {
      const found = mockUsers.find(
        (u) =>
          u.user.email.toLowerCase() === payload.email.trim().toLowerCase() &&
          u.password === payload.password
      );
      if (!found) return fail(401, "Invalid email or password");
      return delay({ token: mockToken(found.user), user: found.user });
    }
    return api<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  register: (payload: RegisterPayload): Promise<AuthResponse> => {
    if (MOCK) {
      const email = payload.email.trim().toLowerCase();
      if (mockUsers.some((u) => u.user.email.toLowerCase() === email)) {
        return fail(409, "Email already registered");
      }
      const user: User = {
        id: String(Date.now()),
        name: payload.name,
        email,
      };
      mockUsers.push({ user, password: payload.password });
      return delay({ token: mockToken(user), user });
    }
    return api<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
};