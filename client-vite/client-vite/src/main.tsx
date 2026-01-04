import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'


// Import the generated route tree
import { routeTree } from './routeTree.gen.ts'
import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'

import './styles.css'
import reportWebVitals from './reportWebVitals.ts'
import { AuthProvider, useAuth } from './auth.tsx'
import './styles.css'

// Create a new router instance
const queryClient = new QueryClient()

// const TanStackQueryProviderContext = TanStackQueryProvider.getContext()

// const router = createRouter({
//   routeTree,
//   context: {
//     ...TanStackQueryProviderContext,
//     queryClient,
//     auth: undefined!,
//   },
//   defaultPreload: 'intent',
//   scrollRestoration: true,
//   defaultStructuralSharing: true,
//   defaultPreloadStaleTime: 0,
// })


const router = createRouter({
  routeTree,
  context: {
    queryClient,
    auth: undefined!, // This will be injected at runtime by InnerApp
  },
  defaultPreload: 'intent',
})

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function InnerApp() {
  const auth = useAuth()
  return <RouterProvider router={router} context={{ auth }} />
}


// Render the app
const rootElement = document.getElementById('app')
if (rootElement && !rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      {/* <TanStackQueryProvider.Provider {...TanStackQueryProviderContext}> */}
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          {/* <RouterProvider router={router} /> */}
          <InnerApp />
        </AuthProvider>
      </QueryClientProvider>
      {/* </TanStackQueryProvider.Provider> */}
    </StrictMode>,
  )
}

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals()
