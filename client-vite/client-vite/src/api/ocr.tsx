export interface OcrInvoice {
  seller: string;
  abn: string;
  document_date: string;
  order_no: string;
  order_date: string;
  billing_address: string;
  delivery_address: string;
  items: Array<{
    description: string;
    qty: string;
    unit_price: string;
    gst_rate: string;
    subtotal: string;
  }>;
  shipping_charges: string;
  total: string;
}

export interface OcrPurchaseInfo {
  buyAt: string | null;
  buyFrom: string;
  buyPrice: string;
}

export interface OcrResponse {
  invoice: OcrInvoice;
  purchaseInfo: OcrPurchaseInfo;
}

export async function scanInvoice(file: File): Promise<OcrResponse> {
  const form = new FormData();
  form.append("file", file);

  const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/ocr`, {
    method: "POST",
    body: form,
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("OCR scan failed");
  }

  return await res.json();
}
