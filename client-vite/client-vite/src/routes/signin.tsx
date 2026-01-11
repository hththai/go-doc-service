import { createFileRoute } from '@tanstack/react-router'
import { redirect, useRouter } from '@tanstack/react-router'
import * as React from 'react'
import { useAuth } from '../auth'
import { z } from 'zod'
import { useState } from "react";
//import { useRefresh } from "../hooks/useRefresh";
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useQueryClient } from '@tanstack/react-query';
import LoginForm from '@/components/Login/Login'

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
    // const refresh = useRefresh();
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

            await navigate({ to: search.redirect || fallback })

        } catch (err) {
            console.error(err);
        }
        finally {
            // setIsSubmitting(false)
        }
    }


    // async function handleRefresh() {
    //     // refresh.mutate();
    //     // const result = await queryClient.fetchQuery({ queryKey: ["login", username] })
    //     const result2 = await queryClient.fetchQuery({ queryKey: ["me"] })
    //     console.log(`result of testing::: ${result2}`)

    // }

    // function handlePing() {
    //     // refetch();
    // }

    return (
        <div>
            <LoginForm
                username={username}
                password={password}
                setUsername={setUsername}
                setPassword={setPassword}
                handleSubmit={handleSubmit} />
        </div>

    );
}