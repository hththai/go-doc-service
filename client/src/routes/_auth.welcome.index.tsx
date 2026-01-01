import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/welcome/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello!</div>
}
