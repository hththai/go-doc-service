import UploadFile from "@/components/UploadFile/UploadFile";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_auth/upload")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex justify-center px-4">
      <div className="w-full max-w-md sm:max-w-2xl">
        <UploadFile />
      </div>
    </div>
  );
}
