import { Pencil, Trash2 } from "lucide-react";
import { type Purchase, formatDate, formatCurrency } from "../Summary/mockData";

interface EditTableRowProps {
  purchase: Purchase;
  onEdit: (p: Purchase) => void;
  onDelete: (p: Purchase) => void;
}

export default function EditTableRow({
  purchase,
  onEdit,
  onDelete,
}: Readonly<EditTableRowProps>) {
  return (
    <>
      {/* Main row */}
      <tr className="hover:bg-gray-50 transition-colors">
        <td className="px-4 py-2.5">
          <div className="font-medium text-gray-900">{purchase.title}</div>
          {purchase.categories && purchase.categories.length > 0 && (
            <div className="mt-1 flex flex-wrap gap-1">
              {purchase.categories.map((cat) => (
                <span
                  key={cat.guid}
                  className="rounded-full px-2 py-0.5 text-xs font-medium text-white"
                  style={{ backgroundColor: cat.color || "#94a3b8" }}
                >
                  {cat.name}
                </span>
              ))}
            </div>
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
        <td className="px-4 py-2.5">
          <div className="flex justify-end gap-1">
            <button
              type="button"
              onClick={() => onEdit(purchase)}
              className="p-1.5 rounded-md text-gray-400 hover:text-sky-600 hover:bg-sky-50 transition-colors"
              aria-label={`Edit ${purchase.title}`}
            >
              <Pencil size={15} />
            </button>
            <button
              type="button"
              onClick={() => onDelete(purchase)}
              className="p-1.5 rounded-md text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
              aria-label={`Delete ${purchase.title}`}
            >
              <Trash2 size={15} />
            </button>
          </div>
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
