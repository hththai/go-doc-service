import { useState, useMemo, useRef, useEffect } from "react";
import { ChevronRight, X } from "lucide-react";

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
  // 2024 – March
  {
    id: "2024-03-1",
    title: "Monthly Groceries",
    filename: "coles_mar2024.pdf",
    buyAt: "2024-03-05",
    buyFrom: "Coles",
    buyPrice: "112.30",
    items: [
      {
        itemName: "Pasta 500g",
        itemQty: "3",
        unitPrice: "2.50",
        subTotal: "7.50",
      },
      {
        itemName: "Tomato Sauce",
        itemQty: "2",
        unitPrice: "3.80",
        subTotal: "7.60",
      },
      {
        itemName: "Beef Mince 500g",
        itemQty: "4",
        unitPrice: "9.00",
        subTotal: "36.00",
      },
    ],
  },
  {
    id: "2024-03-2",
    title: "Gym Membership",
    filename: "gym_invoice_mar24.pdf",
    buyAt: "2024-03-01",
    buyFrom: "Fitness First",
    buyPrice: "79.99",
    items: [
      {
        itemName: "Monthly Membership",
        itemQty: "1",
        unitPrice: "79.99",
        subTotal: "79.99",
      },
    ],
  },

  // 2024 – June
  {
    id: "2024-06-1",
    title: "Mid-Year Stationery",
    filename: "officeworks_jun24.jpg",
    buyAt: "2024-06-12",
    buyFrom: "Officeworks",
    buyPrice: "68.50",
    items: [
      {
        itemName: "Notebook A5 (5pk)",
        itemQty: "1",
        unitPrice: "18.00",
        subTotal: "18.00",
      },
      {
        itemName: "Pens (10pk)",
        itemQty: "2",
        unitPrice: "8.00",
        subTotal: "16.00",
      },
      {
        itemName: "Highlighters",
        itemQty: "1",
        unitPrice: "6.50",
        subTotal: "6.50",
      },
    ],
  },
  {
    id: "2024-06-2",
    title: "Team Building Dinner",
    filename: "italian_bistro_jun24.jpg",
    buyAt: "2024-06-28",
    buyFrom: "Italian Bistro",
    buyPrice: "345.00",
    items: [
      {
        itemName: "Pasta Carbonara (x5)",
        itemQty: "5",
        unitPrice: "28.00",
        subTotal: "140.00",
      },
      {
        itemName: "Wood-fired Pizza (x3)",
        itemQty: "3",
        unitPrice: "32.00",
        subTotal: "96.00",
      },
      {
        itemName: "Wine Bottle",
        itemQty: "2",
        unitPrice: "45.00",
        subTotal: "90.00",
      },
    ],
  },

  // 2024 – September
  {
    id: "2024-09-1",
    title: "Home Office Upgrade",
    filename: "jbhifi_sep24.pdf",
    buyAt: "2024-09-10",
    buyFrom: "JB Hi-Fi",
    buyPrice: "499.00",
    items: [
      {
        itemName: '27" Monitor',
        itemQty: "1",
        unitPrice: "399.00",
        subTotal: "399.00",
      },
      {
        itemName: "Monitor Stand",
        itemQty: "1",
        unitPrice: "59.00",
        subTotal: "59.00",
      },
      {
        itemName: "HDMI Cable 2m",
        itemQty: "1",
        unitPrice: "29.00",
        subTotal: "29.00",
      },
    ],
  },
  {
    id: "2024-09-2",
    title: "Spring Groceries",
    filename: "woolworths_sep24.pdf",
    buyAt: "2024-09-22",
    buyFrom: "Woolworths",
    buyPrice: "93.75",
    items: [
      {
        itemName: "Salmon Fillet 400g",
        itemQty: "2",
        unitPrice: "14.00",
        subTotal: "28.00",
      },
      {
        itemName: "Salad Mix",
        itemQty: "3",
        unitPrice: "4.50",
        subTotal: "13.50",
      },
      {
        itemName: "Orange Juice 2L",
        itemQty: "2",
        unitPrice: "5.50",
        subTotal: "11.00",
      },
    ],
  },

  // 2024 – December
  {
    id: "2024-12-1",
    title: "Christmas Party",
    filename: "catering_dec24.pdf",
    buyAt: "2024-12-20",
    buyFrom: "Fresh Catering Co.",
    buyPrice: "620.00",
    items: [
      {
        itemName: "Catering Platter (x4)",
        itemQty: "4",
        unitPrice: "85.00",
        subTotal: "340.00",
      },
      {
        itemName: "Soft Drinks Crate",
        itemQty: "3",
        unitPrice: "40.00",
        subTotal: "120.00",
      },
      {
        itemName: "Dessert Assortment",
        itemQty: "2",
        unitPrice: "60.00",
        subTotal: "120.00",
      },
    ],
  },
  {
    id: "2024-12-2",
    title: "Year-End Supplies",
    filename: "bunnings_dec24.jpg",
    buyAt: "2024-12-05",
    buyFrom: "Bunnings",
    buyPrice: "157.40",
    items: [
      {
        itemName: "Storage Boxes (6pk)",
        itemQty: "2",
        unitPrice: "39.00",
        subTotal: "78.00",
      },
      {
        itemName: "Label Maker",
        itemQty: "1",
        unitPrice: "49.90",
        subTotal: "49.90",
      },
    ],
  },

  // 2025 – January
  {
    id: "2025-01-1",
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
    id: "2025-01-2",
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

  // 2025 – February
  {
    id: "2025-02-1",
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
    id: "2025-02-2",
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

const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-AU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function formatCurrency(amount: string | number) {
  return `$${Number.parseFloat(amount.toString()).toFixed(2)}`;
}

const selectClass =
  "rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-xs focus:outline-none focus:ring-2 focus:ring-indigo-500";

export default function Summary() {
  const [selectedYear, setSelectedYear] = useState<string>("all");
  const [selectedMonth, setSelectedMonth] = useState<string>("all");
  const [activePurchase, setActivePurchase] = useState<Purchase | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const availableYears = useMemo(() => {
    const years = new Set(MOCK_PURCHASES.map((p) => p.buyAt.slice(0, 4)));
    return Array.from(years).sort((a, b) => b.localeCompare(a));
  }, []);

  function handleYearChange(year: string) {
    setSelectedYear(year);
    setSelectedMonth("all");
  }

  const availableMonths = useMemo(() => {
    const source =
      selectedYear === "all"
        ? MOCK_PURCHASES
        : MOCK_PURCHASES.filter((p) => p.buyAt.startsWith(selectedYear));
    const months = new Set(
      source.map((p) => Number.parseInt(p.buyAt.slice(5, 7), 10)),
    );
    return Array.from(months).sort((a, b) => a - b);
  }, [selectedYear]);

  const filtered = useMemo(() => {
    return MOCK_PURCHASES.filter((p) => {
      const yearMatch =
        selectedYear === "all" || p.buyAt.startsWith(selectedYear);
      const monthMatch =
        selectedMonth === "all" ||
        Number.parseInt(p.buyAt.slice(5, 7), 10) ===
          Number.parseInt(selectedMonth, 10);
      return yearMatch && monthMatch;
    });
  }, [selectedYear, selectedMonth]);

  const total = filtered.reduce(
    (sum, p) => sum + Number.parseFloat(p.buyPrice),
    0,
  );

  // 640px matches Tailwind's `sm` breakpoint
  function openDetail(purchase: Purchase) {
    if (window.innerWidth < 640) setActivePurchase(purchase);
  }

  function closeDetail() {
    setActivePurchase(null);
  }

  // Sync dialog open/close with activePurchase state
  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    if (activePurchase) {
      if (!el.open) el.showModal();
    } else if (el.open) {
      el.close();
    }
  }, [activePurchase]);

  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Purchase Summary
      </div>

      <div className="border border-gray-200 rounded-b-md overflow-hidden">
        {/* Filters + stats bar */}
        <div className="bg-gray-50 px-4 py-3 flex flex-wrap items-center gap-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <label htmlFor="year-filter" className="text-sm text-gray-500">
              Year
            </label>
            <select
              id="year-filter"
              value={selectedYear}
              onChange={(e) => handleYearChange(e.target.value)}
              className={selectClass}
            >
              <option value="all">All</option>
              {availableYears.map((y) => (
                <option key={y} value={y}>
                  {y}
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-2">
            <label htmlFor="month-filter" className="text-sm text-gray-500">
              Month
            </label>
            <select
              id="month-filter"
              value={selectedMonth}
              onChange={(e) => setSelectedMonth(e.target.value)}
              className={selectClass}
            >
              <option value="all">All</option>
              {availableMonths.map((m) => (
                <option key={m} value={m}>
                  {MONTH_NAMES[m - 1]}
                </option>
              ))}
            </select>
          </div>

          <div className="ml-auto flex gap-6 text-sm text-gray-600">
            <span>
              <span className="font-medium text-gray-900">
                {filtered.length}
              </span>{" "}
              purchases
            </span>
            <span>
              Total:{" "}
              <span className="font-medium text-gray-900">
                {formatCurrency(total)}
              </span>
            </span>
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          {filtered.length === 0 ? (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              No purchases found for the selected period.
            </p>
          ) : (
            <table className="min-w-full text-sm">
              <thead className="bg-gray-50 text-xs text-gray-500 uppercase border-b border-gray-200">
                <tr>
                  <th className="px-4 py-2 text-left font-medium">Title</th>
                  <th className="hidden sm:table-cell px-4 py-2 text-left font-medium">
                    File
                  </th>
                  <th className="hidden sm:table-cell px-4 py-2 text-left font-medium w-32">
                    Date
                  </th>
                  <th className="hidden sm:table-cell px-4 py-2 text-left font-medium w-40">
                    Store
                  </th>
                  <th className="px-4 py-2 text-right font-medium w-28">
                    Total
                  </th>
                  <th className="hidden sm:table-cell px-4 py-2 text-center font-medium w-20">
                    Items
                  </th>
                  {/* Mobile-only tap indicator */}
                  <th className="sm:hidden w-8" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((purchase) => (
                  <PurchaseRow
                    key={purchase.id}
                    purchase={purchase}
                    onSelect={openDetail}
                  />
                ))}
              </tbody>
              <tfoot>
                <tr className="bg-gray-50 border-t border-gray-200">
                  <td
                    colSpan={2}
                    className="px-4 py-2 text-sm font-semibold text-gray-700"
                  >
                    Grand Total
                  </td>
                  <td className="hidden sm:table-cell" colSpan={2} />
                  <td className="px-4 py-2 text-right text-sm font-semibold text-gray-900">
                    {formatCurrency(total)}
                  </td>
                  <td className="hidden sm:table-cell" />
                  <td className="sm:hidden" />
                </tr>
              </tfoot>
            </table>
          )}
        </div>
      </div>

      {/* Detail modal (mobile) */}
      <dialog
        ref={dialogRef}
        onMouseDown={(e) => {
          if (e.target === dialogRef.current) closeDetail();
        }}
        className="w-full max-w-sm rounded-xl shadow-2xl p-0 backdrop:bg-black/40 mt-16 mb-auto mx-auto"
      >
        {activePurchase && (
          <PurchaseDetail purchase={activePurchase} onClose={closeDetail} />
        )}
      </dialog>
    </div>
  );
}

function PurchaseRow({
  purchase,
  onSelect,
}: Readonly<{ purchase: Purchase; onSelect: (p: Purchase) => void }>) {
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
        <td className="hidden sm:table-cell px-4 py-2.5 text-gray-500 text-xs truncate max-w-36">
          {purchase.filename}
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

function PurchaseDetail({
  purchase,
  onClose,
}: Readonly<{ purchase: Purchase; onClose: () => void }>) {
  const itemTotal = purchase.items.reduce(
    (sum, item) => sum + Number.parseFloat(item.subTotal),
    0,
  );

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
          <span className="truncate max-w-48 text-right">
            {purchase.filename}
          </span>
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
