import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import PurchaseRow from "./PurchaseRow";
import type { Purchase } from "./mockData";

const basePurchase: Purchase = {
  id: "1",
  title: "Weekly Groceries",
  filename: "receipt.pdf",
  fileUrl: "http://localhost:8080/v1/auth/file/1",
  buyAt: "2025-01-15",
  buyFrom: "Woolworths",
  buyPrice: "87.45",
  items: [
    { itemName: "Milk 2L", itemQty: "2", unitPrice: "3.50", subTotal: "7.00" },
    {
      itemName: "Chicken Breast",
      itemQty: "1",
      unitPrice: "8.50",
      subTotal: "8.50",
    },
  ],
};

function renderRow(
  purchase: Purchase = basePurchase,
  onSelect = vi.fn(),
  onPreview = vi.fn(),
) {
  return render(
    <table>
      <tbody>
        <PurchaseRow
          purchase={purchase}
          onSelect={onSelect}
          onPreview={onPreview}
        />
      </tbody>
    </table>,
  );
}

describe("PurchaseRow", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the purchase title", () => {
    renderRow();
    expect(screen.getByText("Weekly Groceries")).toBeInTheDocument();
  });

  it("renders the store name", () => {
    renderRow();
    expect(screen.getByText("Woolworths")).toBeInTheDocument();
  });

  it("renders the formatted price", () => {
    renderRow();
    expect(screen.getByText("$87.45")).toBeInTheDocument();
  });

  it("renders the item count", () => {
    // Use quantities that differ from the item count (2) to avoid ambiguous matches
    const purchase: Purchase = {
      ...basePurchase,
      items: [
        {
          itemName: "Item A",
          itemQty: "3",
          unitPrice: "5.00",
          subTotal: "15.00",
        },
        {
          itemName: "Item B",
          itemQty: "5",
          unitPrice: "2.00",
          subTotal: "10.00",
        },
      ],
    };
    renderRow(purchase);
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("renders filename as a clickable button when fileUrl is present", () => {
    renderRow();
    expect(
      screen.getByRole("button", { name: "receipt.pdf" }),
    ).toBeInTheDocument();
  });

  it("renders filename as plain text when fileUrl is absent", () => {
    const purchase: Purchase = { ...basePurchase, fileUrl: undefined };
    renderRow(purchase);
    expect(screen.getByText("receipt.pdf")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "receipt.pdf" }),
    ).not.toBeInTheDocument();
  });

  it("calls onPreview with url and filename when file button is clicked", () => {
    const onPreview = vi.fn();
    renderRow(basePurchase, vi.fn(), onPreview);

    fireEvent.click(screen.getByRole("button", { name: "receipt.pdf" }));

    expect(onPreview).toHaveBeenCalledWith(
      basePurchase.fileUrl,
      basePurchase.filename,
    );
  });

  it("does not call onSelect when the file button is clicked", () => {
    const onSelect = vi.fn();
    renderRow(basePurchase, onSelect);

    fireEvent.click(screen.getByRole("button", { name: "receipt.pdf" }));

    expect(onSelect).not.toHaveBeenCalled();
  });

  it("calls onSelect with the purchase when the row is clicked", () => {
    const onSelect = vi.fn();
    renderRow(basePurchase, onSelect);

    fireEvent.click(screen.getByText("Weekly Groceries"));

    expect(onSelect).toHaveBeenCalledWith(basePurchase);
  });

  it("renders the item sub-table when items exist", () => {
    renderRow();
    expect(screen.getByText("Milk 2L")).toBeInTheDocument();
    expect(screen.getByText("Chicken Breast")).toBeInTheDocument();
  });

  it("does not render the item sub-table when items list is empty", () => {
    const purchase: Purchase = { ...basePurchase, items: [] };
    renderRow(purchase);
    expect(screen.queryByText("Milk 2L")).not.toBeInTheDocument();
  });

  it("renders unit price and subtotal for each item", () => {
    renderRow();
    expect(screen.getByText("$3.50")).toBeInTheDocument();
    expect(screen.getByText("$7.00")).toBeInTheDocument();
  });
});
