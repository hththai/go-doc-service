import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import Report from "./Report";
import type { Purchase } from "../Summary/mockData";

// ResponsiveContainer requires real DOM dimensions which jsdom cannot provide.
// Render children directly so the rest of the chart tree is still exercised.
vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div>{children}</div>
    ),
  };
});

vi.mock("@/api/upload", () => ({
  getPurchases: vi.fn(),
}));

import { getPurchases } from "@/api/upload";

const mockPurchases: Purchase[] = [
  {
    id: "1",
    title: "Weekly Groceries",
    filename: "woolworths.pdf",
    buyAt: "2025-01-15",
    buyFrom: "Woolworths",
    buyPrice: "87.45",
    items: [],
  },
  {
    id: "2",
    title: "Office Supplies",
    filename: "officeworks.jpg",
    buyAt: "2025-02-20",
    buyFrom: "Officeworks",
    buyPrice: "134.90",
    items: [],
  },
  {
    id: "3",
    title: "Old Purchase",
    filename: "old.pdf",
    buyAt: "2024-06-01",
    buyFrom: "OldStore",
    buyPrice: "50.00",
    items: [],
  },
];

function renderReport() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <Report />
    </QueryClientProvider>,
  );
}

describe("Report", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows loading state while fetching", () => {
    vi.mocked(getPurchases).mockReturnValue(new Promise(() => {}));
    renderReport();
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it("shows error state when fetch fails", async () => {
    vi.mocked(getPurchases).mockRejectedValue(new Error("Network error"));
    renderReport();
    await waitFor(() => {
      expect(screen.getByText(/failed to load data/i)).toBeInTheDocument();
    });
  });

  it("shows empty state when there are no purchases", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    renderReport();
    await waitFor(() => {
      expect(screen.getByText(/no expense data/i)).toBeInTheDocument();
    });
  });

  it("shows the grand total across all purchases", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderReport();
    // 87.45 + 134.90 + 50.00 = 272.35
    await waitFor(() => {
      expect(screen.getByText("$272.35")).toBeInTheDocument();
    });
  });

  it("populates year filter options from the data", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderReport();
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "2025" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "2024" })).toBeInTheDocument();
    });
  });

  it("updates total when a year is selected", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderReport();
    await waitFor(() => screen.getByRole("option", { name: "2025" }));

    fireEvent.change(screen.getByRole("combobox", { name: /year/i }), {
      target: { value: "2025" },
    });

    // 87.45 + 134.90 = 222.35
    await waitFor(() => {
      expect(screen.getByText("$222.35")).toBeInTheDocument();
    });
  });

  it("shows empty state when selected year has no data", async () => {
    vi.mocked(getPurchases).mockResolvedValue([
      { ...mockPurchases[0], buyAt: "2024-01-10" },
    ]);
    renderReport();
    await waitFor(() => screen.getByRole("option", { name: "2024" }));

    fireEvent.change(screen.getByRole("combobox", { name: /year/i }), {
      target: { value: "2025" },
    });

    // No 2025 data — should not show the chart
    expect(screen.queryByText(/no expense data/i)).not.toBeInTheDocument();
  });
});
