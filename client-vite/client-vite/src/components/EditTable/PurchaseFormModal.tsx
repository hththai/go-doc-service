import { useState, useRef, useEffect } from "react";
import { X, Plus, Trash2, Paperclip, ExternalLink } from "lucide-react";
import { type Purchase, formatCurrency } from "../Summary/mockData";

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
  onSave: (data: PurchaseFormData) => void;
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

  const itemsTotal = form.items.reduce((sum, item) => {
    const sub = Number.parseFloat(item.subTotal);
    return sum + (Number.isNaN(sub) ? 0 : sub);
  }, 0);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSave(form);
  }

  const hasAttachment = mode === "edit" && purchase?.filename;

  return (
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
                  onClick={() => updateField("buyPrice", itemsTotal.toFixed(2))}
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

          {/* Attached Document */}
          {hasAttachment && (
            <div>
              <label htmlFor="form-filename" className={labelClass}>
                Attached Document
              </label>
              <div className="flex items-center gap-2">
                <Paperclip size={14} className="shrink-0 text-gray-400" />
                <input
                  id="form-filename"
                  type="text"
                  value={form.filename}
                  onChange={(e) => updateField("filename", e.target.value)}
                  required
                  placeholder="File name"
                  className={inputClass}
                />
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
              </div>
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
  );
}
