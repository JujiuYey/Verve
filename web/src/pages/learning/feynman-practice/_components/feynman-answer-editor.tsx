import "./feynman-answer-editor.css";
import Placeholder from "@tiptap/extension-placeholder";
import { EditorContent, useEditor, useEditorState, type Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import {
  BoldIcon,
  Code2Icon,
  CodeIcon,
  ItalicIcon,
  ListIcon,
  ListOrderedIcon,
  Trash2Icon,
} from "lucide-react";
import { useEffect } from "react";
import { Markdown } from "tiptap-markdown";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

type FeynmanAnswerEditorProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  compact?: boolean;
};

export function FeynmanAnswerEditor({
  value,
  onChange,
  placeholder = "把你的解释写在这里。可以插入代码块，也可以直接写卡住的地方。",
  disabled = false,
  compact = false,
}: FeynmanAnswerEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Markdown,
      Placeholder.configure({
        placeholder,
      }),
    ],
    content: value,
    editorProps: {
      attributes: {
        class: "px-4 py-3 text-sm leading-7",
      },
    },
    injectCSS: true,
    onUpdate: ({ editor: nextEditor }) => {
      onChange(getMarkdown(nextEditor));
    },
  });

  useEffect(() => {
    if (!editor) return;
    const currentMarkdown = getMarkdown(editor);
    if (currentMarkdown !== value) {
      editor.commands.setContent(value, { emitUpdate: false });
    }
  }, [editor, value]);

  useEffect(() => {
    editor?.setEditable(!disabled);
  }, [disabled, editor]);

  return (
    <div
      className={cn(
        "feynman-answer-editor flex shrink-0 flex-col overflow-hidden rounded-md border bg-background",
        compact ? "h-56 min-h-56" : "h-96 min-h-96",
        disabled && "cursor-not-allowed opacity-60",
      )}
    >
      <EditorToolbar editor={editor} disabled={disabled} onClear={() => onChange("")} />
      <div className="min-h-0 flex-1 overflow-auto">
        <EditorContent editor={editor} />
      </div>
    </div>
  );
}

function EditorToolbar({
  editor,
  disabled: readOnly,
  onClear,
}: {
  editor: Editor | null;
  disabled: boolean;
  onClear: () => void;
}) {
  const state = useEditorState({
    editor,
    selector: ({ editor }) => ({
      isBold: editor?.isActive("bold") ?? false,
      isItalic: editor?.isActive("italic") ?? false,
      isCode: editor?.isActive("code") ?? false,
      isCodeBlock: editor?.isActive("codeBlock") ?? false,
      isBulletList: editor?.isActive("bulletList") ?? false,
      isOrderedList: editor?.isActive("orderedList") ?? false,
      isEmpty: !editor || editor.isEmpty,
    }),
  }) ?? {
    isBold: false,
    isItalic: false,
    isCode: false,
    isCodeBlock: false,
    isBulletList: false,
    isOrderedList: false,
    isEmpty: true,
  };

  const disabled = !editor || readOnly;

  return (
    <div className="flex shrink-0 flex-wrap items-center gap-1 border-b bg-muted/20 p-2">
      <ToolbarIconButton
        label="加粗"
        active={state.isBold}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleBold().run()}
      >
        <BoldIcon />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="斜体"
        active={state.isItalic}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleItalic().run()}
      >
        <ItalicIcon />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="行内代码"
        active={state.isCode}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleCode().run()}
      >
        <CodeIcon />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="代码块"
        active={state.isCodeBlock}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
      >
        <Code2Icon />
      </ToolbarIconButton>
      <Separator orientation="vertical" className="mx-1 h-6" />
      <ToolbarIconButton
        label="无序列表"
        active={state.isBulletList}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleBulletList().run()}
      >
        <ListIcon />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="有序列表"
        active={state.isOrderedList}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleOrderedList().run()}
      >
        <ListOrderedIcon />
      </ToolbarIconButton>
      <div className="ml-auto flex items-center gap-2 pl-2">
        <ToolbarIconButton
          label="清空"
          disabled={disabled || state.isEmpty}
          onClick={() => {
            editor?.commands.clearContent();
            onClear();
          }}
        >
          <Trash2Icon />
        </ToolbarIconButton>
      </div>
    </div>
  );
}

function ToolbarIconButton({
  label,
  active,
  disabled,
  children,
  onClick,
}: {
  label: string;
  active?: boolean;
  disabled?: boolean;
  children: React.ReactNode;
  onClick: () => void;
}) {
  return (
    <Button
      type="button"
      variant={active ? "secondary" : "ghost"}
      size="icon"
      className={cn("size-8", active && "bg-background shadow-xs")}
      aria-label={label}
      title={label}
      disabled={disabled}
      onClick={onClick}
    >
      {children}
    </Button>
  );
}

function getMarkdown(editor: Editor) {
  const storage = editor.storage as unknown as {
    markdown?: { getMarkdown?: () => unknown };
  };
  const markdown = storage.markdown?.getMarkdown?.();
  return typeof markdown === "string" ? markdown : editor.getText();
}
