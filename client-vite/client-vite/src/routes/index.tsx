import { createFileRoute } from '@tanstack/react-router'
import HomePage from '@/components/Home/Homepage'

export const Route = createFileRoute('/')({
  component: App,
})

function App() {
  return (
    <div>
      <HomePage />
    </div>
  )
}
