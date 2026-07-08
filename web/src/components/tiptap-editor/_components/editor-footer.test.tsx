import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { EditorPageSetup } from "../_hooks/use-editor-page-setup";
import { EditorFooter } from "./editor-footer";

function createMockPageSetup(overrides?: Partial<EditorPageSetup>): EditorPageSetup {
  return {
    zoom: 1,
    margins: { top: 16, bottom: 16, left: 32, right: 32 },
    zoomIn: vi.fn(),
    zoomOut: vi.fn(),
    zoomReset: vi.fn(),
    setZoom: vi.fn(),
    setMargins: vi.fn(),
    resetMargins: vi.fn(),
    ...overrides,
  } as EditorPageSetup;
}

describe("EditorFooter", () => {
  it("renders word count, reflects save status, and delegates save clicks", () => {
    const onSave = vi.fn();
    const pageSetup = createMockPageSetup();

    const { rerender } = render(
      <EditorFooter saveStatus="saved" wordCount={128} onSave={onSave} pageSetup={pageSetup} />,
    );

    expect(screen.getByText(/字数:\s*128/)).toBeInTheDocument();
    expect(screen.getByText("已保存")).toBeInTheDocument();

    rerender(
      <EditorFooter saveStatus="saving" wordCount={256} onSave={onSave} pageSetup={pageSetup} />,
    );
    expect(screen.getByText("保存中...")).toBeInTheDocument();
    expect(screen.getByText(/字数:\s*256/)).toBeInTheDocument();

    rerender(
      <EditorFooter saveStatus="unsaved" wordCount={512} onSave={onSave} pageSetup={pageSetup} />,
    );
    expect(screen.getByText("未保存")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "保存" }));
    expect(onSave).toHaveBeenCalledTimes(1);
  });

  it("displays zoom percentage and delegates zoom actions", () => {
    const pageSetup = createMockPageSetup({ zoom: 0.8 });

    render(
      <EditorFooter saveStatus="saved" wordCount={0} onSave={vi.fn()} pageSetup={pageSetup} />,
    );

    expect(screen.getByText("80%")).toBeInTheDocument();

    fireEvent.click(screen.getByTitle("缩小"));
    expect(pageSetup.zoomOut).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByTitle("放大"));
    expect(pageSetup.zoomIn).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByTitle("重置缩放"));
    expect(pageSetup.zoomReset).toHaveBeenCalledTimes(1);
  });
});
