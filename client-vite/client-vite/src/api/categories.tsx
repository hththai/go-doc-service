export type Category = {
  guid: string;
  name: string;
  color: string;
};

const BASE = `${import.meta.env.VITE_API_URL}/v1/auth/categories`;

export async function getCategories(): Promise<Category[]> {
  const res = await fetch(BASE, { credentials: "include" });
  if (!res.ok) throw new Error("Failed to fetch categories");
  return res.json();
}

export async function createCategory(
  name: string,
  color: string,
): Promise<Category> {
  const res = await fetch(BASE, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, color }),
  });
  if (!res.ok) throw new Error("Failed to create category");
  return res.json();
}

export async function updateCategory(
  guid: string,
  name: string,
  color: string,
): Promise<void> {
  const res = await fetch(`${BASE}/${guid}`, {
    method: "PATCH",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, color }),
  });
  if (!res.ok) throw new Error("Failed to update category");
}

export async function deleteCategory(guid: string): Promise<void> {
  const res = await fetch(`${BASE}/${guid}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Failed to delete category");
}
