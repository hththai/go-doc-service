import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { getPurchases } from "../../api/upload";
import {
  MONTH_NAMES,
  formatCurrency,
  type Purchase,
} from "../Summary/mockData";

const selectClass =
  "rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-xs focus:outline-none focus:ring-2 focus:ring-indigo-500";

type ChartDataPoint = {
  label: string;
  total: number;
};

function ChartTooltip({
  active,
  payload,
  label,
}: Readonly<{
  active?: boolean;
  payload?: { value: number }[];
  label?: string;
}>) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-md border border-gray-200 bg-white px-3 py-2 text-sm shadow-sm">
      <p className="text-gray-500 mb-1">{label}</p>
      <p className="font-medium text-gray-900">
        {formatCurrency(payload[0].value)}
      </p>
    </div>
  );
}

function buildChartData(
  purchases: Purchase[],
  selectedYear: string,
): ChartDataPoint[] {
  const totals = new Map<string, number>();

  const filtered =
    selectedYear === "all"
      ? purchases
      : purchases.filter((p) => p.buyAt.startsWith(selectedYear));

  for (const p of filtered) {
    const key =
      selectedYear === "all"
        ? p.buyAt.slice(0, 7) // "YYYY-MM"
        : MONTH_NAMES[Number.parseInt(p.buyAt.slice(5, 7), 10) - 1]; // "January"

    const price = Number.parseFloat(p.buyPrice);
    totals.set(key, (totals.get(key) ?? 0) + (Number.isNaN(price) ? 0 : price));
  }

  const entries = Array.from(totals.entries()).sort(([a], [b]) =>
    a.localeCompare(b),
  );

  return entries.map(([label, total]) => ({ label, total }));
}

export default function Report() {
  const [selectedYear, setSelectedYear] = useState<string>("all");

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

  const chartData = useMemo(
    () => buildChartData(purchases, selectedYear),
    [purchases, selectedYear],
  );

  const grandTotal = chartData.reduce((sum, d) => sum + d.total, 0);

  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Expense Report
      </div>

      <div className="border border-gray-200 rounded-b-md overflow-hidden">
        {/* Filter bar */}
        <div className="bg-gray-50 px-4 py-3 flex flex-wrap items-center gap-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <label
              htmlFor="report-year-filter"
              className="text-sm text-gray-500"
            >
              Year
            </label>
            <select
              id="report-year-filter"
              value={selectedYear}
              onChange={(e) => setSelectedYear(e.target.value)}
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

          <div className="ml-auto text-sm text-gray-600">
            Total:{" "}
            <span className="font-medium text-gray-900">
              {formatCurrency(grandTotal)}
            </span>
          </div>
        </div>

        {/* Chart area */}
        <div className="px-4 py-6">
          {isLoading && (
            <p className="py-8 text-center text-sm text-gray-400">Loading…</p>
          )}
          {isError && (
            <p className="py-8 text-center text-sm text-red-400">
              Failed to load data. Please try again.
            </p>
          )}
          {!isLoading && !isError && chartData.length === 0 && (
            <p className="py-8 text-center text-sm text-gray-400">
              No expense data for the selected period.
            </p>
          )}
          {!isLoading && !isError && chartData.length > 0 && (
            <ResponsiveContainer width="100%" height={320}>
              <LineChart
                data={chartData}
                margin={{ top: 8, right: 16, left: 8, bottom: 8 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis
                  dataKey="label"
                  tick={{ fontSize: 12, fill: "#6b7280" }}
                  tickLine={false}
                />
                <YAxis
                  tickFormatter={(v) =>
                    `$${(v as number).toLocaleString("en-AU")}`
                  }
                  tick={{ fontSize: 12, fill: "#6b7280" }}
                  tickLine={false}
                  axisLine={false}
                  width={72}
                />
                <Tooltip content={<ChartTooltip />} />
                <Line
                  type="monotone"
                  dataKey="total"
                  stroke="#0ea5e9"
                  strokeWidth={2}
                  dot={{ r: 4, fill: "#0ea5e9" }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  );
}
