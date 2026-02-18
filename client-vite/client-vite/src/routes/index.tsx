import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    throw redirect({
      to: "/upload",
    });
    // component: App,
  },
});

// function App() {
//   return (
//     <div>
//       <LoginForm />
//     </div>
//   )
// }
