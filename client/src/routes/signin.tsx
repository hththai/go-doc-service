import { createFileRoute } from '@tanstack/react-router'
import { redirect, useRouter, useRouterState, useSearch } from '@tanstack/react-router'
import * as React from 'react'
import { useAuth } from '../auth'
import { sleep } from '../utils'
import { z } from 'zod'
import { useState } from "react";
import { useLogin } from "../hooks/useLogin";
import { useRefresh } from "../hooks/useRefresh";
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useQueryClient } from '@tanstack/react-query';

const fallback = '/welcome' as const



export const Route = createFileRoute('/signin')({
    validateSearch: z.object({
        redirect: z.string().optional().catch(''),
    }),
    beforeLoad: ({ context, search }) => {
        if (context.auth.isAuthenticated) {
            throw redirect({ to: search.redirect || fallback })
        }
    },
    component: SigninComponent,
})

function SigninComponent() {
    const auth = useAuth();
    const login = useLogin();
    const refresh = useRefresh();
    const { refetch } = useCurrentUser();
    const router = useRouter();
    const { redirect: redirectTarget } = useSearch({ from: '/signin' })
    // const queryClient = useQueryClient();


    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setIsSubmitting(true)

        try {
            // queryClient.setQueryData(["me"], null);
            // 1. Perform Login.
            await auth.login(username, password);

            // 2. Fetch the user profile and wait for it.
            await refetch();

            // 3. Invalidate the router to re-run 'beforeLoad' and refresh context.
            // reforec tanstack to validate auth again.
            // isAuthenticated is validate.
            await router.invalidate();

            // Redirect after successful login.
            // 4. Navigate to the intended destination or the welcome page.
            if (redirectTarget) {
                await router.navigate({ to: redirectTarget })
            } else {
                await router.navigate({
                    to: '/welcome/$username',
                    params: { username },
                })
            }

        } catch (err) {
            console.error(err);
        }
        finally {
            setIsSubmitting(false)
        }
    }


    async function handleRefresh() {
        refresh.mutate();
    }

    function handlePing() {
        refetch();
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

            <div>
                <button onClick={handlePing}>Test</button>
            </div>
        </div>
    );
}