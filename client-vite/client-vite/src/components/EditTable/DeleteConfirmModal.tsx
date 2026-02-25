import { useRef, useEffect } from "react";
import { AlertTriangle } from "lucide-react";
import type { Purchase } from "../Summary/mockData";

interface DeleteConfirmModalProps {
  purchase: Purchase;
  onConfirm: () => void;
  onCancel: () => void;
  isDeleting: boolean;
}

export default function DeleteConfirmModal({
  purchase,
  onConfirm,
  onCancel,
  isDeleting,
}: Readonly<DeleteConfirmModalProps>) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el || el.open) return;
    el.showModal();
  }, []);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    function onBackdrop(e: MouseEvent) {
      if (e.target === el) onCancel();
    }
    function onDialogClose() {
      onCancel();
    }
    el.addEventListener("mousedown", onBackdrop);
    el.addEventListener("close", onDialogClose);
    return () => {
      el.removeEventListener("mousedown", onBackdrop);
      el.removeEventListener("close", onDialogClose);
    };
  }, [onCancel]);

  return (
    <dialog
      ref={dialogRef}
      className="w-full max-w-sm rounded-xl shadow-2xl p-0 backdrop:bg-black/40 mt-32 mb-auto mx-auto"
    >
      <div className="p-6">
        <div className="flex items-start gap-3 mb-5">
          <div className="shrink-0 p-2 rounded-full bg-red-100">
            <AlertTriangle size={20} className="text-red-600" />
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">Delete Purchase</h3>
            <p className="text-sm text-gray-500 mt-1">
              Are you sure you want to delete{" "}
              <span className="font-medium text-gray-700">
                "{purchase.title}"
              </span>
              ? This action cannot be undone.
            </p>
          </div>
        </div>

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            disabled={isDeleting}
            className="px-4 py-1.5 text-sm text-gray-700 rounded-md border border-gray-300 hover:bg-gray-50 disabled:opacity-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isDeleting}
            className="px-4 py-1.5 text-sm text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isDeleting ? "Deleting…" : "Delete"}
          </button>
        </div>
      </div>
    </dialog>
  );
}
