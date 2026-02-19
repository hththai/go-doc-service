import { createFileRoute } from "@tanstack/react-router";
import { Outlet, redirect, useRouter } from "@tanstack/react-router";
import { useAuth } from "../auth";

export const Route = createFileRoute("/_auth")({
  beforeLoad: ({ context, location }) => {
    if (context.auth.isLoading) return; // wait for auth check to complete
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: "/signin",
        search: {
          redirect: location.href,
        },
      });
    }
  },

  component: AuthLayout,
});

function AuthLayout() {
  const router = useRouter();
  const navigate = Route.useNavigate();
  const auth = useAuth();

  if (auth.isLoading || !auth.isAuthenticated) return null;

  const handleLogout = () => {
    if (window.confirm("Are you sure you want to logout?")) {
      auth.logout().then(() => {
        router.invalidate().finally(() => {
          navigate({ to: "/", replace: true });
        });
      });
    }
  };

  // React.useEffect(() => {
  //     if (!auth.isLoading && !auth.isAuthenticated) {
  //         navigate({ to: "/signin" });
  //     }
  // }, [auth.isLoading, auth.isAuthenticated]);

  return (
    <div className="p-2 h-full">
      <ul className="py-2 flex gap-2 justify-end">
        <li>
          <button
            type="button"
            className="flex w-full justify-center rounded-md bg-slate-900 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-xs 
                                hover:bg-slate-700 
                                focus-visible:outline-2 
                                focus-visible:outline-offset-2 
                                focus-visible:outline-indigo-600
                                "
            onClick={handleLogout}
          >
            Logout
          </button>
        </li>
      </ul>
      <hr />
      <Outlet />
    </div>
  );
}
