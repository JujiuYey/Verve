import type { Editor } from "@tiptap/react";
import { AlignCenter, AlignJustify, AlignLeft, AlignRight, List, ListOrdered } from "lucide-react";

import { useToolbarContext } from "../_context/toolbar-context";
import { ToolbarButton } from "../_primitives/toolbar-button";
import { ToolbarSelect } from "../_primitives/toolbar-select";

const HEADING_OPTIONS = [
  { value: "paragraph", label: "正文" },
  { value: "h1", label: "标题 1" },
  { value: "h2", label: "标题 2" },
  { value: "h3", label: "标题 3" },
  { value: "h4", label: "标题 4" },
];

const HEADING_LEVELS = {
  h1: 1,
  h2: 2,
  h3: 3,
  h4: 4,
} as const;

function getCurrentHeading(editor: Editor | null) {
  if (editor?.isActive?.("heading", { level: 1 })) return "h1";
  if (editor?.isActive?.("heading", { level: 2 })) return "h2";
  if (editor?.isActive?.("heading", { level: 3 })) return "h3";
  if (editor?.isActive?.("heading", { level: 4 })) return "h4";
  return "paragraph";
}

export function ParagraphGroup() {
  const { editor } = useToolbarContext();

  return (
    <>
      <ToolbarSelect
        ariaLabel="段落样式"
        value={getCurrentHeading(editor)}
        options={HEADING_OPTIONS}
        onValueChange={(value) => {
          if (!editor) {
            return;
          }

          if (value === "paragraph") {
            editor.chain().focus().setParagraph().run();
            return;
          }

          editor
            .chain()
            .focus()
            .toggleHeading({ level: HEADING_LEVELS[value as keyof typeof HEADING_LEVELS] })
            .run();
        }}
      />
      <ToolbarButton
        icon={<AlignLeft className="h-4 w-4" />}
        title="左对齐"
        isActive={!!editor?.isActive?.({ textAlign: "left" })}
        disabled={!editor}
        onClick={() => editor?.chain().focus().setTextAlign("left").run()}
      />
      <ToolbarButton
        icon={<AlignCenter className="h-4 w-4" />}
        title="居中对齐"
        isActive={!!editor?.isActive?.({ textAlign: "center" })}
        disabled={!editor}
        onClick={() => editor?.chain().focus().setTextAlign("center").run()}
      />
      <ToolbarButton
        icon={<AlignRight className="h-4 w-4" />}
        title="右对齐"
        isActive={!!editor?.isActive?.({ textAlign: "right" })}
        disabled={!editor}
        onClick={() => editor?.chain().focus().setTextAlign("right").run()}
      />
      <ToolbarButton
        icon={<AlignJustify className="h-4 w-4" />}
        title="两端对齐"
        isActive={!!editor?.isActive?.({ textAlign: "justify" })}
        disabled={!editor}
        onClick={() => editor?.chain().focus().setTextAlign("justify").run()}
      />
      <ToolbarButton
        icon={<List className="h-4 w-4" />}
        title="无序列表"
        isActive={!!editor?.isActive?.("bulletList")}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleBulletList().run()}
      />
      <ToolbarButton
        icon={<ListOrdered className="h-4 w-4" />}
        title="有序列表"
        isActive={!!editor?.isActive?.("orderedList")}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleOrderedList().run()}
      />
    </>
  );
}
