import * as React from 'react'
import { sleep } from './utils'
import { useCurrentUser } from './hooks/useCurrentUser'
import { useQueryClient } from '@tanstack/react-query';

const key = 'token'


export interface AuthContext {
    user: string | null
    isAuthenticated: boolean
    login: (username: string, password: string) => Promise<void>
    logout: () => Promise<void>
    isLoading: boolean;
}

const AuthContext = React.createContext<AuthContext | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const { data: user, isLoading } = useCurrentUser();
    const queryClient = useQueryClient();
    const login = async (username: string, password: string) => {
        await fetch("http://localhost:8088/loginjwt", {
            method: "POST",
            credentials: "include",
            body: JSON.stringify({ username, password })
        });

        queryClient.invalidateQueries({ queryKey: ["me"] })
    }

    const logout = async () => {
        await fetch("http://localhost:8088/v1/auth/logout", {
            method: "POST",
            credentials: "include",
        });

        queryClient.setQueryData(["me"], null);
        queryClient.invalidateQueries({ queryKey: ["me"] })
    }

    const value: AuthContext = {
        user: user ?? null,
        isAuthenticated: !!user,
        login,
        logout,
        isLoading,
    };

    console.log(value)
    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
    const context = React.useContext(AuthContext)
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider')
    }
    return context
}

