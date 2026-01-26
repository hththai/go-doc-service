// TODO: store in api folder.
// Sample login function.
export async function loginRequest({ username, password }: { username: string, password: string }) {
    const res = await fetch(`${import.meta.env.VITE_API_URL}/login`, {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();

    if (!res.ok) {
        throw new Error(data.error || "Login failed");
    }

    return data;
}

export async function refreshRequest() {
    const res = await fetch(`${import.meta.env.VITE_API_URL} / v1 / auth / refresh`, {
        method: "POST",
        credentials: "include", // REQUIRED: sends refresh_token cookie
    });

    const data = await res.json();

    if (!res.ok) {
        throw new Error(data.error || "Refresh failed");
    }

    return data;
}

// api folder.
// Add requirement to have tokenId.
export async function getMe({ tokenId }: { tokenId: string }) {
    const res = await fetch(`${import.meta.env.VITE_API_URL}/v1/auth/users/me`, {
        method: "POST",
        credentials: "include", // sends cookies
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ tokenId })
    });

    if (res.status === 401) {
        return null;
    }

    if (!res.ok) {
        throw new Error("Not authenticated");
    }

    return res.json();
}
