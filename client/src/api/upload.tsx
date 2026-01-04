export async function uploadFile(title: string, file?: File | null) {
    if (!title) {
        throw new Error("Title and file are required.")
    }

    // 1. Read tokenId form localStorage.
    // TODO: move this key to env
    const tokenId = localStorage.getItem("tokenId");
    if (!tokenId) {
        throw new Error("Missing token.");
    }

    const form = new FormData();
    form.append("title", title);
    if (file) {
        form.append("file", file);
    }

    // Attach tokenId value into form.
    form.append("tokenId", tokenId)

    const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/upload`, {
        method: "POST",
        body: form,
        credentials: "include",
    });

    if (!res.ok) {
        throw new Error("Upload failed");
    }

    return await res.json()
}