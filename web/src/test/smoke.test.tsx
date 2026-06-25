import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

describe("vitest setup", () => {
  it("provides a browser-like window object", () => {
    expect(typeof window).toBe("object");
  });

  it("renders React components with testing-library", () => {
    render(<button type="button">Save</button>);

    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
  });
});
