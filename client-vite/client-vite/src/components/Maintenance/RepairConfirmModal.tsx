import { useRef, useEffect } from "react";
import { Wrench } from "lucide-react";

interface RepairConfirmModalProps {
  count: number;
  onConfirm: () => void;
  onCancel: () => void;
  isRepairing: boolean;
}

export default function RepairConfirmModal({
  count,
  onConfirm,
  onCancel,
  isRepairing,
}: Readonly<RepairConfirmModalProps>) {
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
          <div className="shrink-0 p-2 rounded-full bg-sky-100">
            <Wrench size={20} className="text-sky-600" />
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">Repair documents</h3>
            <p className="text-sm text-gray-500 mt-1">
              Repair {count} selected issue{count === 1 ? "" : "s"}? Extra files
              will be moved to quarantine, not deleted.
            </p>
          </div>
        </div>

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            disabled={isRepairing}
            className="px-4 py-1.5 text-sm text-gray-700 rounded-md border border-gray-300 hover:bg-gray-50 disabled:opacity-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isRepairing}
            className="px-4 py-1.5 text-sm text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isRepairing ? "Repairing…" : "Repair"}
          </button>
        </div>
      </div>
    </dialog>
  );
}
