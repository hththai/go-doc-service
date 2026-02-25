import { X } from "lucide-react";
import { type Purchase, formatDate, formatCurrency } from "./mockData";

interface PurchaseDetailProps {
  purchase: Purchase;
  onClose: () => void;
  onPreview: (url: string, filename: string) => void;
}

export default function PurchaseDetail({
  purchase,
  onClose,
  onPreview,
}: Readonly<PurchaseDetailProps>) {
  const itemTotal = purchase.items.reduce((sum, item) => {
    const sub = Number.parseFloat(item.subTotal);
    return sum + (Number.isNaN(sub) ? 0 : sub);
  }, 0);

  return (
    <div className="flex flex-col max-h-[85vh]">
      {/* Modal header */}
      <div className="flex items-center justify-between bg-sky-500 text-white px-4 py-3 rounded-t-xl">
        <h2 className="font-semibold text-base truncate pr-2">
          {purchase.title}
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

      {/* Meta info */}
      <div className="px-4 py-3 space-y-1.5 border-b border-gray-100 text-sm">
        <div className="flex justify-between text-gray-600">
          <span className="text-gray-400">Store</span>
          <span>{purchase.buyFrom}</span>
        </div>
        <div className="flex justify-between text-gray-600">
          <span className="text-gray-400">Date</span>
          <span>{formatDate(purchase.buyAt)}</span>
        </div>
        <div className="flex justify-between text-gray-600">
          <span className="text-gray-400">File</span>
          {purchase.fileUrl ? (
            <button
              type="button"
              onClick={() => onPreview(purchase.fileUrl!, purchase.filename)}
              className="truncate max-w-48 text-sky-600 hover:text-sky-800 hover:underline text-right"
            >
              {purchase.filename}
            </button>
          ) : (
            <span className="truncate max-w-48 text-right">
              {purchase.filename}
            </span>
          )}
        </div>
        <div className="flex justify-between font-semibold text-gray-900">
          <span>Total</span>
          <span>{formatCurrency(purchase.buyPrice)}</span>
        </div>
      </div>

      {/* Line items */}
      {purchase.items.length > 0 && (
        <div className="flex-1 overflow-y-auto px-4 py-3">
          <p className="text-xs text-gray-400 uppercase font-medium mb-2">
            Line Items
          </p>
          <div className="rounded-md border border-gray-200 overflow-hidden">
            <table className="min-w-full text-xs">
              <thead className="bg-gray-50 text-gray-500 uppercase">
                <tr>
                  <th className="px-3 py-1.5 text-left font-medium">
                    Description
                  </th>
                  <th className="px-3 py-1.5 text-center font-medium w-10">
                    Qty
                  </th>
                  <th className="px-3 py-1.5 text-right font-medium w-20">
                    Subtotal
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {purchase.items.map((item) => (
                  <tr key={item.itemName} className="text-gray-600">
                    <td className="px-3 py-2">{item.itemName}</td>
                    <td className="px-3 py-2 text-center">{item.itemQty}</td>
                    <td className="px-3 py-2 text-right font-medium text-gray-700">
                      {formatCurrency(item.subTotal)}
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr className="bg-gray-50 border-t border-gray-200">
                  <td
                    colSpan={2}
                    className="px-3 py-1.5 text-xs font-semibold text-gray-600"
                  >
                    Items Total
                  </td>
                  <td className="px-3 py-1.5 text-right text-xs font-semibold text-gray-900">
                    {formatCurrency(itemTotal)}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
