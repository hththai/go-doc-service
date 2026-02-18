// Sample login function.
export async function loginRequest({
  username,
  password,
}: {
  username: string;
  password: string;
}) {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/login`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });

  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || "Login failed");
  }

  return data;
}

export async function refreshRequest() {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/refresh`, {
    method: "POST",
    credentials: "include", // REQUIRED: sends refresh_token cookie
  });

  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || "Refresh failed");
  }

  return data;
}

export async function getMe(): Promise<{ userId: number } | null> {
  let res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/users/me`, {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    // Access token expired — silently try to refresh and retry once
    const refreshRes = await fetch(
      `${import.meta.env.VITE_API_URL}/v1/auth/refresh`,
      { method: "POST", credentials: "include" },
    );
    if (!refreshRes.ok) return null; // refresh token also gone

    res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/users/me`, {
      method: "GET",
      credentials: "include",
    });
  }

  if (!res.ok) return null;

  return res.json();
}
