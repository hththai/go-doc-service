import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import UploadFile from "./UploadFile";

// Mock the upload API
vi.mock("@/api/upload", () => ({
  uploadFile: vi.fn(),
}));

import { uploadFile } from "@/api/upload";

describe("UploadFile", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders all form fields from config", () => {
    render(<UploadFile />);

    expect(screen.getByLabelText(/title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByText(/upload a file/i)).toBeInTheDocument();
  });

  it("renders purchase info fields", () => {
    render(<UploadFile />);

    expect(screen.getByLabelText(/purchase price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase from/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase date/i)).toBeInTheDocument();
  });

  it("shows required indicator for required fields", () => {
    render(<UploadFile />);

    // Title has required indicator (*)
    const titleLabel = screen.getByText(/title/i).closest("label");
    expect(titleLabel?.textContent).toContain("*");
  });

  it("updates field values when typing", () => {
    render(<UploadFile />);

    const titleInput = screen.getByLabelText(/title/i);
    const descriptionInput = screen.getByLabelText(/description/i);

    fireEvent.change(titleInput, { target: { value: "Test Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });

    expect(titleInput).toHaveValue("Test Title");
    expect(descriptionInput).toHaveValue("Test Description");
  });

  it("clears form when clear button is clicked", () => {
    render(<UploadFile />);

    const titleInput = screen.getByLabelText(/title/i);
    const descriptionInput = screen.getByLabelText(/description/i);
    const buyFromInput = screen.getByLabelText(/purchase from/i);
    const buyPriceInput = screen.getByLabelText(/purchase price/i);

    // Fill in fields
    fireEvent.change(titleInput, { target: { value: "Test Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });
    fireEvent.change(buyFromInput, { target: { value: "Woolworths" } });
    fireEvent.change(buyPriceInput, { target: { value: "12.50" } });

    // Click clear
    fireEvent.click(screen.getByRole("button", { name: /clear/i }));

    // All fields should be empty
    expect(titleInput).toHaveValue("");
    expect(descriptionInput).toHaveValue("");
    expect(buyFromInput).toHaveValue("");
    expect(buyPriceInput).toHaveValue(null);
  });

  it("updates purchase info fields when typing", () => {
    render(<UploadFile />);

    const buyFromInput = screen.getByLabelText(/purchase from/i);
    const buyPriceInput = screen.getByLabelText(/purchase price/i);
    const buyAtInput = screen.getByLabelText(/purchase date/i);

    fireEvent.change(buyFromInput, { target: { value: "Coles" } });
    fireEvent.change(buyPriceInput, { target: { value: "9.99" } });
    fireEvent.change(buyAtInput, { target: { value: "2026-02-22" } });

    expect(buyFromInput).toHaveValue("Coles");
    expect(buyPriceInput).toHaveValue(9.99);
    expect(buyAtInput).toHaveValue("2026-02-22");
  });

  it("calls uploadFile with metadata on submit", async () => {
    vi.mocked(uploadFile).mockResolvedValue({ success: true });

    render(<UploadFile />);

    fireEvent.change(screen.getByLabelText(/title/i), {
      target: { value: "My Title" },
    });
    fireEvent.change(screen.getByLabelText(/description/i), {
      target: { value: "My Description" },
    });
    fireEvent.change(screen.getByLabelText(/purchase from/i), {
      target: { value: "Woolworths" },
    });
    fireEvent.change(screen.getByLabelText(/purchase price/i), {
      target: { value: "12.50" },
    });
    fireEvent.change(screen.getByLabelText(/purchase date/i), {
      target: { value: "2026-02-22" },
    });

    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    expect(uploadFile).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "My Title",
        description: "My Description",
        buyFrom: "Woolworths",
        buyPrice: "12.50",
        buyAt: "2026-02-22",
      }),
      null,
    );
  });

  it("shows success message after successful upload", async () => {
    vi.mocked(uploadFile).mockResolvedValue({ success: true });

    render(<UploadFile />);

    fireEvent.change(screen.getByLabelText(/title/i), {
      target: { value: "Test" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    // Wait for success message
    expect(
      await screen.findByText(/uploaded successfully/i),
    ).toBeInTheDocument();
  });
});
