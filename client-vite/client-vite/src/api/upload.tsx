export type UploadMetadata = Record<string, string>;

export async function uploadFile(metadata: UploadMetadata, file?: File | null) {
  if (!metadata.title) {
    throw new Error("Title is required.");
  }

  const form = new FormData();

  // Append all metadata fields dynamically
  for (const [key, value] of Object.entries(metadata)) {
    if (value) {
      // Map 'title' to 'name' for backend compatibility
      const fieldName = key === "title" ? "name" : key;
      // Convert buyAt from YYYY-MM-DD (date input) to DD/MM/YYYY (Australian format expected by server)
      if (key === "buyAt") {
        const [year, month, day] = value.split("-");
        form.append(fieldName, `${day}/${month}/${year}`);
      } else {
        form.append(fieldName, value);
      }
    }
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
