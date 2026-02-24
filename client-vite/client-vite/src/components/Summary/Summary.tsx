type Item = {
  itemName: string;
  itemQty: string;
  unitPrice: string;
  subTotal: string;
};

type Purchase = {
  id: string;
  title: string;
  filename: string;
  buyAt: string;
  buyFrom: string;
  buyPrice: string;
  items: Item[];
};

const MOCK_PURCHASES: Purchase[] = [
  {
    id: "1",
    title: "Weekly Groceries",
    filename: "invoice_woolworths_jan.pdf",
    buyAt: "2025-01-15",
    buyFrom: "Woolworths",
    buyPrice: "87.45",
    items: [
      {
        itemName: "Milk 2L",
        itemQty: "2",
        unitPrice: "3.50",
        subTotal: "7.00",
      },
      { itemName: "Bread", itemQty: "1", unitPrice: "4.20", subTotal: "4.20" },
      {
        itemName: "Chicken Breast 500g",
        itemQty: "3",
        unitPrice: "8.50",
        subTotal: "25.50",
      },
      {
        itemName: "Eggs 12pk",
        itemQty: "2",
        unitPrice: "6.00",
        subTotal: "12.00",
      },
    ],
  },
  {
    id: "2",
    title: "Office Supplies",
    filename: "officeworks_receipt.jpg",
    buyAt: "2025-01-20",
    buyFrom: "Officeworks",
    buyPrice: "134.90",
    items: [
      {
        itemName: "A4 Paper Ream",
        itemQty: "2",
        unitPrice: "12.00",
        subTotal: "24.00",
      },
      {
        itemName: "Stapler",
        itemQty: "1",
        unitPrice: "18.90",
        subTotal: "18.90",
      },
      {
        itemName: "USB-C Hub",
        itemQty: "1",
        unitPrice: "69.00",
        subTotal: "69.00",
      },
    ],
  },
  {
    id: "3",
    title: "Team Lunch",
    filename: "thai_garden_receipt.jpg",
    buyAt: "2025-02-03",
    buyFrom: "Thai Garden Restaurant",
    buyPrice: "210.00",
    items: [
      {
        itemName: "Pad Thai (x4)",
        itemQty: "4",
        unitPrice: "22.00",
        subTotal: "88.00",
      },
      {
        itemName: "Green Curry (x3)",
        itemQty: "3",
        unitPrice: "24.00",
        subTotal: "72.00",
      },
      {
        itemName: "Drinks",
        itemQty: "7",
        unitPrice: "7.14",
        subTotal: "50.00",
      },
    ],
  },
  {
    id: "4",
    title: "Laptop Charger",
    filename: "amazon_invoice_feb.pdf",
    buyAt: "2025-02-10",
    buyFrom: "Amazon AU",
    buyPrice: "59.99",
    items: [
      {
        itemName: "USB-C 65W Charger",
        itemQty: "1",
        unitPrice: "59.99",
        subTotal: "59.99",
      },
    ],
  },
];

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-AU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function formatCurrency(amount: string) {
  return `$${parseFloat(amount).toFixed(2)}`;
}

const totalAll = MOCK_PURCHASES.reduce(
  (sum, p) => sum + parseFloat(p.buyPrice),
  0,
);

export default function Summary() {
  return (
    <div className="mt-6">
      {/* Header */}
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Purchase Summary
      </div>

      <div className="border border-gray-200 rounded-b-md overflow-hidden">
        {/* Stats bar */}
        <div className="bg-gray-50 px-4 py-3 flex flex-wrap gap-6 text-sm text-gray-600 border-b border-gray-200">
          <span>
            <span className="font-medium text-gray-900">
              {MOCK_PURCHASES.length}
            </span>{" "}
            purchases
          </span>
          <span>
            Total spent:{" "}
            <span className="font-medium text-gray-900">
              {formatCurrency(totalAll.toString())}
            </span>
          </span>
        </div>

        {/* Main table */}
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-50 text-xs text-gray-500 uppercase border-b border-gray-200">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Title</th>
                <th className="px-4 py-2 text-left font-medium">File</th>
                <th className="px-4 py-2 text-left font-medium w-32">Date</th>
                <th className="px-4 py-2 text-left font-medium w-40">Store</th>
                <th className="px-4 py-2 text-right font-medium w-28">Total</th>
                <th className="px-4 py-2 text-center font-medium w-20">
                  Items
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {MOCK_PURCHASES.map((purchase) => (
                <PurchaseRow key={purchase.id} purchase={purchase} />
              ))}
            </tbody>
            <tfoot>
              <tr className="bg-gray-50 border-t border-gray-200">
                <td
                  colSpan={4}
                  className="px-4 py-2 text-sm font-semibold text-gray-700"
                >
                  Grand Total
                </td>
                <td className="px-4 py-2 text-right text-sm font-semibold text-gray-900">
                  {formatCurrency(totalAll.toString())}
                </td>
                <td />
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>
  );
}

function PurchaseRow({ purchase }: Readonly<{ purchase: Purchase }>) {
  return (
    <>
      {/* Main row */}
      <tr className="hover:bg-gray-50 transition-colors">
        <td className="px-4 py-2.5 text-gray-900 font-medium">
          {purchase.title}
        </td>
        <td className="px-4 py-2.5 text-gray-500 text-xs truncate max-w-36">
          {purchase.filename}
        </td>
        <td className="px-4 py-2.5 text-gray-600">
          {formatDate(purchase.buyAt)}
        </td>
        <td className="px-4 py-2.5 text-gray-600">{purchase.buyFrom}</td>
        <td className="px-4 py-2.5 text-right text-gray-900 font-medium">
          {formatCurrency(purchase.buyPrice)}
        </td>
        <td className="px-4 py-2.5 text-center text-gray-500">
          {purchase.items.length}
        </td>
      </tr>

      {/* Line items sub-table */}
      {purchase.items.length > 0 && (
        <tr className="bg-gray-50/60">
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
                  {purchase.items.map((item, i) => (
                    <tr key={i} className="text-gray-600">
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
