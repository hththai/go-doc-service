import { createFileRoute } from "@tanstack/react-router";
import { redirect, useRouter } from "@tanstack/react-router";
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
  const navigate = Route.useNavigate();
  const search = Route.useSearch();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [loginError, setLoginError] = useState<string | null>(null);

  async function handleSubmit(username: string, password: string) {
    setLoginError(null);
    try {
      await auth.login(username, password);
      await router.invalidate();
      await queryClient.invalidateQueries({ queryKey: ["me"] });
      await navigate({ to: search.redirect || fallback });
    } catch (err) {
      setLoginError(err instanceof Error ? err.message : "Login failed");
    }
  }

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
            <path
              id="wave-track"
              d="M-50,60 Q150,20 300,60 T600,60 T900,60 T1250,60"
              fill="none"
              stroke="#cbd5e1"
              strokeWidth="3"
            />
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
          <LoginForm onSubmit={handleSubmit} error={loginError} />
        </div>
      </div>
    </>
  );
}
