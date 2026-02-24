import Summary from "@/components/Summary/Summary";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_auth/summary")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex justify-center px-4">
      <div className="w-full max-w-md sm:max-w-4xl">
        <Summary />
      </div>
    </div>
  );
}
