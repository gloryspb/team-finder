import type { Application, ApplicationDetails, AuthResponse, Game, Listing, ListingDetails, PlayerProfile, User } from "../types";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";
const TOKEN_KEY = "team_finder_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

type RequestOptions = RequestInit & {
  auth?: boolean;
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  const token = getToken();
  if (options.auth !== false && token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_URL}${path}`, { ...options, headers });
  const text = await response.text();
  const data = text ? JSON.parse(text) : null;
  if (!response.ok) {
    throw new Error(data?.error ?? "Ошибка запроса");
  }
  return data as T;
}

export const api = {
  register: (payload: { email: string; password: string; nickname: string }) =>
    request<AuthResponse>("/auth/register", { method: "POST", body: JSON.stringify(payload), auth: false }),
  login: (payload: { email: string; password: string }) =>
    request<AuthResponse>("/auth/login", { method: "POST", body: JSON.stringify(payload), auth: false }),
  me: () => request<User>("/me"),
  profile: () => request<PlayerProfile>("/me/profile"),
  updateProfile: (payload: Partial<PlayerProfile>) =>
    request<PlayerProfile>("/me/profile", { method: "PUT", body: JSON.stringify(payload) }),
  games: () => request<Game[]>("/games", { auth: false }),
  listings: (params: Record<string, string> = {}) => {
    const query = new URLSearchParams(Object.entries(params).filter(([, value]) => value !== ""));
    return request<ListingDetails[]>(`/listings${query.toString() ? `?${query}` : ""}`, { auth: false });
  },
  listing: (id: string) => request<ListingDetails>(`/listings/${id}`, { auth: false }),
  createListing: (payload: Partial<Listing>) =>
    request<Listing>("/listings", { method: "POST", body: JSON.stringify(payload) }),
  updateListing: (id: string, payload: Partial<Listing>) =>
    request<Listing>(`/listings/${id}`, { method: "PUT", body: JSON.stringify(payload) }),
  closeListing: (id: string) => request<Listing>(`/listings/${id}/close`, { method: "PATCH" }),
  deleteListing: (id: string) => request<void>(`/listings/${id}`, { method: "DELETE" }),
  apply: (listingId: string, message: string) =>
    request<Application>(`/listings/${listingId}/applications`, { method: "POST", body: JSON.stringify({ message }) }),
  outgoing: () => request<ApplicationDetails[]>("/applications/outgoing"),
  incoming: () => request<ApplicationDetails[]>("/applications/incoming"),
  setApplicationStatus: (id: string, status: "accepted" | "rejected") =>
    request<Application>(`/applications/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) })
};
