import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Document } from "@/api/wiki/document";
import type { Folder } from "@/api/wiki/folder";
import folderIcon from "@/assets/icon/file_icon_folder.svg";
import pdfIcon from "@/assets/icon/file_icon_pdf.svg";

vi.mock("@/api/wiki/document", async () => {
  const actual = await vi.importActual<typeof import("@/api/wiki/document")>("@/api/wiki/document");
  return {
    ...actual,
    documentApi: {
      ...actual.documentApi,
      download: vi.fn(),
    },
  };
});

import { DocumentCard } from "./document-card";
import { FolderCard } from "./folder-card";
import { FolderDetailPanel } from "./folder-detail-panel";

const folder: Folder = {
  id: "folder-1",
  name: "产品文档",
  description: "核心资料",
  created_at: "2026-04-05T00:00:00Z",
  updated_at: "2026-04-05T00:00:00Z",
  created_by: "u1",
  created_by_user: { id: "u1", full_name: "系统管理员" } as Folder["created_by_user"],
  updated_by_user: { id: "u1", full_name: "系统管理员" } as Folder["updated_by_user"],
};

const documentItem: Document = {
  id: "doc-1",
  filename: "产品说明.pdf",
  file_size: 2048,
  content_type: "application/pdf",
  file_path: "documents/doc-1/spec.pdf",
  status: "pending",
  chunk_count: 0,
  created_at: "2026-04-05T00:00:00Z",
  updated_at: "2026-04-05T00:00:00Z",
};

describe("folder icon surfaces", () => {
  it("renders the folder svg asset in folder cards", () => {
    render(
      <FolderCard folder={folder} onEdit={vi.fn()} onDelete={vi.fn()} onPermission={vi.fn()} />,
    );

    const image = screen.getByRole("img", { name: "文件夹图标" }) as HTMLImageElement;
    expect(image.src).toContain(folderIcon);
  });

  it("renders the folder svg asset in the detail panel", () => {
    render(<FolderDetailPanel folder={folder} onEdit={vi.fn()} />);

    const image = screen.getByRole("img", { name: "文件夹图标" }) as HTMLImageElement;
    expect(image.src).toContain(folderIcon);
  });

  it("renders the mapped svg asset in document cards", () => {
    render(<DocumentCard document={documentItem} onDelete={vi.fn()} />);

    const image = screen.getByRole("img", { name: "PDF 文件图标" }) as HTMLImageElement;
    expect(image.src).toContain(pdfIcon);
  });
});
