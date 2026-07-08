import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { EditorToolbar } from "./index";

function createEditorStub() {
  const viewDom = document.createElement("div");
  const chain = {
    clearNodes: vi.fn(() => chain),
    focus: vi.fn(() => chain),
    insertContent: vi.fn(() => chain),
    insertTable: vi.fn(() => chain),
    redo: vi.fn(() => chain),
    run: vi.fn(() => true),
    setBold: vi.fn(() => chain),
    setColor: vi.fn(() => chain),
    setFontFamily: vi.fn(() => chain),
    setFontSize: vi.fn(() => chain),
    setHeading: vi.fn(() => chain),
    setHighlight: vi.fn(() => chain),
    setImage: vi.fn(() => chain),
    setItalic: vi.fn(() => chain),
    setLink: vi.fn(() => chain),
    setParagraph: vi.fn(() => chain),
    setStrike: vi.fn(() => chain),
    setTextAlign: vi.fn(() => chain),
    setTextSelection: vi.fn(() => chain),
    setUnderline: vi.fn(() => chain),
    toggleBold: vi.fn(() => chain),
    toggleBlockquote: vi.fn(() => chain),
    toggleBulletList: vi.fn(() => chain),
    toggleCode: vi.fn(() => chain),
    toggleCodeBlock: vi.fn(() => chain),
    toggleHeading: vi.fn(() => chain),
    toggleHighlight: vi.fn(() => chain),
    toggleItalic: vi.fn(() => chain),
    toggleOrderedList: vi.fn(() => chain),
    toggleStrike: vi.fn(() => chain),
    toggleTaskList: vi.fn(() => chain),
    toggleUnderline: vi.fn(() => chain),
    undo: vi.fn(() => chain),
    unsetAllMarks: vi.fn(() => chain),
  };

  return {
    chain: vi.fn(() => chain),
    getAttributes: vi.fn((name: string) => {
      if (name === "highlight") {
        return { color: "#fff3cd" };
      }
      if (name === "textStyle") {
        return { color: "#e63946" };
      }
      return {};
    }),
    isActive: vi.fn(
      (nameOrAttrs: string | Record<string, unknown>, attrs?: Record<string, unknown>) => {
        if (nameOrAttrs === "bold") return true;
        if (nameOrAttrs === "bulletList") return false;
        if (nameOrAttrs === "orderedList") return true;
        if (nameOrAttrs === "heading" && attrs?.level === 2) return true;
        if (typeof nameOrAttrs === "object" && nameOrAttrs.textAlign === "center") return true;
        return false;
      },
    ),
    off: vi.fn(),
    on: vi.fn(),
    state: {
      doc: {
        textContent: "alpha beta alpha",
      },
    },
    view: {
      dom: viewDom,
    },
  };
}

describe("EditorToolbar", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("executes formatting and insert actions through the TipTap command chain", () => {
    const editor = createEditorStub();
    const promptSpy = vi.spyOn(window, "prompt").mockReturnValue("https://example.com");

    render(<EditorToolbar editor={editor as never} docTitle="测试文档" onSave={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "清除格式" }));
    fireEvent.click(screen.getByRole("button", { name: "粗体" }));
    fireEvent.click(screen.getByRole("button", { name: "斜体" }));
    fireEvent.click(screen.getByRole("button", { name: "下划线" }));
    fireEvent.click(screen.getByRole("button", { name: "删除线" }));
    fireEvent.click(screen.getByRole("button", { name: "行内代码" }));

    fireEvent.click(screen.getByRole("button", { name: "文字颜色" }));
    fireEvent.click(screen.getByRole("button", { name: "#e63946" }));
    fireEvent.click(screen.getByRole("button", { name: "高亮颜色" }));
    fireEvent.click(screen.getByRole("button", { name: "#fff3cd" }));

    fireEvent.click(screen.getByRole("button", { name: "插入链接" }));
    fireEvent.click(screen.getByRole("button", { name: "插入表格" }));
    fireEvent.click(screen.getByRole("button", { name: "代码块" }));
    fireEvent.click(screen.getByRole("button", { name: "任务列表" }));

    expect(chainCalls(editor).clearNodes).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).unsetAllMarks).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleBold).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleItalic).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleUnderline).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleStrike).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleCode).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).setColor).toHaveBeenCalledWith("#e63946");
    expect(chainCalls(editor).toggleHighlight).toHaveBeenCalledWith({ color: "#fff3cd" });
    expect(chainCalls(editor).setLink).toHaveBeenCalledWith({ href: "https://example.com" });
    expect(chainCalls(editor).insertTable).toHaveBeenCalledWith({
      rows: 3,
      cols: 3,
      withHeaderRow: true,
    });
    expect(chainCalls(editor).toggleCodeBlock).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).toggleTaskList).toHaveBeenCalledTimes(1);
    expect(promptSpy).toHaveBeenCalled();
  });

  it("reflects active paragraph state and applies heading, alignment, and list changes", () => {
    const editor = createEditorStub();

    render(<EditorToolbar editor={editor as never} docTitle="测试文档" onSave={vi.fn()} />);

    const headingSelect = screen.getByRole("combobox", { name: "段落样式" });
    expect(headingSelect).toHaveValue("h2");
    fireEvent.change(headingSelect, { target: { value: "paragraph" } });
    fireEvent.click(screen.getByRole("button", { name: "居中对齐" }));
    fireEvent.click(screen.getByRole("button", { name: "有序列表" }));

    expect(chainCalls(editor).setParagraph).toHaveBeenCalledTimes(1);
    expect(chainCalls(editor).setTextAlign).toHaveBeenCalledWith("center");
    expect(chainCalls(editor).toggleOrderedList).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "粗体" })).toHaveAttribute("data-active", "true");
    expect(screen.getByRole("button", { name: "有序列表" })).toHaveAttribute("data-active", "true");
  });

  it("opens search, jumps to the next match, and delegates printing to the browser", () => {
    const editor = createEditorStub();
    const printSpy = vi.spyOn(window, "print").mockImplementation(() => undefined);

    render(<EditorToolbar editor={editor as never} docTitle="测试文档" onSave={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "搜索" }));

    const dialog = screen.getByRole("dialog", { name: "搜索" });
    fireEvent.change(within(dialog).getByLabelText("搜索内容"), {
      target: { value: "alpha" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "下一个" }));
    fireEvent.click(screen.getByRole("button", { name: "打印", hidden: true }));

    expect(chainCalls(editor).setTextSelection).toHaveBeenCalledWith({ from: 1, to: 6 });
    expect(printSpy).toHaveBeenCalledTimes(1);
  });

  it("keeps already-active styles intact when applying the format painter", () => {
    const editor = createEditorStub();

    render(<EditorToolbar editor={editor as never} docTitle="测试文档" onSave={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "格式刷" }));
    editor.getAttributes.mockImplementation((name: string) => {
      if (name === "highlight") {
        return {};
      }
      if (name === "textStyle") {
        return {};
      }
      return {};
    });
    fireEvent.click(editor.view.dom);

    expect(chainCalls(editor).toggleBold).not.toHaveBeenCalled();
    expect(chainCalls(editor).toggleOrderedList).not.toHaveBeenCalled();
    expect(chainCalls(editor).setColor).toHaveBeenCalledWith("#e63946");
    expect(screen.getByRole("button", { name: "格式刷" })).toHaveAttribute("data-active", "false");
  });
});

function chainCalls(editor: ReturnType<typeof createEditorStub>) {
  return editor.chain.mock.results[0]?.value ?? editor.chain();
}
