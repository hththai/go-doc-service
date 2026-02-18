import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import LoginForm from "./Login";

const defaultProps = {
  username: "",
  password: "",
  setUsername: vi.fn(),
  setPassword: vi.fn(),
  handleSubmit: vi.fn(),
};

describe("LoginForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders username and password inputs", () => {
    render(<LoginForm {...defaultProps} />);

    expect(screen.getByRole("textbox")).toBeInTheDocument(); // username
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it("renders sign-in button", () => {
    render(<LoginForm {...defaultProps} />);

    expect(
      screen.getByRole("button", { name: /sign in/i }),
    ).toBeInTheDocument();
  });

  it("reflects controlled username value", () => {
    render(<LoginForm {...defaultProps} username="alice" />);

    expect(screen.getByRole("textbox")).toHaveValue("alice");
  });

  it("reflects controlled password value", () => {
    render(<LoginForm {...defaultProps} password="secret" />);

    expect(screen.getByLabelText(/password/i)).toHaveValue("secret");
  });

  it("calls setUsername when username input changes", () => {
    const setUsername = vi.fn();
    render(<LoginForm {...defaultProps} setUsername={setUsername} />);

    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "alice" },
    });

    expect(setUsername).toHaveBeenCalledWith("alice");
  });

  it("calls setPassword when password input changes", () => {
    const setPassword = vi.fn();
    render(<LoginForm {...defaultProps} setPassword={setPassword} />);

    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: "mypass" },
    });

    expect(setPassword).toHaveBeenCalledWith("mypass");
  });

  it("calls handleSubmit when form is submitted", () => {
    const handleSubmit = vi.fn((e: React.FormEvent) => e.preventDefault());
    render(<LoginForm {...defaultProps} handleSubmit={handleSubmit} />);

    const form = screen
      .getByRole("button", { name: /sign in/i })
      .closest("form")!;
    fireEvent.submit(form);

    expect(handleSubmit).toHaveBeenCalledOnce();
  });

  it("password input has type password", () => {
    render(<LoginForm {...defaultProps} />);

    expect(screen.getByLabelText(/password/i)).toHaveAttribute(
      "type",
      "password",
    );
  });

  it("username input is required", () => {
    render(<LoginForm {...defaultProps} />);

    expect(screen.getByRole("textbox")).toBeRequired();
  });

  it("password input is required", () => {
    render(<LoginForm {...defaultProps} />);

    expect(screen.getByLabelText(/password/i)).toBeRequired();
  });
});
