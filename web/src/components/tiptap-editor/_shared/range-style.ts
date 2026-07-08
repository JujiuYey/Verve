import type { Editor } from "@tiptap/react";

import type { RangeStyle } from "./const";
import { DEFAULT_RANGE_STYLE } from "./const";

type CommandChain = ReturnType<Editor["chain"]> & Record<string, (...args: unknown[]) => unknown>;

export function readRangeStyle(editor: Editor | null | undefined): RangeStyle {
  if (!editor) {
    return DEFAULT_RANGE_STYLE;
  }

  return {
    bold: !!editor.isActive?.("bold"),
    italic: !!editor.isActive?.("italic"),
    underline: !!editor.isActive?.("underline"),
    strikethrough: !!editor.isActive?.("strike"),
    code: !!editor.isActive?.("code"),
    highlight: editor.getAttributes?.("highlight")?.color || null,
    textColor: editor.getAttributes?.("textStyle")?.color || null,
    fontFamily: editor.getAttributes?.("textStyle")?.fontFamily || null,
    fontSize: editor.getAttributes?.("textStyle")?.fontSize || null,
    textAlign: (editor.getAttributes?.("paragraph")?.textAlign as RangeStyle["textAlign"]) || null,
    heading: editor.isActive?.("heading") ? editor.getAttributes?.("heading")?.level : null,
    bulletList: !!editor.isActive?.("bulletList"),
    orderedList: !!editor.isActive?.("orderedList"),
    blockquote: !!editor.isActive?.("blockquote"),
  };
}

function callCommand(chain: CommandChain, command: string, ...args: unknown[]) {
  const action = chain[command];

  if (typeof action !== "function") {
    return false;
  }

  action(...args);
  return true;
}

export function applyRangeStyle(
  editor: Editor | null | undefined,
  style: Partial<RangeStyle> | null | undefined,
) {
  if (!editor || !style) {
    return;
  }

  const chain = editor.chain().focus() as CommandChain;
  let hasChanges = false;

  const applyIfInactive = (
    shouldApply: boolean | undefined,
    isActive: boolean,
    preferredCommand: string,
    fallbackCommand: string,
    ...args: unknown[]
  ) => {
    if (!shouldApply || isActive) {
      return;
    }

    const didApply =
      callCommand(chain, preferredCommand, ...args) || callCommand(chain, fallbackCommand, ...args);

    if (didApply) {
      hasChanges = true;
    }
  };

  applyIfInactive(style.bold, !!editor.isActive?.("bold"), "setBold", "toggleBold");
  applyIfInactive(style.italic, !!editor.isActive?.("italic"), "setItalic", "toggleItalic");
  applyIfInactive(
    style.underline,
    !!editor.isActive?.("underline"),
    "setUnderline",
    "toggleUnderline",
  );
  applyIfInactive(style.strikethrough, !!editor.isActive?.("strike"), "setStrike", "toggleStrike");
  applyIfInactive(style.code, !!editor.isActive?.("code"), "setCode", "toggleCode");

  if (style.textColor && editor.getAttributes?.("textStyle")?.color !== style.textColor) {
    if (callCommand(chain, "setColor", style.textColor)) {
      hasChanges = true;
    }
  }

  if (style.highlight && editor.getAttributes?.("highlight")?.color !== style.highlight) {
    const didApply =
      callCommand(chain, "setHighlight", { color: style.highlight }) ||
      callCommand(chain, "toggleHighlight", { color: style.highlight });

    if (didApply) {
      hasChanges = true;
    }
  }

  if (style.fontFamily && editor.getAttributes?.("textStyle")?.fontFamily !== style.fontFamily) {
    if (callCommand(chain, "setFontFamily", style.fontFamily)) {
      hasChanges = true;
    }
  }

  if (style.fontSize && editor.getAttributes?.("textStyle")?.fontSize !== style.fontSize) {
    if (callCommand(chain, "setFontSize", style.fontSize)) {
      hasChanges = true;
    }
  }

  if (style.textAlign && editor.getAttributes?.("paragraph")?.textAlign !== style.textAlign) {
    if (callCommand(chain, "setTextAlign", style.textAlign)) {
      hasChanges = true;
    }
  }

  if (style.heading) {
    const didApply = callCommand(chain, "setHeading", {
      level: style.heading as 1 | 2 | 3 | 4 | 5 | 6,
    });

    if (didApply) {
      hasChanges = true;
    }
  } else if (style.bulletList) {
    applyIfInactive(
      true,
      !!editor.isActive?.("bulletList"),
      "toggleBulletList",
      "toggleBulletList",
    );
  } else if (style.orderedList) {
    applyIfInactive(
      true,
      !!editor.isActive?.("orderedList"),
      "toggleOrderedList",
      "toggleOrderedList",
    );
  } else if (style.blockquote) {
    applyIfInactive(
      true,
      !!editor.isActive?.("blockquote"),
      "toggleBlockquote",
      "toggleBlockquote",
    );
  }

  if (hasChanges) {
    callCommand(chain, "run");
  }
}
