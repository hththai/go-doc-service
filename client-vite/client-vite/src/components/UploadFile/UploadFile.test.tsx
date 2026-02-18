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

    // Fill in fields
    fireEvent.change(titleInput, { target: { value: "Test Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });

    // Click clear
    fireEvent.click(screen.getByRole("button", { name: /clear/i }));

    // Fields should be empty
    expect(titleInput).toHaveValue("");
    expect(descriptionInput).toHaveValue("");
  });

  it("calls uploadFile with metadata on submit", async () => {
    vi.mocked(uploadFile).mockResolvedValue({ success: true });

    render(<UploadFile />);

    const titleInput = screen.getByLabelText(/title/i);
    const descriptionInput = screen.getByLabelText(/description/i);

    fireEvent.change(titleInput, { target: { value: "My Title" } });
    fireEvent.change(descriptionInput, { target: { value: "My Description" } });

    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    expect(uploadFile).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "My Title",
        description: "My Description",
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
