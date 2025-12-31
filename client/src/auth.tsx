import * as React from 'react'

import { sleep } from './utils'
import { useCurrentUser } from './hooks/useCurrentUser'

// export interface AuthContext {
//     isAuthenticated: boolean
//     login: (username: string) => Promise<void>
//     logout: () => Promise<void>
//     user: string | null
// }

// const AuthContext = React.createContext<AuthContext | null>(null)

// const key = 'usr'

// function getStoredUser() {
//     return localStorage.getItem(key)
// }

// function setStoredUser(user: string | null) {
//     if (user) {
//         localStorage.setItem(key, user)
//     } else {
//         localStorage.removeItem(key)
//     }
// }

// export function AuthProvider({ children }: { children: React.ReactNode }) {
//     const [user, setUser] = React.useState<string | null>(getStoredUser())
//     const isAuthenticated = !!user

//     const logout = React.useCallback(async () => {
//         await sleep(250)

//         setStoredUser(null)
//         setUser(null)
//     }, [])

//     const login = React.useCallback(async (username: string) => {
//         await sleep(500)

//         setStoredUser(username)
//         setUser(username)
//     }, [])

//     React.useEffect(() => {
//         setUser(getStoredUser())
//     }, [])



//     return (
//         <AuthContext.Provider value={{ isAuthenticated, user, login, logout }}>
//             {children}
//         </AuthContext.Provider>
//     )
// }

export interface AuthContext {
    user: { username: string } | null;
    isAuthenticated: boolean;
    isLoading: boolean;
}

const AuthContext = React.createContext<AuthContext | null>(null);


export function AuthProvider({ children }: { children: React.ReactNode }) {
    const { data: user, isLoading, isError } = useCurrentUser();

    const value = {
        user,
        isAuthenticated: !!user && !isError,
        isLoading,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>


}



export function useAuth() {
    const context = React.useContext(AuthContext)
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider')
    }
    return context
}

// TODO: store in api folder
// Sample login function.
export async function loginRequest({ username, password }: { username: string, password: string }) {
    const res = await fetch("http://localhost:8088/loginjwt", {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();

    if (!res.ok) {
        throw new Error(data.error || "Login failed");
    }

    return data;
}

export async function refreshRequest() {
    const res = await fetch("http://localhost:8088/v1/auth/refresh", {
        method: "POST",
        credentials: "include", // REQUIRED: sends refresh_token cookie
    });

    const data = await res.json();

    if (!res.ok) {
        throw new Error(data.error || "Refresh failed");
    }

    return data;
}
// api folder
export async function getMe() {
    const res = await fetch("http://localhost:8088/v1/auth/me", {
        credentials: "include", // sends cookies
    });

    if (!res.ok) {
        throw new Error("Not authenticated");
    }

    return res.json();
}