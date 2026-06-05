import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api, clearToken, getToken, setToken } from "./client";
import type { PlayerProfile, User } from "../types";

type AuthContextValue = {
  user: User | null;
  profile: PlayerProfile | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, nickname: string) => Promise<void>;
  logout: () => void;
  refreshProfile: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<PlayerProfile | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      if (!getToken()) {
        setLoading(false);
        return;
      }
      try {
        const [me, playerProfile] = await Promise.all([api.me(), api.profile()]);
        setUser(me);
        setProfile(playerProfile);
      } catch {
        clearToken();
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const value = useMemo<AuthContextValue>(() => ({
    user,
    profile,
    loading,
    isAuthenticated: Boolean(user),
    login: async (email, password) => {
      const response = await api.login({ email, password });
      setToken(response.token);
      setUser(response.user);
      setProfile(response.profile);
    },
    register: async (email, password, nickname) => {
      const response = await api.register({ email, password, nickname });
      setToken(response.token);
      setUser(response.user);
      setProfile(response.profile);
    },
    logout: () => {
      clearToken();
      setUser(null);
      setProfile(null);
    },
    refreshProfile: async () => {
      const playerProfile = await api.profile();
      setProfile(playerProfile);
    }
  }), [user, profile, loading]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return value;
}
