import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { loginRequest, getMe } from "./auth";

const API_URL = "http://localhost:8080";

beforeEach(() => {
  vi.stubEnv("VITE_API_URL", API_URL);
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

// --- loginRequest ---

describe("loginRequest", () => {
  it("makes a POST to /login with credentials and JSON body", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ userId: 1 }), { status: 200 }),
    );

    await loginRequest({ username: "alice", password: "secret" });

    expect(fetch).toHaveBeenCalledWith(
      `${API_URL}/login`,
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        body: JSON.stringify({ username: "alice", password: "secret" }),
      }),
    );
  });

  it("returns server data on success", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ userId: 42 }), { status: 200 }),
    );

    const result = await loginRequest({
      username: "alice",
      password: "secret",
    });

    expect(result).toEqual({ userId: 42 });
  });

  it("throws the server error message on non-ok response", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ error: "Incorrect Password" }), {
        status: 400,
      }),
    );

    await expect(
      loginRequest({ username: "alice", password: "wrong" }),
    ).rejects.toThrow("Incorrect Password");
  });

  it("throws a fallback message when server returns no error field", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({}), { status: 500 }),
    );

    await expect(
      loginRequest({ username: "alice", password: "secret" }),
    ).rejects.toThrow("Login failed");
  });
});

// --- getMe ---

describe("getMe", () => {
  it("returns user data when access token is valid", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ userId: 99 }), { status: 200 }),
    );

    const result = await getMe();

    expect(result).toEqual({ userId: 99 });
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch).toHaveBeenCalledWith(
      `${API_URL}/v1/auth/users/me`,
      expect.objectContaining({ method: "GET", credentials: "include" }),
    );
  });

  it("attempts refresh on 401 and returns user when refresh succeeds", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // /users/me → expired
      .mockResolvedValueOnce(new Response(null, { status: 200 })) // /refresh → ok
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ userId: 5 }), { status: 200 }), // retry /users/me → ok
      );

    const result = await getMe();

    expect(result).toEqual({ userId: 5 });
    expect(fetch).toHaveBeenCalledTimes(3);
    expect(fetch).toHaveBeenNthCalledWith(
      2,
      `${API_URL}/v1/auth/refresh`,
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
  });

  it("returns null when both access and refresh tokens are expired", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // /users/me → 401
      .mockResolvedValueOnce(new Response(null, { status: 401 })); // /refresh → 401

    const result = await getMe();

    expect(result).toBeNull();
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("returns null when retry after refresh still fails", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // /users/me → 401
      .mockResolvedValueOnce(new Response(null, { status: 200 })) // /refresh → ok
      .mockResolvedValueOnce(new Response(null, { status: 500 })); // retry → server error

    const result = await getMe();

    expect(result).toBeNull();
  });
});
