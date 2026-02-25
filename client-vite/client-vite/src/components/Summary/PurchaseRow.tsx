import { ChevronRight } from "lucide-react";
import { type Purchase, formatDate, formatCurrency } from "./mockData";

interface PurchaseRowProps {
  purchase: Purchase;
  onSelect: (p: Purchase) => void;
  onPreview: (url: string, filename: string) => void;
}

export default function PurchaseRow({
  purchase,
  onSelect,
  onPreview,
}: Readonly<PurchaseRowProps>) {
  return (
    <>
      {/* Main row — tappable on mobile */}
      <tr
        tabIndex={0}
        onClick={() => onSelect(purchase)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") onSelect(purchase);
        }}
        className="hover:bg-gray-50 transition-colors cursor-pointer sm:cursor-default"
      >
        <td className="px-4 py-2.5 text-gray-900 font-medium">
          {purchase.title}
        </td>
        <td className="hidden sm:table-cell px-4 py-2.5 text-xs w-36 max-w-36 overflow-hidden">
          {purchase.fileUrl ? (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onPreview(purchase.fileUrl!, purchase.filename);
              }}
              className="block w-full truncate text-sky-600 hover:text-sky-800 hover:underline text-left"
            >
              {purchase.filename}
            </button>
          ) : (
            <span className="block w-full truncate text-gray-500">
              {purchase.filename}
            </span>
          )}
        </td>
        <td className="hidden sm:table-cell px-4 py-2.5 text-gray-600">
          {formatDate(purchase.buyAt)}
        </td>
        <td className="hidden sm:table-cell px-4 py-2.5 text-gray-600">
          {purchase.buyFrom}
        </td>
        <td className="px-4 py-2.5 text-right text-gray-900 font-medium">
          {formatCurrency(purchase.buyPrice)}
        </td>
        <td className="hidden sm:table-cell px-4 py-2.5 text-center text-gray-500">
          {purchase.items.length}
        </td>
        {/* Chevron visible only on mobile */}
        <td className="sm:hidden px-2 py-2.5 text-gray-400">
          <ChevronRight size={16} />
        </td>
      </tr>

      {/* Inline sub-table — desktop only */}
      {purchase.items.length > 0 && (
        <tr className="hidden sm:table-row bg-gray-50/60">
          <td colSpan={6} className="px-4 pb-3 pt-0">
            <div className="ml-4 overflow-x-auto rounded border border-gray-100">
              <table className="min-w-full text-xs">
                <thead className="text-gray-400 uppercase">
                  <tr>
                    <th className="px-3 py-1.5 text-left font-medium">
                      Description
                    </th>
                    <th className="px-3 py-1.5 text-left font-medium w-16">
                      Qty
                    </th>
                    <th className="px-3 py-1.5 text-right font-medium w-24">
                      Unit Price
                    </th>
                    <th className="px-3 py-1.5 text-right font-medium w-24">
                      Subtotal
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {purchase.items.map((item) => (
                    <tr key={item.itemName} className="text-gray-600">
                      <td className="px-3 py-1.5">{item.itemName}</td>
                      <td className="px-3 py-1.5">{item.itemQty}</td>
                      <td className="px-3 py-1.5 text-right">
                        {formatCurrency(item.unitPrice)}
                      </td>
                      <td className="px-3 py-1.5 text-right font-medium text-gray-700">
                        {formatCurrency(item.subTotal)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}
