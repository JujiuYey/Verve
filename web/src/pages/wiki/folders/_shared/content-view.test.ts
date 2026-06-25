import { describe, expect, it } from "vitest";

import type { Document } from "@/api/wiki/document";
import type { Folder } from "@/api/wiki/folder";

import { getFolderContentView } from "./content-view";

const baseFolder: Folder = {
  id: "folder-1",
  name: "产品文档",
  description: "产品资料",
  created_at: "2026-04-05T00:00:00Z",
  updated_at: "2026-04-05T00:00:00Z",
  created_by_user: { id: "u1", full_name: "系统管理员" } as Folder["created_by_user"],
  updated_by_user: { id: "u1", full_name: "系统管理员" } as Folder["updated_by_user"],
};

const baseDocument: Document = {
  id: "doc-1",
  filename: "产品需求.md",
  file_size: 2048,
  content_type: "text/markdown",
  file_path: "documents/doc-1/产品需求.md",
  status: "pending",
  chunk_count: 0,
  created_at: "2026-04-05T00:00:00Z",
  updated_at: "2026-04-05T00:00:00Z",
};

describe("getFolderContentView", () => {
  it("filters folders and documents with the same keyword and reports visible counts", () => {
    const view = getFolderContentView({
      folders: [
        baseFolder,
        { ...baseFolder, id: "folder-2", name: "开发文档", description: "工程说明" },
      ],
      documents: [baseDocument, { ...baseDocument, id: "doc-2", filename: "hr-database.md" }],
      searchKeyword: "产品",
    });

    expect(view.folders.map((folder) => folder.id)).toEqual(["folder-1"]);
    expect(view.documents.map((document) => document.id)).toEqual(["doc-1"]);
    expect(view.counts).toEqual({
      all: 2,
      folders: 1,
      documents: 1,
    });
  });

  it("returns original content when there is no keyword", () => {
    const view = getFolderContentView({
      folders: [baseFolder],
      documents: [baseDocument],
      searchKeyword: "",
    });

    expect(view.folders).toHaveLength(1);
    expect(view.documents).toHaveLength(1);
    expect(view.counts).toEqual({
      all: 2,
      folders: 1,
      documents: 1,
    });
  });
});
