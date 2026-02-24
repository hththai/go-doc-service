import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import UploadFile from "./UploadFile";

vi.mock("@/api/upload", () => ({
  uploadFile: vi.fn(),
}));

vi.mock("@/api/ocr", () => ({
  scanInvoice: vi.fn(),
  parseOcrDate: vi.fn(),
}));

import { uploadFile } from "@/api/upload";

const mockImageFile = () =>
  new File(["content"], "test.jpg", { type: "image/jpeg" });

const mockPdfFile = () =>
  new File(["content"], "invoice.pdf", { type: "application/pdf" });

const getFileInput = () =>
  document.querySelector('input[type="file"]') as HTMLInputElement;

describe("UploadFile", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders all form fields from config", () => {
    render(<UploadFile />);

    expect(screen.getByLabelText(/title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByText(/upload files/i)).toBeInTheDocument();
    expect(screen.getByText(/choose file/i)).toBeInTheDocument();
  });

  it("renders purchase info fields", () => {
    render(<UploadFile />);

    expect(screen.getByLabelText(/purchase price/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase from/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/purchase date/i)).toBeInTheDocument();
  });

  it("shows required indicator for required fields", () => {
    render(<UploadFile />);

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

    fireEvent.change(titleInput, { target: { value: "Test Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });
    fireEvent.change(buyFromInput, { target: { value: "Woolworths" } });
    fireEvent.change(buyPriceInput, { target: { value: "12.50" } });

    fireEvent.click(screen.getByRole("button", { name: /clear/i }));

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

    await waitFor(() => {
      expect(uploadFile).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "My Title",
          description: "My Description",
          buyFrom: "Woolworths",
          buyPrice: "12.50",
          buyAt: "2026-02-22",
        }),
        [],
        null,
      );
    });
  });

  it("shows success message after successful upload", async () => {
    vi.mocked(uploadFile).mockResolvedValue({ success: true });

    render(<UploadFile />);

    fireEvent.change(screen.getByLabelText(/title/i), {
      target: { value: "Test" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    expect(
      await screen.findByText(/uploaded successfully/i),
    ).toBeInTheDocument();
  });

  // File selection and preview tests

  it("shows file card with name and size after selecting a file", () => {
    render(<UploadFile />);

    fireEvent.change(getFileInput(), { target: { files: [mockImageFile()] } });

    expect(screen.getByText("test.jpg")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /scan/i })).toBeInTheDocument();
    expect(screen.queryByText(/choose file/i)).not.toBeInTheDocument();
  });

  it("shows file card for a pdf file", () => {
    render(<UploadFile />);

    fireEvent.change(getFileInput(), { target: { files: [mockPdfFile()] } });

    expect(screen.getByText("invoice.pdf")).toBeInTheDocument();
  });

  it("opens preview dialog when thumbnail button is clicked", () => {
    render(<UploadFile />);

    fireEvent.change(getFileInput(), { target: { files: [mockImageFile()] } });
    fireEvent.click(screen.getByTitle(/preview file/i));

    expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
  });

  it("closes preview dialog when close button is clicked", () => {
    render(<UploadFile />);

    fireEvent.change(getFileInput(), { target: { files: [mockImageFile()] } });
    fireEvent.click(screen.getByTitle(/preview file/i));
    fireEvent.click(screen.getByRole("button", { name: /close/i }));

    expect(HTMLDialogElement.prototype.close).toHaveBeenCalled();
  });

  it("removes file and restores upload input when × is clicked", () => {
    render(<UploadFile />);

    fireEvent.change(getFileInput(), { target: { files: [mockImageFile()] } });
    expect(screen.getByText("test.jpg")).toBeInTheDocument();

    fireEvent.click(screen.getByTitle(/remove file/i));

    expect(screen.queryByText("test.jpg")).not.toBeInTheDocument();
    expect(screen.getByText(/choose file/i)).toBeInTheDocument();
  });
});
