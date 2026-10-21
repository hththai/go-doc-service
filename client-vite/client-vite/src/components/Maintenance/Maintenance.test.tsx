import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import Maintenance from "./Maintenance";
import type { DocIssue } from "@/api/maintenance";

vi.mock("@/api/maintenance", async () => {
  const actual =
    await vi.importActual<typeof import("@/api/maintenance")>(
      "@/api/maintenance",
    );
  return {
    ...actual,
    verifyDocuments: vi.fn(),
    repairIssues: vi.fn(),
  };
});

import { verifyDocuments, repairIssues } from "@/api/maintenance";

const mockIssues: DocIssue[] = [
  {
    objId: 8,
    title: "Laptop receipt",
    code: 1,
    message: "1 extra file(s) next to 8.png",
    repairable: true,
    fileName: "8.png",
    extraFiles: ["8.pdf"],
  },
  {
    objId: 8,
    title: "Laptop receipt",
    code: 3,
    message: 'saved as ".pdf", 3 bytes; file is ".png", 4 bytes',
    repairable: true,
    fileName: "8.png",
  },
  {
    objId: 9,
    title: "Groceries",
    code: 2,
    message: "file is missing; re-upload it from Edit",
    repairable: false,
    fileName: "9.pdf",
  },
];

function renderMaintenance() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <Maintenance />
    </QueryClientProvider>,
  );
}

async function verify() {
  fireEvent.click(screen.getByRole("button", { name: "Verify my documents" }));
}

describe("Maintenance", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows a start hint before verifying", () => {
    renderMaintenance();
    expect(screen.getByText(/Click "Verify my documents"/)).toBeInTheDocument();
    expect(verifyDocuments).not.toHaveBeenCalled();
  });

  it("shows 'No issues found' when verify returns nothing", async () => {
    vi.mocked(verifyDocuments).mockResolvedValue([]);
    renderMaintenance();
    await verify();
    expect(await screen.findByText("No issues found.")).toBeInTheDocument();
  });

  it("lists issues and disables non-repairable ones", async () => {
    vi.mocked(verifyDocuments).mockResolvedValue(mockIssues);
    renderMaintenance();
    await verify();

    expect(await screen.findByText("Duplicate file")).toBeInTheDocument();
    expect(screen.getByText("Wrong file details")).toBeInTheDocument();
    expect(screen.getByText("Missing file")).toBeInTheDocument();
    expect(screen.getByText("extra: 8.pdf")).toBeInTheDocument();
    expect(
      screen.getByLabelText("Select Missing file for Groceries"),
    ).toBeDisabled();
    expect(
      screen.getByLabelText("Select Duplicate file for Laptop receipt"),
    ).toBeEnabled();
  });

  it("shows an error when verify fails", async () => {
    vi.mocked(verifyDocuments).mockRejectedValue(new Error("boom"));
    renderMaintenance();
    await verify();
    expect(
      await screen.findByText("Could not verify documents. Please try again."),
    ).toBeInTheDocument();
  });

  it("repairs only the checked issues after confirming", async () => {
    vi.mocked(verifyDocuments).mockResolvedValue(mockIssues);
    vi.mocked(repairIssues).mockResolvedValue({
      repaired: [{ objId: 8, code: 1, message: "moved 8.pdf to quarantine" }],
      skipped: [],
    });
    renderMaintenance();
    await verify();

    const repairButton = await screen.findByRole("button", {
      name: /Repair selected/,
    });
    expect(repairButton).toBeDisabled();

    fireEvent.click(
      screen.getByLabelText("Select Duplicate file for Laptop receipt"),
    );
    expect(repairButton).toHaveTextContent("Repair selected (1)");
    fireEvent.click(repairButton);

    // Nothing is sent until the user confirms.
    expect(repairIssues).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Repair" }));

    await waitFor(() => expect(repairIssues).toHaveBeenCalledTimes(1));
    expect(vi.mocked(repairIssues).mock.calls[0][0]).toEqual([
      { objId: 8, code: 1 },
    ]);
    expect(
      await screen.findByText("#8: moved 8.pdf to quarantine"),
    ).toBeInTheDocument();
    // Verifies again after repairing.
    await waitFor(() => expect(verifyDocuments).toHaveBeenCalledTimes(2));
  });

  it("select all picks only repairable issues", async () => {
    vi.mocked(verifyDocuments).mockResolvedValue(mockIssues);
    vi.mocked(repairIssues).mockResolvedValue({ repaired: [], skipped: [] });
    renderMaintenance();
    await verify();

    fireEvent.click(
      await screen.findByLabelText("Select all repairable issues"),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Repair selected (2)" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Repair" }));

    await waitFor(() => expect(repairIssues).toHaveBeenCalledTimes(1));
    expect(vi.mocked(repairIssues).mock.calls[0][0]).toEqual([
      { objId: 8, code: 1 },
      { objId: 8, code: 3 },
    ]);
  });

  it("cancel closes the confirm dialog without repairing", async () => {
    vi.mocked(verifyDocuments).mockResolvedValue(mockIssues);
    renderMaintenance();
    await verify();

    fireEvent.click(
      await screen.findByLabelText("Select Duplicate file for Laptop receipt"),
    );
    fireEvent.click(screen.getByRole("button", { name: /Repair selected/ }));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(screen.queryByRole("button", { name: "Repair" })).toBeNull();
    expect(repairIssues).not.toHaveBeenCalled();
  });
});
