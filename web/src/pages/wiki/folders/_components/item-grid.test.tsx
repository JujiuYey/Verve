import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Document } from "@/api/wiki/document";
import type { Folder } from "@/api/wiki/folder";

vi.mock("./document-card", () => ({
  DocumentCard: ({ document }: { document: Document }) => <div>{document.filename}</div>,
}));

vi.mock("./folder-card", () => ({
  FolderCard: ({ folder }: { folder: Folder }) => <div>{folder.name}</div>,
}));

import { ItemGrid } from "./item-grid";

const documents: Document[] = [
  {
    id: "doc-1",
    filename: "系统设计说明书.md",
    file_size: 1024,
    content_type: "text/markdown",
    file_path: "documents/doc-1/design.md",
    status: "completed",
    chunk_count: 4,
    created_at: "2026-04-05T00:00:00Z",
    updated_at: "2026-04-05T00:00:00Z",
  },
];

describe("ItemGrid", () => {
  it("uses a wider grid for the documents section", () => {
    render(
      <ItemGrid
        folders={[]}
        documents={documents}
        activeTab="documents"
        onEditFolder={vi.fn()}
        onDeleteFolder={vi.fn()}
        onFolderPermission={vi.fn()}
        onEnterFolder={vi.fn()}
        onDeleteDocument={vi.fn()}
      />,
    );

    const section = screen.getByRole("heading", { level: 3, name: /文档/ }).parentElement;
    const grid = section?.querySelector("div.grid");

    expect(grid).toHaveClass("grid-cols-1", "md:grid-cols-2", "xl:grid-cols-3");
    expect(grid).not.toHaveClass("lg:grid-cols-3", "xl:grid-cols-4");
  });
});
