export type UploadMetadata = Record<string, string>;

export async function uploadFile(metadata: UploadMetadata, file?: File | null) {
  if (!metadata.title) {
    throw new Error("Title is required.");
  }

  // 1. Read tokenId form localStorage.
  // TODO: move this key to env.
  const tokenId = localStorage.getItem("tokenId");
  if (!tokenId) {
    throw new Error("Missing token.");
  }

  const form = new FormData();

  // Append all metadata fields dynamically
  for (const [key, value] of Object.entries(metadata)) {
    if (value) {
      // Map 'title' to 'name' for backend compatibility
      const fieldName = key === "title" ? "name" : key;
      form.append(fieldName, value);
    }
  }

  if (file) {
    form.append("file", file);
  }

  // Attach tokenId value into form.
  form.append("tokenId", tokenId);

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
