import { createFileRoute } from '@tanstack/react-router'
import { Outlet, redirect, useRouter } from '@tanstack/react-router'
import { useAuth } from '../auth'

export const Route = createFileRoute('/_auth')({

    beforeLoad: ({ context, location }) => {

        if (!context.auth.isAuthenticated) {
            throw redirect({
                to: '/signin',
                search: {
                    redirect: location.href,
                },
            })
        }
    },

    component: AuthLayout
})

function AuthLayout() {
    const router = useRouter()
    const navigate = Route.useNavigate()
    const auth = useAuth()

    const handleLogout = () => {
        if (window.confirm('Are you sure you want to logout?')) {
            auth.logout().then(() => {
                router.invalidate().finally(() => {
                    navigate({ to: '/', replace: true })
                })
            })
        }
    }

    // React.useEffect(() => {
    //     if (!auth.isLoading && !auth.isAuthenticated) {
    //         navigate({ to: "/signin" });
    //     }
    // }, [auth.isLoading, auth.isAuthenticated]);

    return (
        <div className="p-2 h-full">
            <h1>Authenticated Route</h1>
            <p>This route's content is only visible to authenticated users.</p>
            <ul className="py-2 flex gap-2">
                <li>
                    <button
                        type="button"
                        className="hover:underline"
                        onClick={handleLogout}
                    >
                        Logout
                    </button>
                </li>
            </ul>
            <hr />
            <Outlet />
        </div>
    )
}
