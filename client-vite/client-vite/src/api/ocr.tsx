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

export interface OcrResponse {
  invoice: OcrInvoice;
}

/** Converts OCR date string to YYYY-MM-DD for <input type="date"> */
export function parseOcrDate(dateStr: string): string {
  if (!dateStr) return "";

  // DD.MM.YYYY (e.g. Amazon EU format)
  const dotMatch = /^(\d{1,2})\.(\d{1,2})\.(\d{4})$/.exec(dateStr);
  if (dotMatch) {
    const [, d, m, y] = dotMatch;
    return `${y}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
  }

  // DD/MM/YYYY (Australian format)
  const slashMatch = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/.exec(dateStr);
  if (slashMatch) {
    const [, d, m, y] = slashMatch;
    return `${y}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
  }

  // DD-MM-YYYY (e.g. "10-11-2025")
  const dashMatch = /^(\d{1,2})-(\d{1,2})-(\d{4})$/.exec(dateStr);
  if (dashMatch) {
    const [, d, m, y] = dashMatch;
    return `${y}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
  }

  // DD/MM/YY (e.g. "17/08/24" → 2024)
  const shortSlashMatch = /^(\d{1,2})\/(\d{1,2})\/(\d{2})$/.exec(dateStr);
  if (shortSlashMatch) {
    const [, d, m, y] = shortSlashMatch;
    const fullYear = `20${y}`;
    return `${fullYear}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
  }

  // ISO 8601: YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS...
  const isoMatch = /^(\d{4})-(\d{2})-(\d{2})/.exec(dateStr);
  if (isoMatch) {
    return `${isoMatch[1]}-${isoMatch[2]}-${isoMatch[3]}`;
  }

  return "";
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
