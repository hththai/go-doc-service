type Item = {
  itemName: string;
  itemQty: string;
  unitPrice: string;
  subTotal: string;
};

interface PurchaseInfoProps {
  values: {
    buyAt: string;
    buyFrom: string;
    buyPrice: string;
  };
  items: Item[];
  onChange: (name: string, value: string) => void;
  onItemsChange: (items: Item[]) => void;
}

const inputClass =
  "block min-w-0 grow bg-white py-1.5 pr-3 pl-1 text-base text-gray-900 placeholder:text-gray-400 focus:outline-none sm:text-sm/6";
const wrapperClass =
  "flex items-center rounded-md bg-white pl-3 outline-1 -outline-offset-1 outline-gray-300 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-indigo-600";
const labelClass = "block text-sm/6 font-medium text-gray-900";

import { useRef } from "react";

const emptyItem = (): Item => ({
  itemName: "",
  itemQty: "",
  unitPrice: "",
  subTotal: "",
});

export default function PurchaseInfo({
  values,
  items,
  onChange,
  onItemsChange,
}: Readonly<PurchaseInfoProps>) {
  const keyCounterRef = useRef(0);
  const itemKeysRef = useRef<string[]>([]);

  // Sync stable keys with the current items array length.
  while (itemKeysRef.current.length < items.length) {
    itemKeysRef.current.push(`item-${keyCounterRef.current++}`);
  }
  itemKeysRef.current = itemKeysRef.current.slice(0, items.length);

  function updateItem(index: number, field: keyof Item, value: string) {
    const updated = items.map((item, i) =>
      i === index ? { ...item, [field]: value } : item,
    );
    onItemsChange(updated);
  }

  function removeItem(index: number) {
    itemKeysRef.current.splice(index, 1);
    onItemsChange(items.filter((_, i) => i !== index));
  }

  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Purchase Info
      </div>
      <div className="border border-gray-200 rounded-b-md p-4 space-y-4">
        <div className="col-span-full">
          <label htmlFor="buyPrice" className={labelClass}>
            Purchase Price
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyPrice"
                name="buyPrice"
                type="number"
                min="0"
                step="0.01"
                placeholder="0.00"
                value={values.buyPrice}
                onChange={(e) => onChange("buyPrice", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        <div className="col-span-full">
          <label htmlFor="buyFrom" className={labelClass}>
            Purchase From
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyFrom"
                name="buyFrom"
                type="text"
                placeholder="e.g. Woolworths"
                value={values.buyFrom}
                onChange={(e) => onChange("buyFrom", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        <div className="col-span-full">
          <label htmlFor="buyAt" className={labelClass}>
            Purchase Date
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyAt"
                name="buyAt"
                type="date"
                value={values.buyAt}
                onChange={(e) => onChange("buyAt", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        {/* Line Items */}
        <div className="col-span-full">
          <div className="flex items-center justify-between mb-2">
            <span className={labelClass}>Line Items</span>
            <button
              type="button"
              onClick={() => onItemsChange([...items, emptyItem()])}
              className="text-xs font-medium text-sky-600 hover:text-sky-800"
            >
              + Add item
            </button>
          </div>

          {items.length > 0 && (
            <div className="overflow-x-auto rounded-md border border-gray-200">
              <table className="min-w-full text-sm">
                <thead className="bg-gray-50 text-xs text-gray-500 uppercase">
                  <tr>
                    <th className="px-3 py-2 text-left font-medium">
                      Description
                    </th>
                    <th className="px-3 py-2 text-left font-medium w-20">
                      Qty
                    </th>
                    <th className="px-3 py-2 text-left font-medium w-28">
                      Unit Price
                    </th>
                    <th className="px-3 py-2 text-left font-medium w-28">
                      Subtotal
                    </th>
                    <th className="w-8" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {items.map((item, i) => (
                    <tr key={itemKeysRef.current[i]}>
                      <td className="px-3 py-1.5">
                        <input
                          type="text"
                          value={item.itemName}
                          onChange={(e) =>
                            updateItem(i, "itemName", e.target.value)
                          }
                          placeholder="Item description"
                          className="w-full bg-transparent focus:outline-none text-gray-900 placeholder:text-gray-400"
                        />
                      </td>
                      <td className="px-3 py-1.5">
                        <input
                          type="text"
                          value={item.itemQty}
                          onChange={(e) =>
                            updateItem(i, "itemQty", e.target.value)
                          }
                          placeholder="1"
                          className="w-full bg-transparent focus:outline-none text-gray-900 placeholder:text-gray-400"
                        />
                      </td>
                      <td className="px-3 py-1.5">
                        <input
                          type="text"
                          value={item.unitPrice}
                          onChange={(e) =>
                            updateItem(i, "unitPrice", e.target.value)
                          }
                          placeholder="0.00"
                          className="w-full bg-transparent focus:outline-none text-gray-900 placeholder:text-gray-400"
                        />
                      </td>
                      <td className="px-3 py-1.5">
                        <input
                          type="text"
                          value={item.subTotal}
                          onChange={(e) =>
                            updateItem(i, "subTotal", e.target.value)
                          }
                          placeholder="0.00"
                          className="w-full bg-transparent focus:outline-none text-gray-900 placeholder:text-gray-400"
                        />
                      </td>
                      <td className="px-2 py-1.5 text-center">
                        <button
                          type="button"
                          onClick={() => removeItem(i)}
                          className="text-gray-400 hover:text-red-500"
                          title="Remove item"
                        >
                          <svg
                            viewBox="0 0 20 20"
                            fill="currentColor"
                            className="size-4"
                          >
                            <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
                          </svg>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
