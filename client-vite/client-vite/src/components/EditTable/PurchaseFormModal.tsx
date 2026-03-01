import { useState, useRef, useEffect } from "react";
import { X, Plus, Trash2, Paperclip, ExternalLink } from "lucide-react";
import { type Purchase, formatCurrency } from "../Summary/mockData";
import { scanInvoice } from "@/api/ocr";

export type PurchaseFormData = {
  title: string;
  buyFrom: string;
  buyAt: string;
  buyPrice: string;
  filename: string;
  items: Array<{
    itemName: string;
    itemQty: string;
    unitPrice: string;
    subTotal: string;
  }>;
};

interface PurchaseFormModalProps {
  mode: "create" | "edit";
  purchase?: Purchase;
  onSave: (data: PurchaseFormData, file?: File | null) => void;
  onClose: () => void;
  isSaving: boolean;
}

const inputClass =
  "w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-xs focus:outline-none focus:ring-2 focus:ring-sky-500";
const labelClass = "block text-xs text-gray-500 mb-1";

export default function PurchaseFormModal({
  mode,
  purchase,
  onSave,
  onClose,
  isSaving,
}: Readonly<PurchaseFormModalProps>) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const previewDialogRef = useRef<HTMLDialogElement>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const [form, setForm] = useState<PurchaseFormData>(() => {
    if (mode === "edit" && purchase) {
      return {
        title: purchase.title,
        buyFrom: purchase.buyFrom,
        buyAt: purchase.buyAt,
        buyPrice: purchase.buyPrice,
        filename: purchase.filename ?? "",
        items: purchase.items.map((i) => ({ ...i })),
      };
    }
    return {
      title: "",
      buyFrom: "",
      buyAt: "",
      buyPrice: "",
      filename: "",
      items: [],
    };
  });

  const [file, setFile] = useState<File | null>(null);
  const [scanning, setScanning] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el || el.open) return;
    el.showModal();
  }, []);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    function onBackdrop(e: MouseEvent) {
      if (e.target === el) onClose();
    }
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
    } else if (dialog.open) {
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

  function updateField(
    field: keyof Omit<PurchaseFormData, "items">,
    value: string,
  ) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  function addItem() {
    setForm((prev) => ({
      ...prev,
      items: [
        ...prev.items,
        { itemName: "", itemQty: "", unitPrice: "", subTotal: "" },
      ],
    }));
  }

  function removeItem(index: number) {
    setForm((prev) => ({
      ...prev,
      items: prev.items.filter((_, i) => i !== index),
    }));
  }

  function updateItem(
    index: number,
    field: keyof PurchaseFormData["items"][number],
    value: string,
  ) {
    setForm((prev) => {
      const items = prev.items.map((item, i) => {
        if (i !== index) return item;
        const updated = { ...item, [field]: value };
        if (field === "itemQty" || field === "unitPrice") {
          const qty = Number.parseFloat(
            field === "itemQty" ? value : item.itemQty,
          );
          const price = Number.parseFloat(
            field === "unitPrice" ? value : item.unitPrice,
          );
          updated.subTotal =
            !Number.isNaN(qty) && !Number.isNaN(price)
              ? (qty * price).toFixed(2)
              : "";
        }
        return updated;
      });
      return { ...prev, items };
    });
  }

  async function handleScan() {
    if (!file) return;
    setScanning(true);
    setScanError(null);
    try {
      const { invoice, purchaseInfo } = await scanInvoice(file);
      setForm((prev) => {
        const updated = { ...prev };
        if (!prev.buyFrom) updated.buyFrom = invoice.seller ?? "";
        if (!prev.buyPrice) updated.buyPrice = invoice.total ?? "";
        if (!prev.buyAt && purchaseInfo.buyAt) {
          const [d, m, y] = purchaseInfo.buyAt.split("/");
          updated.buyAt = `${y}-${m}-${d}`;
        }
        if (prev.items.length === 0 && invoice.items?.length > 0) {
          updated.items = invoice.items.map((i) => ({
            itemName: i.description,
            itemQty: i.qty,
            unitPrice: i.unit_price,
            subTotal: i.subtotal,
          }));
        }
        return updated;
      });
    } catch {
      setScanError("Could not extract invoice data. Please fill in manually.");
    } finally {
      setScanning(false);
    }
  }

  function removeFile() {
    setFile(null);
    setScanError(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  const itemsTotal = form.items.reduce((sum, item) => {
    const sub = Number.parseFloat(item.subTotal);
    return sum + (Number.isNaN(sub) ? 0 : sub);
  }, 0);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSave(form, file);
  }

  const hasAttachment = mode === "edit" && purchase?.filename;

  return (
    <>
      <dialog
        ref={dialogRef}
        className="w-full max-w-lg rounded-xl shadow-2xl p-0 backdrop:bg-black/40 mt-12 mb-auto mx-auto"
      >
        <form onSubmit={handleSubmit}>
          {/* Header */}
          <div className="flex items-center justify-between bg-sky-500 text-white px-4 py-3 rounded-t-xl">
            <h2 className="font-semibold text-base">
              {mode === "create" ? "New Purchase" : "Edit Purchase"}
            </h2>
            <button
              type="button"
              onClick={onClose}
              className="shrink-0 p-1 rounded-lg hover:bg-sky-600 transition-colors"
              aria-label="Close"
            >
              <X size={18} />
            </button>
          </div>

          {/* Form body */}
          <div className="px-4 py-4 space-y-4 max-h-[65vh] overflow-y-auto">
            {/* Title */}
            <div>
              <label htmlFor="form-title" className={labelClass}>
                Title <span className="text-red-500">*</span>
              </label>
              <input
                id="form-title"
                type="text"
                value={form.title}
                onChange={(e) => updateField("title", e.target.value)}
                required
                placeholder="e.g. Weekly Groceries"
                className={inputClass}
              />
            </div>

            {/* Store + Date */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="form-store" className={labelClass}>
                  Store <span className="text-red-500">*</span>
                </label>
                <input
                  id="form-store"
                  type="text"
                  value={form.buyFrom}
                  onChange={(e) => updateField("buyFrom", e.target.value)}
                  required
                  placeholder="e.g. Woolworths"
                  className={inputClass}
                />
              </div>
              <div>
                <label htmlFor="form-date" className={labelClass}>
                  Date <span className="text-red-500">*</span>
                </label>
                <input
                  id="form-date"
                  type="date"
                  value={form.buyAt}
                  onChange={(e) => updateField("buyAt", e.target.value)}
                  required
                  className={inputClass}
                />
              </div>
            </div>

            {/* Total Price */}
            <div>
              <label htmlFor="form-price" className={labelClass}>
                Total Price <span className="text-red-500">*</span>
                {form.items.length > 0 && (
                  <button
                    type="button"
                    onClick={() =>
                      updateField("buyPrice", itemsTotal.toFixed(2))
                    }
                    className="ml-2 text-sky-600 hover:text-sky-800 hover:underline"
                  >
                    Use items total ({formatCurrency(itemsTotal)})
                  </button>
                )}
              </label>
              <input
                id="form-price"
                type="number"
                min="0"
                step="0.01"
                value={form.buyPrice}
                onChange={(e) => updateField("buyPrice", e.target.value)}
                required
                placeholder="0.00"
                className={inputClass}
              />
            </div>

            {/* File upload (create mode only) */}
            {mode === "create" && (
              <div>
                <span className={labelClass}>Attach File</span>

                {file ? (
                  <div className="rounded-md border border-gray-200 p-3">
                    <div className="flex items-center gap-3">
                      {/* Thumbnail / file icon — click to preview */}
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

                      {/* File info */}
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-gray-900">
                          {file.name}
                        </p>
                        <p className="text-xs text-gray-400">
                          {(file.size / 1024).toFixed(0)} KB
                        </p>
                      </div>

                      {/* Actions */}
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
                          onClick={removeFile}
                          className="rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100"
                          title="Remove file"
                        >
                          <X size={14} />
                        </button>
                      </div>
                    </div>
                  </div>
                ) : (
                  <label
                    htmlFor="modal-file"
                    aria-label="Upload a file"
                    className="mt-1 flex items-center gap-3 cursor-pointer rounded-md border border-dashed border-gray-300 px-4 py-3 hover:border-gray-400 hover:bg-gray-50 transition-colors"
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
                      id="modal-file"
                      type="file"
                      ref={fileInputRef}
                      className="sr-only"
                      onChange={(e) => setFile(e.target.files?.[0] || null)}
                    />
                  </label>
                )}

                {scanError && (
                  <p className="mt-1 text-xs text-red-600">{scanError}</p>
                )}
              </div>
            )}

            {/* File attachment / replacement (edit mode) */}
            {mode === "edit" && (
              <div>
                <span className={labelClass}>Attached Document</span>

                {file ? (
                  /* New replacement file selected */
                  <div className="rounded-md border border-gray-200 p-3">
                    <div className="flex items-center gap-3">
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

                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-gray-900">
                          {file.name}
                        </p>
                        <p className="text-xs text-gray-400">
                          {(file.size / 1024).toFixed(0)} KB · replaces current
                          file
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
                          onClick={removeFile}
                          className="rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100"
                          title="Cancel replacement"
                        >
                          <X size={14} />
                        </button>
                      </div>
                    </div>
                  </div>
                ) : hasAttachment ? (
                  /* Existing file — show info + Replace button */
                  <div className="flex items-center gap-2 rounded-md border border-gray-200 px-3 py-2">
                    <Paperclip size={14} className="shrink-0 text-gray-400" />
                    <span className="flex-1 truncate text-sm text-gray-700">
                      {form.filename}
                    </span>
                    {purchase?.fileUrl && (
                      <a
                        href={purchase.fileUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="shrink-0 inline-flex items-center gap-1 px-2.5 py-1.5 text-xs rounded-md border border-gray-300 text-gray-500 hover:text-sky-600 hover:border-sky-300 transition-colors"
                        title="Open attached file"
                      >
                        <ExternalLink size={13} />
                        View
                      </a>
                    )}
                    <label
                      htmlFor="edit-replace-file"
                      className="shrink-0 inline-flex items-center gap-1 cursor-pointer px-2.5 py-1.5 text-xs rounded-md border border-gray-300 text-gray-500 hover:text-sky-600 hover:border-sky-300 transition-colors"
                      title="Upload a replacement file"
                    >
                      Replace
                      <input
                        id="edit-replace-file"
                        type="file"
                        ref={fileInputRef}
                        className="sr-only"
                        onChange={(e) => setFile(e.target.files?.[0] || null)}
                      />
                    </label>
                  </div>
                ) : (
                  /* No existing file — show upload dropzone */
                  <label
                    htmlFor="edit-attach-file"
                    aria-label="Upload a file"
                    className="mt-1 flex items-center gap-3 cursor-pointer rounded-md border border-dashed border-gray-300 px-4 py-3 hover:border-gray-400 hover:bg-gray-50 transition-colors"
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
                      id="edit-attach-file"
                      type="file"
                      ref={fileInputRef}
                      className="sr-only"
                      onChange={(e) => setFile(e.target.files?.[0] || null)}
                    />
                  </label>
                )}

                {scanError && (
                  <p className="mt-1 text-xs text-red-600">{scanError}</p>
                )}
              </div>
            )}

            {/* Line Items */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-gray-500 uppercase font-medium">
                  Line Items
                </span>
                <button
                  type="button"
                  onClick={addItem}
                  className="flex items-center gap-1 text-xs text-sky-600 hover:text-sky-800 transition-colors"
                >
                  <Plus size={14} />
                  Add Item
                </button>
              </div>

              {form.items.length === 0 ? (
                <p className="text-xs text-gray-400 text-center py-4 border border-dashed border-gray-200 rounded-md">
                  No items yet
                </p>
              ) : (
                <div className="space-y-2">
                  {/* Column headers */}
                  <div className="grid grid-cols-[1fr_4rem_5rem_5rem_2rem] gap-2">
                    <span className="text-xs text-gray-400">Description</span>
                    <span className="text-xs text-gray-400">Qty</span>
                    <span className="text-xs text-gray-400">Unit Price</span>
                    <span className="text-xs text-gray-400">Subtotal</span>
                    <span />
                  </div>

                  {form.items.map((item, index) => (
                    // biome-ignore lint/suspicious/noArrayIndexKey: items use positional index during editing
                    <div
                      key={index}
                      className="grid grid-cols-[1fr_4rem_5rem_5rem_2rem] gap-2 items-center"
                    >
                      <input
                        type="text"
                        value={item.itemName}
                        onChange={(e) =>
                          updateItem(index, "itemName", e.target.value)
                        }
                        placeholder="Description"
                        aria-label={`Item ${index + 1} name`}
                        className={inputClass}
                        required
                      />
                      <input
                        type="number"
                        min="0"
                        step="1"
                        value={item.itemQty}
                        onChange={(e) =>
                          updateItem(index, "itemQty", e.target.value)
                        }
                        placeholder="1"
                        aria-label={`Item ${index + 1} quantity`}
                        className={inputClass}
                        required
                      />
                      <input
                        type="number"
                        min="0"
                        step="0.01"
                        value={item.unitPrice}
                        onChange={(e) =>
                          updateItem(index, "unitPrice", e.target.value)
                        }
                        placeholder="0.00"
                        aria-label={`Item ${index + 1} unit price`}
                        className={inputClass}
                      />
                      <input
                        type="text"
                        value={
                          item.subTotal ? formatCurrency(item.subTotal) : "—"
                        }
                        readOnly
                        aria-label={`Item ${index + 1} subtotal`}
                        className="w-full rounded-md border border-gray-200 bg-gray-50 px-3 py-1.5 text-sm text-gray-500 text-right"
                      />
                      <button
                        type="button"
                        onClick={() => removeItem(index)}
                        className="p-1.5 rounded-md text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                        aria-label={`Remove item ${index + 1}`}
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="flex justify-end gap-3 px-4 py-3 border-t border-gray-200">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-1.5 text-sm text-gray-700 rounded-md border border-gray-300 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="px-4 py-1.5 text-sm text-white bg-sky-500 rounded-md hover:bg-sky-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isSaving ? "Saving…" : "Save"}
            </button>
          </div>
        </form>
      </dialog>

      {/* File preview dialog */}
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
    </>
  );
}
