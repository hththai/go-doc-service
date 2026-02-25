import { useState, useMemo, useRef, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { MONTH_NAMES, formatCurrency, type Purchase } from "./mockData";
import { getPurchases } from "../../api/upload";
import PurchaseRow from "./PurchaseRow";
import PurchaseDetail from "./PurchaseDetail";
import FilePreviewModal from "./FilePreviewModal";

const selectClass =
  "rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-xs focus:outline-none focus:ring-2 focus:ring-indigo-500";

export default function Summary() {
  const [selectedYear, setSelectedYear] = useState<string>("all");
  const [selectedMonth, setSelectedMonth] = useState<string>("all");
  const [activePurchase, setActivePurchase] = useState<Purchase | null>(null);
  const [previewFile, setPreviewFile] = useState<{
    url: string;
    filename: string;
  } | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const {
    data: purchases = [],
    isLoading,
    isError,
  } = useQuery<Purchase[]>({
    queryKey: ["purchases"],
    queryFn: getPurchases,
  });

  const availableYears = useMemo(() => {
    const years = new Set(purchases.map((p) => p.buyAt.slice(0, 4)));
    return Array.from(years).sort((a, b) => b.localeCompare(a));
  }, [purchases]);

  function handleYearChange(year: string) {
    setSelectedYear(year);
    setSelectedMonth("all");
  }

  const availableMonths = useMemo(() => {
    const source =
      selectedYear === "all"
        ? purchases
        : purchases.filter((p) => p.buyAt.startsWith(selectedYear));
    const months = new Set(
      source.map((p) => Number.parseInt(p.buyAt.slice(5, 7), 10)),
    );
    return Array.from(months).sort((a, b) => a - b);
  }, [selectedYear, purchases]);

  const filtered = useMemo(() => {
    return purchases.filter((p) => {
      const yearMatch =
        selectedYear === "all" || p.buyAt.startsWith(selectedYear);
      const monthMatch =
        selectedMonth === "all" ||
        Number.parseInt(p.buyAt.slice(5, 7), 10) ===
          Number.parseInt(selectedMonth, 10);
      return yearMatch && monthMatch;
    });
  }, [selectedYear, selectedMonth, purchases]);

  const total = filtered.reduce((sum, p) => {
    const price = Number.parseFloat(p.buyPrice);
    return sum + (Number.isNaN(price) ? 0 : price);
  }, 0);

  // 640px matches Tailwind's `sm` breakpoint
  function openDetail(purchase: Purchase) {
    if (window.innerWidth < 640) setActivePurchase(purchase);
  }

  function closeDetail() {
    setActivePurchase(null);
  }

  function openPreview(url: string, filename: string) {
    setActivePurchase(null); // ensure detail modal is closed first
    setPreviewFile({ url, filename });
  }

  function closePreview() {
    setPreviewFile(null);
  }

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    if (activePurchase) {
      if (!el.open) el.showModal();
    } else if (el.open) {
      el.close();
    }
  }, [activePurchase]);

  // Close on backdrop click — attached imperatively to avoid linter warnings on <dialog>
  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    function onBackdrop(e: MouseEvent) {
      if (e.target === el) setActivePurchase(null);
    }
    el.addEventListener("mousedown", onBackdrop);
    return () => el.removeEventListener("mousedown", onBackdrop);
  }, []);

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
          {isLoading && (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              Loading purchases…
            </p>
          )}
          {isError && (
            <p className="px-4 py-8 text-center text-sm text-red-400">
              Failed to load purchases. Please try again.
            </p>
          )}
          {!isLoading && !isError && filtered.length === 0 && (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              No purchases found for the selected period.
            </p>
          )}
          {!isLoading && !isError && filtered.length > 0 && (
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
                  <th className="sm:hidden w-8" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((purchase) => (
                  <PurchaseRow
                    key={purchase.id}
                    purchase={purchase}
                    onSelect={openDetail}
                    onPreview={openPreview}
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

      {/* Detail modal — mobile only */}
      <dialog
        ref={dialogRef}
        className="w-full max-w-sm rounded-xl shadow-2xl p-0 backdrop:bg-black/40 mt-16 mb-auto mx-auto"
      >
        {activePurchase && (
          <PurchaseDetail
            purchase={activePurchase}
            onClose={closeDetail}
            onPreview={openPreview}
          />
        )}
      </dialog>

      {/* File preview modal */}
      {previewFile && (
        <FilePreviewModal
          url={previewFile.url}
          filename={previewFile.filename}
          onClose={closePreview}
        />
      )}
    </div>
  );
}
