import { useRef, useEffect } from "react";
import { X } from "lucide-react";

interface FilePreviewModalProps {
  url: string;
  filename: string;
  onClose: () => void;
}

function getFileType(filename: string): "image" | "pdf" | "unknown" {
  const ext = filename.split(".").pop()?.toLowerCase() ?? "";
  if (["jpg", "jpeg", "png", "webp", "gif"].includes(ext)) return "image";
  if (ext === "pdf") return "pdf";
  return "unknown";
}

export default function FilePreviewModal({
  url,
  filename,
  onClose,
}: Readonly<FilePreviewModalProps>) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const fileType = getFileType(filename);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el || el.open) return;
    el.showModal();
    // No cleanup — removing el.close() here prevents Strict Mode's cleanup
    // from firing the native `close` event and immediately unmounting this component.
    // The element is removed from the top layer automatically when unmounted from the DOM.
  }, []);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    function onBackdrop(e: MouseEvent) {
      if (e.target === el) onClose();
    }
    // Handle Escape key via the native close event instead of onClose prop
    function onDialogClose() {
      onClose();
    }
    el.addEventListener("mousedown", onBackdrop);
    el.addEventListener("close", onDialogClose);
    return () => {
      el.removeEventListener("mousedown", onBackdrop);
      el.removeEventListener("close", onDialogClose);
    };
  }, [onClose]);

  return (
    <dialog
      ref={dialogRef}
      className="m-auto w-full max-w-3xl bg-transparent p-4 backdrop:bg-black/80"
    >
      <div className="relative flex flex-col gap-3">
        {/* Close button */}
        <button
          type="button"
          onClick={onClose}
          className="absolute -top-1 -right-1 z-10 rounded-full bg-white/10 p-1.5 text-white hover:bg-white/20 transition-colors"
          aria-label="Close preview"
        >
          <X size={20} />
        </button>

        {/* Filename label */}
        <p className="text-sm text-white/70 truncate pr-8">{filename}</p>

        {/* Preview content */}
        {fileType === "image" && (
          <img
            src={url}
            alt={filename}
            className="max-h-[85vh] w-full rounded object-contain"
          />
        )}

        {fileType === "pdf" && (
          <iframe
            src={url}
            title={filename}
            className="h-[85vh] w-full rounded bg-white"
          />
        )}

        {fileType === "unknown" && (
          <div className="flex flex-col items-center gap-3 rounded bg-white px-8 py-12 text-center">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth={1.5}
              className="size-12 text-gray-400"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"
              />
            </svg>
            <p className="font-medium text-gray-900">{filename}</p>
            <p className="text-sm text-gray-500">
              No preview available for this file type.
            </p>
          </div>
        )}
      </div>
    </dialog>
  );
}
