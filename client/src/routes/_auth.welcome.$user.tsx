import { createFileRoute } from '@tanstack/react-router'
import { useAuth } from '../auth'


export const Route = createFileRoute('/_auth/welcome/$user')({

    component: UserHomePage,
})

function UserHomePage() {
    const auth = useAuth()
    const { user } = Route.useParams()

    return (
        <section className="grid gap-2 p-2">
            <p>Hi {user}!</p>
            <p>You are currently on the dashboard route.</p>
        </section>
    )
}