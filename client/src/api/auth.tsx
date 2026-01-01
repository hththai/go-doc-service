// TODO: store in api folder.
// Sample login function.
export async function loginRequest({ username, password }: { username: string, password: string }) {
    const res = await fetch("http://localhost:8088/loginjwt", {
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
    const res = await fetch("http://localhost:8088/v1/auth/refresh", {
        method: "POST",
        credentials: "include", // REQUIRED: sends refresh_token cookie
    });

    const data = await res.json();

    if (!res.ok) {
        throw new Error(data.error || "Refresh failed");
    }

    return data;
}
// api folder
export async function getMe() {
    const res = await fetch("http://localhost:8088/v1/auth/me", {
        credentials: "include", // sends cookies
    });

    if (res.status === 401) {
        return null;
    }

    if (!res.ok) {
        throw new Error("Not authenticated");
    }

    return res.json();
}
