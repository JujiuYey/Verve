import { Editor } from "@tiptap/react";
import { ListTree } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CatalogItem {
  id: string;
  text: string;
  level: number;
  pos: number;
}

interface EditorCatalogProps {
  open: boolean;
  onToggle: () => void;
  editor: Editor | null;
}

function getCatalogItems(editor: EditorCatalogProps["editor"]): CatalogItem[] {
  const items: CatalogItem[] = [];
  let index = 0;

  editor?.state?.doc?.descendants?.((node, pos) => {
    if (node.type?.name === "heading") {
      items.push({
        id: `heading-${index}`,
        text: node.textContent || "无标题",
        level: Number(node.attrs?.level) || 1,
        pos,
      });
      index += 1;
    }
  });

  return items;
}

export function EditorCatalog({ open, onToggle, editor }: EditorCatalogProps) {
  const [items, setItems] = useState<CatalogItem[]>([]);

  useEffect(() => {
    if (!editor) {
      // eslint-disable-next-line
      setItems([]);
      return;
    }

    const syncItems = () => {
      setItems(getCatalogItems(editor));
    };

    syncItems();
    editor.on?.("update", syncItems);
    editor.on?.("selectionUpdate", syncItems);

    return () => {
      editor.off?.("update", syncItems);
      editor.off?.("selectionUpdate", syncItems);
    };
  }, [editor]);

  if (!open) {
    return null;
  }

  return (
    <aside className="w-64 shrink-0 border-r bg-white/80 backdrop-blur supports-backdrop-filter:bg-white/65 h-full flex flex-col">
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-2 text-sm font-medium text-slate-700">
          <ListTree className="h-4 w-4" />
          目录
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          aria-label="切换目录"
          onClick={onToggle}
        >
          <ListTree className="h-3.5 w-3.5" />
        </Button>
      </div>

      <ScrollArea className="min-h-0 flex-1 overflow-auto px-2 py-3">
        {items.length > 0 ? (
          items.map((item) => (
            <button
              key={item.id}
              type="button"
              className="flex w-full rounded-md px-2 py-1.5 text-left text-sm text-slate-700 transition hover:bg-slate-100"
              style={{ paddingLeft: `${(item.level - 1) * 8}px` }}
              onClick={() => editor?.chain?.().focus().setTextSelection(item.pos).run()}
            >
              {item.text}
            </button>
          ))
        ) : (
          <p className="px-2 text-xs text-slate-500">暂无目录</p>
        )}
      </ScrollArea>
    </aside>
  );
}
