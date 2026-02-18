import { createFileRoute } from "@tanstack/react-router";
import { redirect, useRouter } from "@tanstack/react-router";
import * as React from "react";
import { useAuth } from "../auth";
import { z } from "zod";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import LoginForm from "@/components/Login/Login";

const fallback = "/welcome" as const;

export const Route = createFileRoute("/signin")({
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: ({ context, search }) => {
    if (context.auth.isAuthenticated) {
      throw redirect({ to: search.redirect || fallback });
    }
  },
  component: SigninComponent,
});

function SigninComponent() {
  const auth = useAuth();

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
    console.log("current auth::: ", auth);
    try {
      e.preventDefault();
      // queryClient.setQueryData(["me"], null);
      // 1. Perform Login.
      await auth.login(username, password);

      await router.invalidate();

      await queryClient.invalidateQueries({ queryKey: ["me"] });

      await navigate({ to: search.redirect || fallback });
    } catch (err) {
      console.error(err);
    } finally {
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
    <>
      <div className="relative w-full h-full flex items-center justify-center">
        {/* Animated wave */}
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <svg
            className="w-full h-32"
            viewBox="0 0 1200 120"
            preserveAspectRatio="none"
          >
            {/* Simple wave track */}
            <path
              id="wave-track"
              d="M-50,60 Q150,20 300,60 T600,60 T900,60 T1250,60"
              fill="none"
              stroke="#cbd5e1"
              strokeWidth="3"
            />
            {/* Small roller coaster cart */}
            <g className="animate-cart">
              <rect
                x="-12"
                y="-10"
                width="24"
                height="12"
                rx="3"
                fill="#1e293b"
              />
              <circle cx="-6" cy="4" r="3" fill="#475569" />
              <circle cx="6" cy="4" r="3" fill="#475569" />
            </g>
          </svg>
        </div>
        <div className="backdrop-blur-md bg-white/30 p-6 rounded-xl w-full max-w-md sm:max-w-lg md:max-w-xl">
          <LoginForm
            username={username}
            password={password}
            setUsername={setUsername}
            setPassword={setPassword}
            handleSubmit={handleSubmit}
          />
        </div>
      </div>
    </>
  );
}
