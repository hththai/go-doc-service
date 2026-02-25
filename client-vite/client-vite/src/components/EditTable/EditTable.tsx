import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import {
  MONTH_NAMES,
  formatCurrency,
  type Purchase,
} from "../Summary/mockData";
import {
  getPurchases,
  createPurchase,
  updatePurchase,
  deletePurchase,
} from "../../api/upload";
import EditTableRow from "./EditTableRow";
import PurchaseFormModal, { type PurchaseFormData } from "./PurchaseFormModal";
import DeleteConfirmModal from "./DeleteConfirmModal";

const selectClass =
  "rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-xs focus:outline-none focus:ring-2 focus:ring-indigo-500";

type FormState = { mode: "create" } | { mode: "edit"; purchase: Purchase };

export default function EditTable() {
  const [selectedYear, setSelectedYear] = useState<string>("all");
  const [selectedMonth, setSelectedMonth] = useState<string>("all");
  const [formState, setFormState] = useState<FormState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Purchase | null>(null);

  const queryClient = useQueryClient();

  const {
    data: purchases = [],
    isLoading,
    isError,
  } = useQuery<Purchase[]>({
    queryKey: ["purchases"],
    queryFn: getPurchases,
  });

  const createMutation = useMutation({
    mutationFn: (data: PurchaseFormData) => createPurchase(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["purchases"] });
      setFormState(null);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: PurchaseFormData }) =>
      updatePurchase(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["purchases"] });
      setFormState(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deletePurchase(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["purchases"] });
      setDeleteTarget(null);
    },
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

  function handleSave(data: PurchaseFormData) {
    if (formState?.mode === "edit") {
      updateMutation.mutate({ id: formState.purchase.id, data });
    } else {
      createMutation.mutate(data);
    }
  }

  const isSaving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md flex items-center justify-between">
        <span>Manage Purchases</span>
        <button
          type="button"
          onClick={() => setFormState({ mode: "create" })}
          className="flex items-center gap-1 text-sm text-white/90 hover:text-white transition-colors"
          aria-label="New purchase"
        >
          <Plus size={16} />
          New
        </button>
      </div>

      <div className="border border-gray-200 rounded-b-md overflow-hidden">
        {/* Filters + stats bar */}
        <div className="bg-gray-50 px-4 py-3 flex flex-wrap items-center gap-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <label htmlFor="edit-year-filter" className="text-sm text-gray-500">
              Year
            </label>
            <select
              id="edit-year-filter"
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
            <label
              htmlFor="edit-month-filter"
              className="text-sm text-gray-500"
            >
              Month
            </label>
            <select
              id="edit-month-filter"
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
                  <th className="px-4 py-2 text-right font-medium w-24">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((purchase) => (
                  <EditTableRow
                    key={purchase.id}
                    purchase={purchase}
                    onEdit={(p) => setFormState({ mode: "edit", purchase: p })}
                    onDelete={setDeleteTarget}
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
                  <td className="hidden sm:table-cell" />
                  <td className="px-4 py-2 text-right text-sm font-semibold text-gray-900">
                    {formatCurrency(total)}
                  </td>
                  <td className="hidden sm:table-cell" />
                  <td />
                </tr>
              </tfoot>
            </table>
          )}
        </div>
      </div>

      {/* Create / Edit form modal */}
      {formState && (
        <PurchaseFormModal
          mode={formState.mode}
          purchase={formState.mode === "edit" ? formState.purchase : undefined}
          onSave={handleSave}
          onClose={() => setFormState(null)}
          isSaving={isSaving}
        />
      )}

      {/* Delete confirm modal */}
      {deleteTarget && (
        <DeleteConfirmModal
          purchase={deleteTarget}
          onConfirm={() => deleteMutation.mutate(deleteTarget.id)}
          onCancel={() => setDeleteTarget(null)}
          isDeleting={deleteMutation.isPending}
        />
      )}
    </div>
  );
}
