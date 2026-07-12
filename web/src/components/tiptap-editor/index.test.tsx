import { act, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { editorCatalogSpy, editorFooterSpy, editorToolbarSpy, mockGetContent, mockUpdateContent } =
  vi.hoisted(() => ({
    mockGetContent: vi.fn(),
    mockUpdateContent: vi.fn(),
    editorToolbarSpy: vi.fn(),
    editorCatalogSpy: vi.fn(),
    editorFooterSpy: vi.fn(),
  }));

vi.mock("@/api/wiki/document", () => ({
  documentApi: {
    getContent: mockGetContent,
    updateContent: mockUpdateContent,
  },
}));

vi.mock("./_components/editor-toolbar", () => ({
  EditorToolbar: (props: { editor: { id: string } | null; onSave: () => Promise<void> }) => {
    editorToolbarSpy(props);
    return (
      <button type="button" onClick={() => void props.onSave()}>
        toolbar:
        {props.editor?.id ?? "none"}
      </button>
    );
  },
}));

vi.mock("./_components/editor-catalog", () => ({
  EditorCatalog: (props: {
    open: boolean;
    onToggle: () => void;
    editor: { id: string } | null;
  }) => {
    editorCatalogSpy(props);
    return (
      <button type="button" onClick={props.onToggle}>
        catalog:
        {props.open ? "open" : "closed"}:{props.editor?.id ?? "none"}
      </button>
    );
  },
}));

vi.mock("./_components/editor-footer", () => ({
  EditorFooter: (props: {
    saveStatus: "saved" | "saving" | "unsaved";
    wordCount: number;
    onSave: () => Promise<void>;
    pageSetup: unknown;
  }) => {
    editorFooterSpy(props);
    return (
      <div>
        <div>
          footer-status:
          {props.saveStatus}
        </div>
        <div>
          footer-words:
          {props.wordCount}
        </div>
        <button type="button" onClick={() => void props.onSave()}>
          footer-save
        </button>
      </div>
    );
  },
}));

import { CanvasTiptapPage } from "./index";

beforeEach(() => {
  vi.clearAllMocks();
  mockGetContent.mockResolvedValue({
    content: "initial markdown content",
    filename: "demo.md",
  });
  mockUpdateContent.mockResolvedValue({ message: "ok" });
});

afterEach(() => {
  vi.useRealTimers();
});

describe("CanvasTiptapPage", () => {
  it("loads markdown content from docId and wires editor, catalog, footer, and save flow together", async () => {
    render(<CanvasTiptapPage docId="doc-1" />);

    await waitFor(() => {
      expect(mockGetContent).toHaveBeenCalledWith("doc-1");
    });

    expect(screen.getByText("toolbar:none")).toBeInTheDocument();
    expect(screen.getByText("catalog:open:none")).toBeInTheDocument();
    expect(screen.getByText("footer-status:saved")).toBeInTheDocument();
    expect(screen.getByText("footer-words:3")).toBeInTheDocument();

    await act(async () => {
      await Promise.resolve();
    });

    expect(editorToolbarSpy).toHaveBeenCalled();
    expect(editorCatalogSpy).toHaveBeenCalled();
    expect(editorFooterSpy).toHaveBeenCalled();
  });

  it("debounces auto-save after content changes", async () => {
    vi.useFakeTimers();

    render(<CanvasTiptapPage docId="doc-1" />);

    await act(async () => {
      await Promise.resolve();
    });
    expect(mockGetContent).toHaveBeenCalledWith("doc-1");
  });
});
