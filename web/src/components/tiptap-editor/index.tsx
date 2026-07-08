import "./editor.css";
import { getRouteApi } from "@tanstack/react-router";
import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight";
import Color from "@tiptap/extension-color";
import Highlight from "@tiptap/extension-highlight";
import Image from "@tiptap/extension-image";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { Table, TableCell, TableHeader, TableRow } from "@tiptap/extension-table";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import TextAlign from "@tiptap/extension-text-align";
import { TextStyle, TextStyleKit } from "@tiptap/extension-text-style";
import Typography from "@tiptap/extension-typography";
import Underline from "@tiptap/extension-underline";
import { Editor, EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { common, createLowlight } from "lowlight";
import { useCallback, useEffect, useRef, useState } from "react";
import { Markdown } from "tiptap-markdown";

import { documentApi } from "@/api/wiki/document";
import { ScrollArea } from "@/components/ui/scroll-area";

import { EditorCatalog } from "./_components/editor-catalog";
import { EditorFooter } from "./_components/editor-footer";
import { EditorToolbar } from "./_components/editor-toolbar";
import { type EditorPageSetup, useEditorPageSetup } from "./_hooks/use-editor-page-setup";

type SaveStatus = "saved" | "saving" | "unsaved";
const routeApi = getRouteApi("/_layout/wiki/tiptap-editor");

const lowlight = createLowlight(common);

interface TiptapEditorProps {
  content: string;
  onChange: (content: string) => void;
  onEditorReady: (editor: Editor | null) => void;
  pageSetup: EditorPageSetup;
}

function TiptapEditor({
  content,
  onChange,
  onEditorReady,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  pageSetup: _pageSetup,
}: TiptapEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        codeBlock: false,
      }),
      Markdown,
      Underline,
      TextStyle,
      TextStyleKit,
      Color,
      TextAlign.configure({
        types: ["heading", "paragraph"],
      }),
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Highlight.configure({
        multicolor: true,
      }),
      Typography,
      Image,
      Link.configure({
        openOnClick: false,
      }),
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
      CodeBlockLowlight.configure({
        lowlight,
      }),
      Placeholder.configure({
        placeholder: "开始输入内容，或按 / 插入块...",
      }),
    ],
    content: content,
    onUpdate: ({ editor: nextEditor }) => {
      const markdown = (nextEditor.storage as Record<string, any>).markdown?.getMarkdown?.();
      onChange(markdown ?? nextEditor.getHTML());
    },
    editorProps: {
      attributes: {
        class: "min-h-[500px] focus:outline-none",
      },
    },
    injectCSS: true,
  });

  useEffect(() => {
    onEditorReady(editor);
  }, [editor, onEditorReady]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    const currentMarkdown = (editor.storage as Record<string, any>).markdown?.getMarkdown?.();
    if (currentMarkdown !== content) {
      editor.commands.setContent(content, { emitUpdate: false });
    }
  }, [content, editor]);

  if (!editor) {
    return null;
  }

  return <EditorContent editor={editor} />;
}

export function CanvasTiptapPage() {
  const { docId } = routeApi.useSearch();

  const [content, setContent] = useState("");
  const [docTitle, setDocTitle] = useState("文档");
  const [catalogOpen, setCatalogOpen] = useState(true);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>("saved");
  const [editor, setEditor] = useState<Editor | null>(null);
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const pageSetup = useEditorPageSetup();

  // 加载文档内容
  useEffect(() => {
    if (!docId) {
      // eslint-disable-next-line
      setContent("");
      return;
    }

    let cancelled = false;

    const loadContent = async () => {
      const response = await documentApi.getContent(docId);
      console.log(
        "[Editor] loadContent: received content length=",
        response.content?.length,
        "first 100 chars:",
        response.content?.substring(0, 100),
      );
      if (!cancelled) {
        setContent(response.content);
        setDocTitle(response.filename?.replace(/\.[^.]+$/, "") || "文档");
        setSaveStatus("saved");
      }
    };

    void loadContent();

    return () => {
      cancelled = true;
    };
  }, [docId]);

  // 处理内容变化
  const handleContentChange = useCallback((nextContent: string) => {
    setContent(nextContent);
    setSaveStatus("unsaved");
  }, []);

  // 自动保存
  const handleSave = useCallback(async () => {
    if (!docId) {
      return;
    }

    setSaveStatus("saving");
    await documentApi.updateContent(docId, { content });
    setSaveStatus("saved");
  }, [content, docId]);

  // 自动保存逻辑
  useEffect(() => {
    if (saveStatus !== "unsaved" || !docId) {
      return;
    }

    saveTimeoutRef.current = setTimeout(() => {
      void handleSave();
    }, 2000);

    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, [docId, handleSave, saveStatus]);

  // 组件卸载时清除定时器
  useEffect(() => {
    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, []);

  // 文档字数
  const wordCount = content.split(/\s+/).filter(Boolean).length;

  return (
    <div className="flex h-screen flex-col bg-background">
      <EditorToolbar editor={editor} docTitle={docTitle} onSave={handleSave} />

      <div className="flex min-h-0 flex-1 overflow-hidden">
        <EditorCatalog
          open={catalogOpen}
          onToggle={() => setCatalogOpen((open) => !open)}
          editor={editor}
        />

        <ScrollArea className="flex-1 bg-[#f5f5f5] p-6">
          <div
            className="mx-auto bg-white shadow-sm"
            style={
              {
                width: `${pageSetup.paperSize.width}px`,
                zoom: pageSetup.zoom,
                paddingTop: `${pageSetup.margins.top}px`,
                paddingBottom: `${pageSetup.margins.bottom}px`,
                paddingLeft: `${pageSetup.margins.left}px`,
                paddingRight: `${pageSetup.margins.right}px`,
              } as React.CSSProperties
            }
          >
            <TiptapEditor
              content={content}
              onChange={handleContentChange}
              onEditorReady={setEditor}
              pageSetup={pageSetup}
            />
          </div>
        </ScrollArea>
      </div>

      <EditorFooter
        saveStatus={saveStatus}
        wordCount={wordCount}
        onSave={handleSave}
        pageSetup={pageSetup}
      />
    </div>
  );
}
