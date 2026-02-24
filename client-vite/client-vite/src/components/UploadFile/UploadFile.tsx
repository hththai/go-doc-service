import { uploadFile } from "@/api/upload";
import { scanInvoice, parseOcrDate } from "@/api/ocr";
import { useState, useRef, useEffect } from "react";
import { useForm } from "@tanstack/react-form";
import PurchaseInfo from "./PurchaseInfo/PurchaseInfo";

const defaultValues = {
  title: "",
  description: "",
  buyAt: "",
  buyFrom: "",
  buyPrice: "",
};

export default function UploadFile() {
  const [file, setFile] = useState<File | null>(null);
  const [items, setItems] = useState<
    {
      itemDescription: string;
      itemQty: string;
      unitPrice: string;
      subTotal: string;
    }[]
  >([]);
  const [success, setSuccess] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const previewDialogRef = useRef<HTMLDialogElement | null>(null);

  const form = useForm({
    defaultValues,
    onSubmit: async ({ value }) => {
      setUploadError(null);
      try {
        const ok = await uploadFile(value, items, file);
        if (ok) {
          setSuccess(true);
          handleClear();
          setTimeout(() => setSuccess(false), 2000);
        }
      } catch (err) {
        setUploadError(err instanceof Error ? err.message : "Upload failed");
      }
    },
  });

  useEffect(() => {
    if (!file) {
      setPreviewUrl(null);
      return;
    }
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  useEffect(() => {
    const dialog = previewDialogRef.current;
    if (!dialog) return;
    if (previewOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [previewOpen]);

  useEffect(() => {
    const dialog = previewDialogRef.current;
    if (!dialog) return;
    const handleBackdropClick = (e: MouseEvent) => {
      if (e.target === dialog) setPreviewOpen(false);
    };
    dialog.addEventListener("click", handleBackdropClick);
    return () => dialog.removeEventListener("click", handleBackdropClick);
  }, []);

  function handleClear() {
    form.reset();
    setFile(null);
    setItems([]);
    setScanError(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  async function handleScan() {
    if (!file) return;
    setScanning(true);
    setScanError(null);
    try {
      const { invoice } = await scanInvoice(file);
      // Pre-fill only empty fields so user input is not overridden.
      if (!form.getFieldValue("buyFrom")) {
        form.setFieldValue("buyFrom", invoice.seller ?? "");
      }
      if (!form.getFieldValue("buyPrice")) {
        form.setFieldValue("buyPrice", invoice.total ?? "");
      }
      if (!form.getFieldValue("buyAt")) {
        form.setFieldValue("buyAt", parseOcrDate(invoice.document_date) ?? "");
      }
      if (items.length === 0 && invoice.items?.length > 0) {
        setItems(
          invoice.items.map((i) => ({
            itemDescription: i.description,
            itemQty: i.qty,
            unitPrice: i.unit_price,
            subTotal: i.subtotal,
          })),
        );
      }
    } catch {
      setScanError("Could not extract invoice data. Please fill in manually.");
    } finally {
      setScanning(false);
    }
  }

  return (
    <div>
      {success && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="flex flex-col items-center gap-3 rounded-2xl bg-white px-12 py-10 shadow-xl">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <svg
                className="h-8 w-8 text-green-600"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M4.5 12.75l6 6 9-13.5"
                />
              </svg>
            </div>
            <p className="text-base font-semibold text-gray-900">
              Uploaded successfully
            </p>
          </div>
        </div>
      )}
      <div>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
        >
          <div className="space-y-12">
            <div className="border-b border-gray-900/10 pb-12">
              <div className="mt-10 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                {/* Title */}
                <form.Field
                  name="title"
                  validators={{
                    onBlur: ({ value }) =>
                      value.trim() ? undefined : "Title is required",
                  }}
                >
                  {(field) => (
                    <div className="col-span-full">
                      <label
                        htmlFor="title"
                        className="block text-sm/6 font-medium text-gray-900"
                      >
                        Title<span className="text-red-500">*</span>
                      </label>
                      <div className="mt-2">
                        <div className="flex items-center rounded-md bg-white pl-3 outline-1 -outline-offset-1 outline-gray-300 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-indigo-600">
                          <input
                            id="title"
                            name="title"
                            type="text"
                            value={field.state.value}
                            onBlur={field.handleBlur}
                            onChange={(e) => field.handleChange(e.target.value)}
                            className="block min-w-0 grow bg-white py-1.5 pr-3 pl-1 text-base text-gray-900 placeholder:text-gray-400 focus:outline-none sm:text-sm/6"
                          />
                        </div>
                      </div>
                      {field.state.meta.errors.length > 0 && (
                        <p className="mt-1 text-xs text-red-600">
                          {field.state.meta.errors.join(", ")}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>

                {/* Description */}
                <form.Field
                  name="description"
                  validators={{
                    onBlur: ({ value }) =>
                      value.length <= 500
                        ? undefined
                        : "Description must be 500 characters or fewer",
                  }}
                >
                  {(field) => (
                    <div className="col-span-full">
                      <label
                        htmlFor="description"
                        className="block text-sm/6 font-medium text-gray-900"
                      >
                        Description
                      </label>
                      <div className="mt-2">
                        <textarea
                          id="description"
                          name="description"
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          className="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                        />
                      </div>
                      {field.state.meta.errors.length > 0 && (
                        <p className="mt-1 text-xs text-red-600">
                          {field.state.meta.errors.join(", ")}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>

                <div className="col-span-full">
                  <label
                    htmlFor="upload"
                    className="block text-sm font-medium text-gray-900"
                  >
                    Upload Files
                  </label>

                  {/* Compact upload row */}
                  {file ? (
                    <div className="mt-2 rounded-md border border-gray-200 p-3">
                      <div className="flex items-center gap-3">
                        {/* Thumbnail or file icon — click to preview */}
                        <button
                          type="button"
                          onClick={() => setPreviewOpen(true)}
                          className="shrink-0 rounded focus:outline-none focus:ring-2 focus:ring-sky-500"
                          title="Preview file"
                        >
                          {file.type.startsWith("image/") ? (
                            <img
                              src={previewUrl ?? ""}
                              alt="Preview"
                              className="h-14 w-14 rounded object-cover border border-gray-200 hover:opacity-80 transition-opacity"
                            />
                          ) : (
                            <div className="h-14 w-14 flex items-center justify-center rounded border border-gray-200 bg-gray-50 hover:bg-gray-100 transition-colors">
                              <svg
                                viewBox="0 0 24 24"
                                fill="currentColor"
                                className="size-7 text-gray-400"
                              >
                                <path
                                  fillRule="evenodd"
                                  d="M5.625 1.5H9a3.75 3.75 0 0 1 3.75 3.75v1.875c0 1.036.84 1.875 1.875 1.875H16.5a3.75 3.75 0 0 1 3.75 3.75v7.875c0 1.035-.84 1.875-1.875 1.875H5.625a1.875 1.875 0 0 1-1.875-1.875V3.375c0-1.036.84-1.875 1.875-1.875Zm5.845 17.03a.75.75 0 0 0 1.06 0l3-3a.75.75 0 1 0-1.06-1.06l-1.72 1.72V12a.75.75 0 0 0-1.5 0v4.19l-1.72-1.72a.75.75 0 0 0-1.06 1.06l3 3Z"
                                  clipRule="evenodd"
                                />
                                <path d="M14.25 5.25a5.23 5.23 0 0 0-1.279-3.434 9.768 9.768 0 0 1 6.963 6.963A5.23 5.23 0 0 0 16.5 7.5h-1.875a.375.375 0 0 1-.375-.375V5.25Z" />
                              </svg>
                            </div>
                          )}
                        </button>

                        {/* File info + actions */}
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium text-gray-900">
                            {file.name}
                          </p>
                          <p className="text-xs text-gray-400">
                            {(file.size / 1024).toFixed(0)} KB
                          </p>
                        </div>

                        <div className="flex shrink-0 items-center gap-2">
                          <button
                            type="button"
                            onClick={handleScan}
                            disabled={scanning}
                            className="inline-flex items-center gap-1.5 rounded-md bg-sky-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-sky-600 disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            {scanning ? (
                              <>
                                <svg
                                  className="animate-spin h-3 w-3"
                                  viewBox="0 0 24 24"
                                  fill="none"
                                >
                                  <circle
                                    className="opacity-25"
                                    cx="12"
                                    cy="12"
                                    r="10"
                                    stroke="currentColor"
                                    strokeWidth="4"
                                  />
                                  <path
                                    className="opacity-75"
                                    fill="currentColor"
                                    d="M4 12a8 8 0 018-8v8H4z"
                                  />
                                </svg>
                                Scanning…
                              </>
                            ) : (
                              "Scan"
                            )}
                          </button>
                          <button
                            type="button"
                            onClick={handleClear}
                            className="rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100"
                            title="Remove file"
                          >
                            <svg
                              viewBox="0 0 20 20"
                              fill="currentColor"
                              className="size-4"
                            >
                              <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <label
                      htmlFor="file"
                      aria-label="Upload a file"
                      className="mt-2 flex items-center gap-3 cursor-pointer rounded-md border border-dashed border-gray-300 px-4 py-3 hover:border-gray-400 hover:bg-gray-50 transition-colors"
                    >
                      <svg
                        viewBox="0 0 24 24"
                        fill="currentColor"
                        className="size-5 shrink-0 text-gray-400"
                      >
                        <path d="M11.47 1.72a.75.75 0 0 1 1.06 0l3 3a.75.75 0 0 1-1.06 1.06l-1.72-1.72V7.5h-1.5V4.06L9.53 5.78a.75.75 0 0 1-1.06-1.06l3-3ZM11.25 7.5V15a.75.75 0 0 0 1.5 0V7.5h3.75a3 3 0 0 1 3 3v6.75a3 3 0 0 1-3 3H6.75a3 3 0 0 1-3-3V10.5a3 3 0 0 1 3-3h4.5Z" />
                      </svg>
                      <span className="text-sm text-gray-600">
                        <span className="font-medium text-slate-900">
                          Choose file
                        </span>
                        <span className="ml-1 text-gray-400">
                          or drag and drop
                        </span>
                      </span>
                      <input
                        id="file"
                        type="file"
                        ref={fileInputRef}
                        className="sr-only"
                        onChange={(e) => setFile(e.target.files?.[0] || null)}
                      />
                    </label>
                  )}

                  {/* Scan error */}
                  {scanError && (
                    <p className="mt-2 text-sm text-red-600">{scanError}</p>
                  )}
                </div>
              </div>

              <form.Subscribe selector={(state) => state.values}>
                {(values) => (
                  <PurchaseInfo
                    values={{
                      buyAt: values.buyAt,
                      buyFrom: values.buyFrom,
                      buyPrice: values.buyPrice,
                    }}
                    items={items}
                    onChange={(name, value) =>
                      form.setFieldValue(
                        name as keyof typeof defaultValues,
                        value,
                      )
                    }
                    onItemsChange={setItems}
                  />
                )}
              </form.Subscribe>
            </div>
          </div>

          {uploadError && (
            <p className="mt-4 text-sm text-red-600">{uploadError}</p>
          )}

          <div className="mt-6 flex items-center justify-end gap-x-6">
            <button
              type="button"
              onClick={handleClear}
              className="text-sm font-semibold text-gray-900 px-3 py-2 rounded-md
                    hover:bg-gray-200
                    focus-visible:outline-2
                    focus-visible:outline-offset-2
                    focus-visible:outline-indigo-600"
            >
              Clear
            </button>
            <button
              type="submit"
              className="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white shadow-xs
                    hover:bg-slate-600
                    focus-visible:outline-2
                    focus-visible:outline-offset-2
                    focus-visible:outline-indigo-600"
            >
              Save
            </button>
          </div>
        </form>
      </div>

      {/* Preview modal */}
      <dialog
        ref={previewDialogRef}
        onClose={() => setPreviewOpen(false)}
        className="m-auto w-full max-w-3xl bg-transparent p-4 backdrop:bg-black/80"
      >
        {file && (
          <div className="relative">
            <button
              type="button"
              onClick={() => setPreviewOpen(false)}
              className="absolute -top-10 right-0 flex items-center gap-1 text-sm text-white/80 hover:text-white"
            >
              <svg viewBox="0 0 20 20" fill="currentColor" className="size-5">
                <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
              </svg>
              Close
            </button>

            {file.type.startsWith("image/") && (
              <img
                src={previewUrl ?? ""}
                alt={file.name}
                className="max-h-[85vh] w-full rounded object-contain"
              />
            )}

            {file.type === "application/pdf" && (
              <embed
                src={previewUrl ?? ""}
                type="application/pdf"
                className="h-[85vh] w-full rounded"
              />
            )}

            {!file.type.startsWith("image/") &&
              file.type !== "application/pdf" && (
                <div className="flex flex-col items-center gap-3 rounded bg-white px-8 py-12 text-center">
                  <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="size-12 text-gray-400"
                  >
                    <path
                      fillRule="evenodd"
                      d="M5.625 1.5H9a3.75 3.75 0 0 1 3.75 3.75v1.875c0 1.036.84 1.875 1.875 1.875H16.5a3.75 3.75 0 0 1 3.75 3.75v7.875c0 1.035-.84 1.875-1.875 1.875H5.625a1.875 1.875 0 0 1-1.875-1.875V3.375c0-1.036.84-1.875 1.875-1.875Zm5.845 17.03a.75.75 0 0 0 1.06 0l3-3a.75.75 0 1 0-1.06-1.06l-1.72 1.72V12a.75.75 0 0 0-1.5 0v4.19l-1.72-1.72a.75.75 0 0 0-1.06 1.06l3 3Z"
                      clipRule="evenodd"
                    />
                    <path d="M14.25 5.25a5.23 5.23 0 0 0-1.279-3.434 9.768 9.768 0 0 1 6.963 6.963A5.23 5.23 0 0 0 16.5 7.5h-1.875a.375.375 0 0 1-.375-.375V5.25Z" />
                  </svg>
                  <p className="font-medium text-gray-900">{file.name}</p>
                  <p className="text-sm text-gray-500">
                    No preview available for this file type.
                  </p>
                </div>
              )}
          </div>
        )}
      </dialog>
    </div>
  );
}
