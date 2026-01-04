import { createFileRoute } from '@tanstack/react-router'
import { redirect, useRouter } from '@tanstack/react-router'
import * as React from 'react'
import { useAuth } from '../auth'
import { z } from 'zod'
import { useState } from "react";
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

    const tokenId = auth.tokenId
    const { data: user } = useCurrentUser(tokenId);

    console.log("\nValue return current user::: ", user)
    // const login = useLogin();
    const refresh = useRefresh();
    const navigate = Route.useNavigate();
    const search = Route.useSearch();

    const router = useRouter();
    // const { redirect: redirectTarget } = useSearch({ from: '/signin' })
    const queryClient = useQueryClient();


    const [username, setUsername] = useState("");
    // const { refetch } = useCurrentUser(username);
    const [password, setPassword] = useState("");
    // const [isSubmitting, setIsSubmitting] = useState(false)

    async function handleSubmit(e: React.FormEvent) {
        // setIsSubmitting(true)
        console.log("current auth::: ", auth)
        try {
            e.preventDefault();
            // queryClient.setQueryData(["me"], null);
            // 1. Perform Login.
            await auth.login(username, password);

            await router.invalidate();

            await queryClient.invalidateQueries({ queryKey: ["me", auth.tokenId] })

            // await sleep(1)

            // if (!isLoading) {
            //     await navigate({ to: search.redirect || fallback })
            // }
            await navigate({ to: search.redirect || fallback })

        } catch (err) {
            console.error(err);
        }
        finally {
            // setIsSubmitting(false)
        }
    }


    async function handleRefresh() {
        // refresh.mutate();
        // const result = await queryClient.fetchQuery({ queryKey: ["login", username] })
        const result2 = await queryClient.fetchQuery({ queryKey: ["me"] })
        console.log(`result of testing::: ${result2}`)

    }

    function handlePing() {
        // refetch();
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

                <button type="submit" disabled={false} style={{ width: "100%" }}>
                    Sign In
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
            {/* {login.isError && (
                <p style={{ color: "red", marginTop: 10 }}>{(login.error as Error).message}</p>
            )}

            {login.isSuccess && (
                <p style={{ color: "green", marginTop: 10 }}>Login successful</p>
            )} */}

            <div>
                <button onClick={handlePing}>Test</button>
            </div>
        </div>
    );
}