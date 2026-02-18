import * as React from "react";
import { useCurrentUser } from "./hooks/useCurrentUser";
import { useQueryClient } from "@tanstack/react-query";
import { useLogin } from "./hooks/useLogin";

export interface AuthUser {
  userId: number;
}

export interface AuthContext {
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isLoading: boolean;
}

const AuthContext = React.createContext<AuthContext | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();
  const loginMutation = useLogin();
  const { data: user, isLoading } = useCurrentUser();

  const login = async (username: string, password: string) => {
    await loginMutation.mutateAsync({ username, password });
    // Refresh the session check — cookie is now set by the server
    await queryClient.invalidateQueries({ queryKey: ["me"] });
  };

  const logout = async () => {
    await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
    queryClient.setQueryData(["me"], null);
  };

  const value: AuthContext = {
    user: user ?? null,
    isAuthenticated: !!user,
    login,
    logout,
    isLoading,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
