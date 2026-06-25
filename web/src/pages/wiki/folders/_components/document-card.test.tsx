import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Document } from "@/api/wiki/document";

import { DocumentCard } from "./document-card";

const documentItem: Document = {
  id: "doc-1",
  filename: "1.1_系统架构设计与部署说明文档-最终版.md",
  file_size: 9728,
  content_type: "text/markdown",
  file_path: "documents/doc-1/architecture.md",
  status: "pending",
  chunk_count: 12,
  created_at: "2026-04-05T00:00:00Z",
  updated_at: "2026-04-05T00:00:00Z",
};

describe("DocumentCard", () => {
  it("keeps long filenames readable in a more compact card layout", () => {
    render(<DocumentCard document={documentItem} onDelete={vi.fn()} />);

    const title = screen.getByRole("heading", { name: documentItem.filename });
    const card = title.closest("div[class*='group relative']");
    const menuButton = screen.getByRole("button");
    const icon = screen.getByAltText("Markdown 文件图标");

    expect(card).toHaveClass("min-h-28");
    expect(card).toHaveClass("p-3");
    expect(card).not.toHaveClass("min-h-36");
    expect(card).not.toHaveClass("p-4");
    expect(title).toHaveClass("line-clamp-3");
    expect(title).toHaveClass("[overflow-wrap:anywhere]");
    expect(title).not.toHaveClass("truncate");
    expect(icon).toHaveClass("h-6", "w-6");
    expect(menuButton).toHaveClass("h-7", "w-7");
    expect(screen.queryByText("9.5 KB")).not.toBeInTheDocument();
    expect(screen.queryByText("12 块")).not.toBeInTheDocument();
    expect(screen.queryByText("待处理")).not.toBeInTheDocument();
  });
});
