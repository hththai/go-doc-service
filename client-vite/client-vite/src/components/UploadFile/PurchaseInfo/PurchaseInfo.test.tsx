import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import PurchaseInfo from "./PurchaseInfo";

vi.mock("@/api/categories", () => ({
  getCategories: vi.fn().mockResolvedValue([]),
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  deleteCategory: vi.fn(),
}));

const defaultValues = { buyAt: "", buyFrom: "", buyPrice: "" };

function renderPurchaseInfo(
  props: Partial<Parameters<typeof PurchaseInfo>[0]> = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <PurchaseInfo
        values={defaultValues}
        items={[]}
        categoryGuids={[]}
        onChange={vi.fn()}
        onItemsChange={vi.fn()}
        onCategoryGuidsChange={vi.fn()}
        {...props}
      />
    </QueryClientProvider>,
  );
}

describe("PurchaseInfo", () => {
  it("renders all three purchase fields", () => {
    renderPurchaseInfo();

    expect(screen.getByLabelText(/purchase price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase from/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase date/i)).toBeInTheDocument();
  });

  it("displays values passed via props", () => {
    renderPurchaseInfo({
      values: {
        buyAt: "2026-02-22",
        buyFrom: "Woolworths",
        buyPrice: "12.50",
      },
    });

    expect(screen.getByLabelText(/purchase price/i)).toHaveValue(12.5);
    expect(screen.getByLabelText(/purchase from/i)).toHaveValue("Woolworths");
    expect(screen.getByLabelText(/purchase date/i)).toHaveValue("2026-02-22");
  });

  it("calls onChange with 'buyPrice' when purchase price changes", () => {
    const onChange = vi.fn();
    renderPurchaseInfo({ onChange });

    fireEvent.change(screen.getByLabelText(/purchase price/i), {
      target: { value: "25.99" },
    });

    expect(onChange).toHaveBeenCalledWith("buyPrice", "25.99");
  });

  it("calls onChange with 'buyFrom' when purchase from changes", () => {
    const onChange = vi.fn();
    renderPurchaseInfo({ onChange });

    fireEvent.change(screen.getByLabelText(/purchase from/i), {
      target: { value: "Coles" },
    });

    expect(onChange).toHaveBeenCalledWith("buyFrom", "Coles");
  });

  it("calls onChange with 'buyAt' when purchase date changes", () => {
    const onChange = vi.fn();
    renderPurchaseInfo({ onChange });

    fireEvent.change(screen.getByLabelText(/purchase date/i), {
      target: { value: "2026-02-22" },
    });

    expect(onChange).toHaveBeenCalledWith("buyAt", "2026-02-22");
  });
});
