import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import type { FolderTreeNode } from "@/api/wiki/folder";

const { uploadMock, listMock, treeMock } = vi.hoisted(() => ({
  uploadMock: vi.fn(),
  listMock: vi.fn(),
  treeMock: vi.fn(),
}));

vi.mock("@/api/wiki/document", () => ({
  documentApi: {
    upload: uploadMock,
  },
}));

vi.mock("@/api/wiki/folder", () => ({
  folderApi: {
    list: listMock,
    tree: treeMock,
  },
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

import { UploadDialog } from "./upload-dialog";

beforeAll(() => {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
});

const baseTreeNode = {
  description: "",
  parent_id: undefined,
  user_id: "",
  created_at: "2026-04-04T00:00:00Z",
  updated_at: "2026-04-04T00:00:00Z",
};

const providedTree: FolderTreeNode[] = [
  {
    ...baseTreeNode,
    id: "folder-1",
    name: "项目文档",
    hasChildren: false,
    children: [],
  },
];

const fetchedTree: FolderTreeNode[] = [
  {
    ...baseTreeNode,
    id: "folder-2",
    name: "接口文档",
    hasChildren: false,
    children: [],
  },
];

describe("UploadDialog", () => {
  beforeEach(() => {
    uploadMock.mockReset();
    listMock.mockReset();
    treeMock.mockReset();
  });

  it("uses provided folder tree data in the tree select", async () => {
    listMock.mockResolvedValue([]);
    treeMock.mockResolvedValue([]);

    render(<UploadDialog open onOpenChange={vi.fn()} folderTree={providedTree} />);

    fireEvent.click(screen.getByRole("combobox"));

    expect(screen.getByText("项目文档")).toBeVisible();

    fireEvent.click(screen.getByText("项目文档"));

    await waitFor(() => {
      expect(screen.getByRole("combobox")).toHaveTextContent("项目文档");
    });
  });

  it("falls back to folderApi.tree when folderTree is not provided", async () => {
    listMock.mockResolvedValue([]);
    treeMock.mockResolvedValue(fetchedTree);

    render(<UploadDialog open onOpenChange={vi.fn()} />);

    await waitFor(() => {
      expect(treeMock).toHaveBeenCalledTimes(1);
    });

    fireEvent.click(screen.getByRole("combobox"));

    expect(screen.getByText("接口文档")).toBeVisible();
  });
});
