import { createFileRoute, redirect } from '@tanstack/react-router'
import HomePage from '@/components/Home/Homepage'
import LoginForm from '@/components/Login/Login'

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    throw redirect({
      to: '/upload'
    })
    // component: App,
  },
})

// function App() {
//   return (
//     <div>
//       <LoginForm />
//     </div>
//   )
// }
