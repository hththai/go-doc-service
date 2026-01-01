import { useQueryClient } from '@tanstack/react-query'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useAuth } from '@/auth'
import { useRouter } from '@tanstack/react-router'

export function useSession() {
    const queryClient = useQueryClient()
    const router = useRouter()
    const auth = useAuth()
    const { data: user, isLoading, refetch } = useCurrentUser()

    async function login(username: string, password: string, redirectTo?: string) {
        // Prevent stale user from triggering redirects
        queryClient.setQueryData(["me"], null)

        await auth.login(username, password)
        await refetch()
        await router.invalidate()

        const finalUser = queryClient.getQueryData(["me"])

        if (finalUser) {
            await router.navigate({
                to: redirectTo || '/welcome/$username',
                params: { username: finalUser.username },
            })
        }

        return finalUser
    }

    async function logout(redirectTo = '/signin') {
        await auth.logout()

        // Clear cached user immediately
        queryClient.setQueryData(["me"], null)
        await router.invalidate()

        await router.navigate({ to: redirectTo })
    }

    async function refresh() {
        await auth.refresh()
        await refetch()
    }

    return {
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        refresh,
        refetchUser: refetch,
    }
}