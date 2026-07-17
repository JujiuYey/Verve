import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Source } from "./sources";

describe("Source", () => {
  it("renders an anchor when href is available", () => {
    render(<Source href="https://example.com/wiki" title="Wiki" />);
    expect(screen.getByRole("link", { name: "Wiki" })).toHaveAttribute(
      "href",
      "https://example.com/wiki",
    );
  });

  it("renders a non-interactive source when href is absent", () => {
    render(<Source title="database.md" />);
    expect(screen.getByText("database.md")).toBeInTheDocument();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });
});
