import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import EditTableRow from "./EditTableRow";
import type { Purchase } from "../Summary/mockData";

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
  onEdit = vi.fn(),
  onDelete = vi.fn(),
) {
  return render(
    <table>
      <tbody>
        <EditTableRow purchase={purchase} onEdit={onEdit} onDelete={onDelete} />
      </tbody>
    </table>,
  );
}

describe("EditTableRow", () => {
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

  it("renders the edit and delete action buttons", () => {
    renderRow();
    expect(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /delete weekly groceries/i }),
    ).toBeInTheDocument();
  });

  it("calls onEdit with the purchase when edit button is clicked", () => {
    const onEdit = vi.fn();
    renderRow(basePurchase, onEdit);

    fireEvent.click(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    );

    expect(onEdit).toHaveBeenCalledWith(basePurchase);
  });

  it("calls onDelete with the purchase when delete button is clicked", () => {
    const onDelete = vi.fn();
    renderRow(basePurchase, vi.fn(), onDelete);

    fireEvent.click(
      screen.getByRole("button", { name: /delete weekly groceries/i }),
    );

    expect(onDelete).toHaveBeenCalledWith(basePurchase);
  });

  it("does not call onDelete when edit button is clicked", () => {
    const onDelete = vi.fn();
    renderRow(basePurchase, vi.fn(), onDelete);

    fireEvent.click(
      screen.getByRole("button", { name: /edit weekly groceries/i }),
    );

    expect(onDelete).not.toHaveBeenCalled();
  });

  it("does not call onEdit when delete button is clicked", () => {
    const onEdit = vi.fn();
    renderRow(basePurchase, onEdit);

    fireEvent.click(
      screen.getByRole("button", { name: /delete weekly groceries/i }),
    );

    expect(onEdit).not.toHaveBeenCalled();
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
