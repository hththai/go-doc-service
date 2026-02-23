import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import LoginForm from "./Login";

const noop = vi.fn().mockResolvedValue(undefined);

describe("LoginForm", () => {
  it("renders username and password inputs", () => {
    render(<LoginForm onSubmit={noop} />);

    expect(screen.getByRole("textbox")).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it("renders sign-in button", () => {
    render(<LoginForm onSubmit={noop} />);

    expect(
      screen.getByRole("button", { name: /sign in/i }),
    ).toBeInTheDocument();
  });

  it("username input is required", () => {
    render(<LoginForm onSubmit={noop} />);

    expect(screen.getByRole("textbox")).toBeRequired();
  });

  it("password input is required", () => {
    render(<LoginForm onSubmit={noop} />);

    expect(screen.getByLabelText(/password/i)).toBeRequired();
  });

  it("password input has type password", () => {
    render(<LoginForm onSubmit={noop} />);

    expect(screen.getByLabelText(/password/i)).toHaveAttribute(
      "type",
      "password",
    );
  });

  it("calls onSubmit with username and password on submit", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(<LoginForm onSubmit={onSubmit} />);

    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "alice" },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: "secret123" },
    });

    fireEvent.submit(
      screen.getByRole("button", { name: /sign in/i }).closest("form")!,
    );

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith("alice", "secret123");
    });
  });

  it("shows server error when error prop is set", () => {
    render(<LoginForm onSubmit={noop} error="Invalid credentials" />);

    expect(screen.getByText("Invalid credentials")).toBeInTheDocument();
  });

  it("shows username validation error after blur when empty", async () => {
    render(<LoginForm onSubmit={noop} />);

    fireEvent.blur(screen.getByRole("textbox"));

    await waitFor(() => {
      expect(screen.getByText("Username is required")).toBeInTheDocument();
    });
  });

  it("shows password validation error after blur when too short", async () => {
    render(<LoginForm onSubmit={noop} />);

    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: "abc" },
    });
    fireEvent.blur(screen.getByLabelText(/password/i));

    await waitFor(() => {
      expect(screen.getByText(/at least 6 characters/i)).toBeInTheDocument();
    });
  });
});
