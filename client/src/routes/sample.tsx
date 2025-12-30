import { createFileRoute } from '@tanstack/react-router'
import { redirect, useRouter, useRouterState } from '@tanstack/react-router'
import * as React from 'react'
import { useAuth } from '../auth'
import { sleep } from '../utils'
import { z } from 'zod'
import { useState } from "react";
import { useLogin } from "../hooks/useLogin";
import { useRefresh } from "../hooks/useRefresh";



export const Route = createFileRoute('/sample')({
    component: LoginComponent,
})

function LoginComponent() {
    const login = useLogin();
    const refresh = useRefresh();

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        login.mutate({ username, password });
    }
    function handleRefresh() {
        refresh.mutate();
    }

    return (
        <div style={{ width: 300, margin: "40px auto", padding: 20, border: "1px solid #ccc" }}>
            <h2>Login</h2>

            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    style={{ width: "100%", marginBottom: 10 }}
                />

                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    style={{ width: "100%", marginBottom: 10 }}
                />

                <button type="submit" disabled={login.isPending} style={{ width: "100%" }}>
                    {login.isPending ? "Logging in..." : "Login"}
                </button>
            </form>

            <div style={{ marginTop: 20 }}>
                <button onClick={handleRefresh} disabled={refresh.isPending}>
                    {refresh.isPending ? "Refreshing..." : "Refresh"}
                </button>

                {refresh.isError && (
                    <p style={{ color: "red" }}>{(refresh.error as Error).message}</p>
                )}

                {refresh.isSuccess && (
                    <p style={{ color: "green" }}>Access token refreshed</p>
                )}
            </div>
            {login.isError && (
                <p style={{ color: "red", marginTop: 10 }}>{(login.error as Error).message}</p>
            )}

            {login.isSuccess && (
                <p style={{ color: "green", marginTop: 10 }}>Login successful</p>
            )}
        </div>
    );
}