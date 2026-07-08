import { useEditorState } from "@tiptap/react";
import { Bold, Code, Italic, Strikethrough, Underline } from "lucide-react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { FONT_FAMILIES, FONT_SIZES } from "../../../_shared/const";
import { useToolbarContext } from "../_context/toolbar-context";
import { ToolbarButton } from "../_primitives/toolbar-button";

export function TextStyleGroup() {
  const { editor } = useToolbarContext();

  const { currentFont, currentSize, bold, italic, underline, strikethrough, inlineCode } =
    useEditorState({
      editor,
      selector: (ctx) => {
        const currentFont = FONT_FAMILIES.find((f) =>
          ctx.editor?.isActive("textStyle", { fontFamily: f.value }),
        );
        const currentSize = FONT_SIZES.find((s) =>
          ctx.editor?.isActive("textStyle", { fontSize: `${s.value}px` }),
        );
        const bold = ctx.editor?.isActive("bold") ?? false;
        const italic = ctx.editor?.isActive("italic") ?? false;
        const underline = ctx.editor?.isActive("underline") ?? false;
        const strikethrough = ctx.editor?.isActive("strike") ?? false;
        const inlineCode = ctx.editor?.isActive("code") ?? false;
        return {
          currentFont: currentFont?.value ?? "Microsoft YaHei",
          currentSize: currentSize?.value ?? "16",
          bold,
          italic,
          underline,
          strikethrough,
          inlineCode: inlineCode,
        };
      },
    }) ?? {
      currentFont: "Microsoft YaHei",
      currentSize: "16",
      bold: false,
      italic: false,
      underline: false,
      strikethrough: false,
      inlineCode: false,
    };

  return (
    <>
      <Select
        value={currentFont!}
        onValueChange={(value) => {
          editor?.chain().focus().setFontFamily(value).run();
        }}
        disabled={!editor}
      >
        <SelectTrigger className="w-20 text-xs border-none bg-transparent shadow-none focus:ring-0 px-1">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {FONT_FAMILIES.map((f) => (
            <SelectItem key={f.value} value={f.value} style={{ fontFamily: f.value }}>
              {f.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={currentSize!}
        onValueChange={(value) => {
          editor?.chain().focus().setFontSize(`${value}px`).run();
        }}
        disabled={!editor}
      >
        <SelectTrigger className="w-16 text-xs border-none bg-transparent shadow-none focus:ring-0 px-1">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {FONT_SIZES.map((s) => (
            <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <ToolbarButton
        icon={
          <span className="text-xs">
            A<sup>+</sup>
          </span>
        }
        title="增大字号"
        disabled={!editor}
        onClick={() => {
          const idx = FONT_SIZES.findIndex((s) => s.value === currentSize);
          const prev = FONT_SIZES[idx - 1];
          if (prev) editor?.chain().focus().setFontSize(`${prev.value}px`).run();
        }}
      />
      <ToolbarButton
        icon={
          <span className="text-xs">
            A<sup>-</sup>
          </span>
        }
        title="减小字号"
        disabled={!editor}
        onClick={() => {
          const idx = FONT_SIZES.findIndex((s) => s.value === currentSize);
          const next = FONT_SIZES[idx + 1];
          if (next) editor?.chain().focus().setFontSize(`${next.value}px`).run();
        }}
      />

      <ToolbarButton
        icon={<Bold className="h-4 w-4" />}
        title="粗体"
        isActive={bold}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleBold().run()}
      />
      <ToolbarButton
        icon={<Italic className="h-4 w-4" />}
        title="斜体"
        isActive={italic}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleItalic().run()}
      />
      <ToolbarButton
        icon={<Underline className="h-4 w-4" />}
        title="下划线"
        isActive={underline}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleUnderline().run()}
      />
      <ToolbarButton
        icon={<Strikethrough className="h-4 w-4" />}
        title="删除线"
        isActive={strikethrough}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleStrike().run()}
      />
      <ToolbarButton
        icon={<Code className="h-4 w-4" />}
        title="行内代码"
        isActive={inlineCode}
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleCode().run()}
      />
    </>
  );
}
