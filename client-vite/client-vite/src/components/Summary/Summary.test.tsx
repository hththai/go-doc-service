import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import Summary from "./Summary";
import type { Purchase } from "./mockData";

vi.mock("@/api/upload", () => ({
  getPurchases: vi.fn(),
}));

import { getPurchases } from "@/api/upload";

const mockPurchases: Purchase[] = [
  {
    id: "1",
    title: "Weekly Groceries",
    filename: "woolworths.pdf",
    fileUrl: "http://localhost:8080/v1/auth/file/1",
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
    ],
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
];

function renderSummary() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <Summary />
    </QueryClientProvider>,
  );
}

describe("Summary", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows loading state while fetching", () => {
    vi.mocked(getPurchases).mockReturnValue(new Promise(() => {})); // never resolves
    renderSummary();
    expect(screen.getByText(/loading purchases/i)).toBeInTheDocument();
  });

  it("shows error state when fetch fails", async () => {
    vi.mocked(getPurchases).mockRejectedValue(new Error("Network error"));
    renderSummary();
    await waitFor(() => {
      expect(screen.getByText(/failed to load purchases/i)).toBeInTheDocument();
    });
  });

  it("renders a row for each purchase", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    await waitFor(() => {
      expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
      expect(screen.getByText("Office Supplies")).toBeInTheDocument();
    });
  });

  it("shows the grand total across all purchases", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    // 87.45 + 134.90 = 222.35
    await waitFor(() => {
      expect(screen.getAllByText("$222.35").length).toBeGreaterThan(0);
    });
  });

  it("populates year filter options from the data", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "2025" })).toBeInTheDocument();
    });
  });

  it("populates month filter options from the data", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    await waitFor(() => {
      expect(
        screen.getByRole("option", { name: "January" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: "February" }),
      ).toBeInTheDocument();
    });
  });

  it("filters purchases by month", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.change(screen.getByRole("combobox", { name: /month/i }), {
      target: { value: "1" }, // January
    });

    expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
    expect(screen.queryByText("Office Supplies")).not.toBeInTheDocument();
  });

  it("resets month filter when year changes", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderSummary();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    // Select January
    fireEvent.change(screen.getByRole("combobox", { name: /month/i }), {
      target: { value: "1" },
    });
    // Change year — month should reset to "all"
    fireEvent.change(screen.getByRole("combobox", { name: /year/i }), {
      target: { value: "2025" },
    });

    expect(screen.getByRole("combobox", { name: /month/i })).toHaveValue("all");
  });

  it("filters purchases by year, hiding purchases from other years", async () => {
    const mixedYearPurchases: Purchase[] = [
      {
        id: "10",
        title: "Old Purchase",
        filename: "old.pdf",
        buyAt: "2024-06-01",
        buyFrom: "OldStore",
        buyPrice: "50.00",
        items: [],
      },
      mockPurchases[0], // 2025 purchase
    ];
    vi.mocked(getPurchases).mockResolvedValue(mixedYearPurchases);
    renderSummary();
    await waitFor(() => screen.getByText("Old Purchase"));

    // "2025" is a valid option since there is 2025 data
    fireEvent.change(screen.getByRole("combobox", { name: /year/i }), {
      target: { value: "2025" },
    });

    await waitFor(() => {
      expect(screen.queryByText("Old Purchase")).not.toBeInTheDocument();
    });
    expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
  });

  it("shows no purchases message when server returns empty list", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    renderSummary();
    await waitFor(() => {
      expect(screen.getByText(/no purchases found/i)).toBeInTheDocument();
    });
  });
});
