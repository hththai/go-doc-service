import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import EditTable from "./EditTable";
import type { Purchase } from "../Summary/mockData";

vi.mock("@/api/upload", () => ({
  getPurchases: vi.fn(),
  createPurchase: vi.fn(),
  updatePurchase: vi.fn(),
  deletePurchase: vi.fn(),
  uploadFile: vi.fn(),
}));

vi.mock("@/api/ocr", () => ({
  scanInvoice: vi.fn(),
}));

import {
  getPurchases,
  createPurchase,
  updatePurchase,
  deletePurchase,
  uploadFile,
} from "@/api/upload";
import { scanInvoice } from "@/api/ocr";

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

function renderEditTable() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <EditTable />
    </QueryClientProvider>,
  );
}

describe("EditTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows loading state while fetching", () => {
    vi.mocked(getPurchases).mockReturnValue(new Promise(() => {}));
    renderEditTable();
    expect(screen.getByText(/loading purchases/i)).toBeInTheDocument();
  });

  it("shows error state when fetch fails", async () => {
    vi.mocked(getPurchases).mockRejectedValue(new Error("Network error"));
    renderEditTable();
    await waitFor(() => {
      expect(screen.getByText(/failed to load purchases/i)).toBeInTheDocument();
    });
  });

  it("renders a row for each purchase", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => {
      expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
      expect(screen.getByText("Office Supplies")).toBeInTheDocument();
    });
  });

  it("shows the grand total across all purchases", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    // 87.45 + 134.90 = 222.35
    await waitFor(() => {
      expect(screen.getAllByText("$222.35").length).toBeGreaterThan(0);
    });
  });

  it("shows no purchases message when server returns empty list", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    renderEditTable();
    await waitFor(() => {
      expect(screen.getByText(/no purchases found/i)).toBeInTheDocument();
    });
  });

  it("populates year filter options from the data", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "2025" })).toBeInTheDocument();
    });
  });

  it("populates month filter options from the data", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
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
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.change(screen.getByRole("combobox", { name: /month/i }), {
      target: { value: "1" }, // January
    });

    expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
    expect(screen.queryByText("Office Supplies")).not.toBeInTheDocument();
  });

  it("resets month filter when year changes", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.change(screen.getByRole("combobox", { name: /month/i }), {
      target: { value: "1" },
    });
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
      mockPurchases[0],
    ];
    vi.mocked(getPurchases).mockResolvedValue(mixedYearPurchases);
    renderEditTable();
    await waitFor(() => screen.getByText("Old Purchase"));

    fireEvent.change(screen.getByRole("combobox", { name: /year/i }), {
      target: { value: "2025" },
    });

    await waitFor(() => {
      expect(screen.queryByText("Old Purchase")).not.toBeInTheDocument();
    });
    expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
  });

  it("opens the form modal when New button is clicked", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    renderEditTable();
    await waitFor(() => screen.getByText(/no purchases found/i));

    fireEvent.click(screen.getByRole("button", { name: /new purchase/i }));

    expect(screen.getByText("New Purchase")).toBeInTheDocument();
  });

  it("opens the edit form modal when edit is clicked", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.click(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    );

    expect(screen.getByText("Edit Purchase")).toBeInTheDocument();
  });

  it("pre-fills the edit form with the purchase title", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.click(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    );

    expect(screen.getByDisplayValue("Weekly Groceries")).toBeInTheDocument();
  });

  it("opens the delete confirm modal when delete is clicked", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.click(
      screen.getByRole("button", { name: /delete weekly groceries/i }),
    );

    expect(screen.getByText(/are you sure/i)).toBeInTheDocument();
  });

  it("closes the form modal when Cancel is clicked", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    renderEditTable();
    await waitFor(() => screen.getByText(/no purchases found/i));

    fireEvent.click(screen.getByRole("button", { name: /new purchase/i }));
    expect(screen.getByText("New Purchase")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /cancel/i }));
    expect(screen.queryByText("New Purchase")).not.toBeInTheDocument();
  });

  it("calls createPurchase when the form is submitted in create mode", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    vi.mocked(createPurchase).mockResolvedValue({ id: "new-1" });
    renderEditTable();
    await waitFor(() => screen.getByText(/no purchases found/i));

    fireEvent.click(screen.getByRole("button", { name: /new purchase/i }));

    fireEvent.change(screen.getByLabelText(/title/i), {
      target: { value: "Test Purchase" },
    });
    fireEvent.change(screen.getByLabelText(/store/i), {
      target: { value: "Test Store" },
    });
    fireEvent.change(screen.getByLabelText(/date/i), {
      target: { value: "2025-06-01" },
    });
    fireEvent.change(screen.getByLabelText(/total price/i), {
      target: { value: "99.99" },
    });

    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(createPurchase).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "Test Purchase",
          buyFrom: "Test Store",
          buyAt: "2025-06-01",
          buyPrice: "99.99",
        }),
      );
    });
  });

  it("calls updatePurchase when the form is submitted in edit mode", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    vi.mocked(updatePurchase).mockResolvedValue({ id: "1" });
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.click(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    );

    fireEvent.change(screen.getByDisplayValue("Weekly Groceries"), {
      target: { value: "Updated Title" },
    });

    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(updatePurchase).toHaveBeenCalledWith(
        "1",
        expect.objectContaining({ title: "Updated Title" }),
      );
    });
  });

  it("calls deletePurchase when delete is confirmed", async () => {
    vi.mocked(getPurchases).mockResolvedValue(mockPurchases);
    vi.mocked(deletePurchase).mockResolvedValue(undefined);
    renderEditTable();
    await waitFor(() => screen.getByText("Weekly Groceries"));

    fireEvent.click(
      screen.getByRole("button", { name: /delete weekly groceries/i }),
    );
    fireEvent.click(screen.getByRole("button", { name: /^delete$/i }));

    await waitFor(() => {
      expect(deletePurchase).toHaveBeenCalledWith("1");
    });
  });

  it("calls uploadFile instead of createPurchase when a file is attached", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    vi.mocked(uploadFile).mockResolvedValue({ id: "new-1" });
    renderEditTable();
    await waitFor(() => screen.getByText(/no purchases found/i));

    fireEvent.click(screen.getByRole("button", { name: /new purchase/i }));

    fireEvent.change(screen.getByLabelText(/title/i), {
      target: { value: "Test Purchase" },
    });
    fireEvent.change(screen.getByLabelText(/store/i), {
      target: { value: "Test Store" },
    });
    fireEvent.change(screen.getByLabelText(/date/i), {
      target: { value: "2025-06-01" },
    });
    fireEvent.change(screen.getByLabelText(/total price/i), {
      target: { value: "99.99" },
    });

    const file = new File(["content"], "receipt.jpg", { type: "image/jpeg" });
    const fileInput = document.querySelector(
      'input[type="file"]',
    ) as HTMLInputElement;
    fireEvent.change(fileInput, { target: { files: [file] } });

    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(uploadFile).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "Test Purchase",
          buyFrom: "Test Store",
          buyAt: "2025-06-01",
          buyPrice: "99.99",
        }),
        [],
        file,
      );
      expect(createPurchase).not.toHaveBeenCalled();
    });
  });

  it("pre-fills store, date and price fields after scanning a file", async () => {
    vi.mocked(getPurchases).mockResolvedValue([]);
    vi.mocked(scanInvoice).mockResolvedValue({
      invoice: {
        seller: "ScannedStore",
        abn: "",
        document_date: "",
        order_no: "",
        order_date: "",
        billing_address: "",
        delivery_address: "",
        items: [],
        shipping_charges: "",
        total: "55.00",
      },
      purchaseInfo: {
        buyAt: "01/06/2025",
        buyFrom: "ScannedStore",
        buyPrice: "55.00",
      },
    });
    renderEditTable();
    await waitFor(() => screen.getByText(/no purchases found/i));

    fireEvent.click(screen.getByRole("button", { name: /new purchase/i }));

    const file = new File(["content"], "receipt.jpg", { type: "image/jpeg" });
    const fileInput = document.querySelector(
      'input[type="file"]',
    ) as HTMLInputElement;
    fireEvent.change(fileInput, { target: { files: [file] } });

    fireEvent.click(screen.getByRole("button", { name: /scan/i }));

    await waitFor(() => {
      expect(screen.getByDisplayValue("ScannedStore")).toBeInTheDocument();
      expect(screen.getByDisplayValue("55.00")).toBeInTheDocument();
      expect(screen.getByDisplayValue("2025-06-01")).toBeInTheDocument();
    });
  });
});
