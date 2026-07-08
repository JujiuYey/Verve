import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { EditorCatalog } from "./editor-catalog";

function createEditorStub() {
  const commandCalls: number[] = [];

  return {
    state: {
      doc: {
        descendants: (callback: (node: any, pos: number) => boolean | void) => {
          callback({ type: { name: "heading" }, attrs: { level: 1 }, textContent: "第一章" }, 3);
          callback({ type: { name: "paragraph" }, attrs: {}, textContent: "正文" }, 8);
          callback({ type: { name: "heading" }, attrs: { level: 2 }, textContent: "第一节" }, 16);
        },
      },
    },
    on: vi.fn(),
    off: vi.fn(),
    chain: vi.fn(() => ({
      focus: () => ({
        setTextSelection: (from: number) => ({
          run: () => {
            commandCalls.push(from);
          },
        }),
      }),
    })),
    commandCalls,
  };
}

describe("EditorCatalog", () => {
  it("extracts heading items from the editor doc, supports toggle, and jumps to the selected heading", () => {
    const editor = createEditorStub();
    const onToggle = vi.fn();

    render(<EditorCatalog editor={editor as never} open onToggle={onToggle} />);

    expect(screen.getByRole("button", { name: "第一章" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "第一节" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "第一节" }));
    fireEvent.click(screen.getByRole("button", { name: "切换目录" }));

    expect(editor.commandCalls).toEqual([16]);
    expect(onToggle).toHaveBeenCalledTimes(1);
    expect(editor.on).toHaveBeenCalledWith("update", expect.any(Function));
    expect(editor.on).toHaveBeenCalledWith("selectionUpdate", expect.any(Function));
  });

  it("returns null when the catalog is closed", () => {
    const editor = createEditorStub();
    const { container } = render(
      <EditorCatalog editor={editor as never} open={false} onToggle={vi.fn()} />,
    );

    expect(container).toBeEmptyDOMElement();
  });
});
