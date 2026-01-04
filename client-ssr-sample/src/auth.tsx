import * as React from 'react'
import { sleep } from './utils'
import { useCurrentUser } from './hooks/useCurrentUser'
import { useQueryClient } from '@tanstack/react-query';
import { useLogin } from './hooks/useLogin';

const key = 'tokenId'

function getStoredTokenId() {
    return localStorage.getItem(key)
}

function setStoredTokenId(tokenId: string | null) {
    if (tokenId) {
        localStorage.setItem(key, tokenId)
    }
    else {
        localStorage.removeItem(key)
    }
}

export interface AuthUser {
    tokenId: string;
    username: string;
}

export interface AuthContext {
    user: AuthUser | null
    tokenId: string | null
    isAuthenticated: boolean
    login: (username: string, password: string) => Promise<void>
    logout: () => Promise<void>
    isLoading: boolean;
}

const AuthContext = React.createContext<AuthContext | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [tokenId, setTokenId] = React.useState<string | null>(getStoredTokenId())
    const queryClient = useQueryClient();
    const loginMutation = useLogin();
    const { data: user, isLoading } = useCurrentUser(tokenId);


    const login = async (username: string, password: string) => {
        // Clear stale user before login attempt.
        queryClient.setQueryData(["me"], null);
        const result = await loginMutation.mutateAsync({ username, password });

        const newToken = result?.token?.tokenId;
        setStoredTokenId(newToken)
        setTokenId(newToken)

        await queryClient.invalidateQueries({ queryKey: ["me", tokenId] })
        console.log("\nresult is:::", result)

    }

    const logout = async () => {
        await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/logout`, {
            method: "POST",
            credentials: "include",
        });
        queryClient.setQueryData(["me"], null);
        setStoredTokenId(null)
        await queryClient.invalidateQueries({ queryKey: ["me"] })
    }

    const value: AuthContext = {
        user: user ?? null,
        tokenId: tokenId,
        isAuthenticated: !!tokenId,
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

