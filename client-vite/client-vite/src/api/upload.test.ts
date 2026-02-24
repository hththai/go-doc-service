import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { getPurchases, uploadFile } from "./upload";

const API_URL = "http://localhost:8080";

beforeEach(() => {
  vi.stubEnv("VITE_API_URL", API_URL);
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

// --- getPurchases ---

describe("getPurchases", () => {
  it("makes a GET to /v1/auth/purchases with credentials", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify([]), { status: 200 }),
    );

    await getPurchases();

    expect(fetch).toHaveBeenCalledWith(
      `${API_URL}/v1/auth/purchases`,
      expect.objectContaining({ credentials: "include" }),
    );
  });

  it("returns purchases and makes fileUrl absolute", async () => {
    const serverData = [
      {
        id: "1",
        title: "Groceries",
        filename: "receipt.pdf",
        fileUrl: "/v1/auth/file/1",
        buyAt: "2025-01-15",
        buyFrom: "Woolworths",
        buyPrice: "87.45",
        items: [],
      },
    ];
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify(serverData), { status: 200 }),
    );

    const result = await getPurchases();

    expect(result[0].fileUrl).toBe(`${API_URL}/v1/auth/file/1`);
  });

  it("leaves purchases without fileUrl unchanged", async () => {
    const serverData = [
      {
        id: "2",
        title: "No File",
        filename: "manual.pdf",
        buyAt: "2025-02-20",
        buyFrom: "Coles",
        buyPrice: "50.00",
        items: [],
      },
    ];
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify(serverData), { status: 200 }),
    );

    const result = await getPurchases();

    expect(result[0].fileUrl).toBeUndefined();
  });

  it("returns all purchase fields from the server", async () => {
    const serverData = [
      {
        id: "3",
        title: "Office Supplies",
        filename: "receipt.jpg",
        fileUrl: "/v1/auth/file/3",
        buyAt: "2025-03-10",
        buyFrom: "Officeworks",
        buyPrice: "134.90",
        items: [
          {
            itemName: "A4 Paper",
            itemQty: "2",
            unitPrice: "12.00",
            subTotal: "24.00",
          },
        ],
      },
    ];
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify(serverData), { status: 200 }),
    );

    const result = await getPurchases();

    expect(result[0]).toMatchObject({
      id: "3",
      title: "Office Supplies",
      filename: "receipt.jpg",
      buyAt: "2025-03-10",
      buyFrom: "Officeworks",
      buyPrice: "134.90",
    });
    expect(result[0].items).toHaveLength(1);
  });

  it("throws when response is not ok", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 401 }));

    await expect(getPurchases()).rejects.toThrow("Failed to fetch purchases");
  });
});

// --- uploadFile ---

describe("uploadFile", () => {
  it("makes a POST to /v1/auth/upload with credentials", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ message: "Success" }), { status: 200 }),
    );

    await uploadFile({ title: "Test" }, []);

    expect(fetch).toHaveBeenCalledWith(
      `${API_URL}/v1/auth/upload`,
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
  });

  it("throws when title is missing", async () => {
    await expect(uploadFile({}, [])).rejects.toThrow("Title is required.");
  });

  it("throws when response is not ok", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 400 }));

    await expect(uploadFile({ title: "Test" }, [])).rejects.toThrow(
      "Upload failed",
    );
  });

  it("converts buyAt from YYYY-MM-DD to DD/MM/YYYY in the form body", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ message: "Success" }), { status: 200 }),
    );

    await uploadFile({ title: "Test", buyAt: "2025-01-15" }, []);

    const formData = vi.mocked(fetch).mock.calls[0][1]?.body as FormData;
    expect(formData.get("buyAt")).toBe("15/01/2025");
  });
});
