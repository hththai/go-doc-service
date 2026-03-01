export type UploadMetadata = Record<string, string>;

export async function getPurchases() {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/purchases`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Failed to fetch purchases");
  const data = await res.json();
  // fileUrl from server is a relative path like /v1/auth/file/:id — make it absolute
  return data.map((p: { fileUrl?: string }) =>
    p.fileUrl
      ? { ...p, fileUrl: `${import.meta.env.VITE_API_URL}${p.fileUrl}` }
      : p,
  );
}

export type UploadItem = {
  itemName: string;
  itemQty: string;
  unitPrice: string;
  subTotal: string;
};

export async function uploadFile(
  metadata: UploadMetadata,
  items: UploadItem[],
  file?: File | null,
) {
  if (!metadata.title) {
    throw new Error("Title is required.");
  }

  const form = new FormData();

  // Append all metadata fields dynamically
  for (const [key, value] of Object.entries(metadata)) {
    if (value) {
      // Convert buyAt from YYYY-MM-DD (date input) to DD/MM/YYYY (Australian format expected by server)
      if (key === "buyAt") {
        const [year, month, day] = value.split("-");
        form.append(key, `${day}/${month}/${year}`);
      } else {
        form.append(key, value);
      }
    }
  }

  if (items.length > 0) {
    form.append("items", JSON.stringify(items));
  }

  if (file) {
    form.append("file", file);
  }

  const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/upload`, {
    method: "POST",
    body: form,
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Upload failed");
  }

  return await res.json();
}

export type PurchasePayload = {
  title: string;
  buyFrom: string;
  buyAt: string;
  buyPrice: string;
  filename: string;
  items: UploadItem[];
};

export async function createPurchase(data: PurchasePayload) {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/purchases`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title: data.title,
      buyFrom: data.buyFrom,
      buyAt: data.buyAt,
      buyPrice: data.buyPrice,
      items: data.items,
    }),
  });
  if (!res.ok) throw new Error("Failed to create purchase");
  return res.json();
}

export async function updatePurchase(id: string, data: PurchasePayload) {
  const res = await fetch(
    `${import.meta.env.VITE_API_URL}/v1/auth/purchases/${id}`,
    {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: data.title,
        buyFrom: data.buyFrom,
        buyAt: data.buyAt,
        buyPrice: data.buyPrice,
        filename: data.filename,
        items: data.items,
      }),
    },
  );
  if (!res.ok) throw new Error("Failed to update purchase");
  return res.json();
}

export async function updatePurchaseWithFile(
  id: string,
  metadata: UploadMetadata,
  items: UploadItem[],
  file: File,
) {
  const form = new FormData();

  for (const [key, value] of Object.entries(metadata)) {
    if (value) {
      if (key === "buyAt") {
        const [year, month, day] = value.split("-");
        form.append("buyAt", `${day}/${month}/${year}`);
      } else {
        form.append(key, value);
      }
    }
  }

  if (items.length > 0) {
    form.append("items", JSON.stringify(items));
  }

  form.append("file", file);

  const res = await fetch(
    `${import.meta.env.VITE_API_URL}/v1/auth/purchases/${id}`,
    {
      method: "PATCH",
      body: form,
      credentials: "include",
    },
  );
  if (!res.ok) throw new Error("Failed to update purchase with file");
  return res.json();
}

export async function deletePurchase(id: string): Promise<void> {
  const res = await fetch(
    `${import.meta.env.VITE_API_URL}/v1/auth/purchases/${id}`,
    {
      method: "DELETE",
      credentials: "include",
    },
  );
  if (!res.ok) throw new Error("Failed to delete purchase");
}
