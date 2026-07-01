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
};

export function FeynmanAnswerEditor({
  value,
  onChange,
  placeholder = "把你的解释写在这里。可以插入代码块，也可以直接写卡住的地方。",
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

  return (
    <div className="feynman-answer-editor flex h-96 min-h-96 shrink-0 flex-col overflow-hidden rounded-md border bg-background">
      <EditorToolbar editor={editor} onClear={() => onChange("")} />
      <div className="min-h-0 flex-1 overflow-auto">
        <EditorContent editor={editor} />
      </div>
    </div>
  );
}

function EditorToolbar({ editor, onClear }: { editor: Editor | null; onClear: () => void }) {
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

  const disabled = !editor;

  return (
    <div className="flex shrink-0 flex-wrap items-center gap-1 border-b bg-muted/20 p-2">
      <ToolbarIconButton
        label="加粗"
        active={state.isBold}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleBold().run()}
      >
        <BoldIcon className="size-4" />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="斜体"
        active={state.isItalic}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleItalic().run()}
      >
        <ItalicIcon className="size-4" />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="行内代码"
        active={state.isCode}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleCode().run()}
      >
        <CodeIcon className="size-4" />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="代码块"
        active={state.isCodeBlock}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
      >
        <Code2Icon className="size-4" />
      </ToolbarIconButton>
      <Separator orientation="vertical" className="mx-1 h-6" />
      <ToolbarIconButton
        label="无序列表"
        active={state.isBulletList}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleBulletList().run()}
      >
        <ListIcon className="size-4" />
      </ToolbarIconButton>
      <ToolbarIconButton
        label="有序列表"
        active={state.isOrderedList}
        disabled={disabled}
        onClick={() => editor?.chain().focus().toggleOrderedList().run()}
      >
        <ListOrderedIcon className="size-4" />
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
          <Trash2Icon className="size-4" />
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
  const markdown = (editor.storage as Record<string, any>).markdown?.getMarkdown?.();
  return typeof markdown === "string" ? markdown : editor.getText();
}
