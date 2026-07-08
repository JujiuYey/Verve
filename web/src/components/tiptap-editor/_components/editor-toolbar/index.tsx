import { Editor } from "@tiptap/react";

import { TooltipProvider } from "@/components/ui/tooltip";

import { ToolbarContext } from "./_context/toolbar-context";
import { ColorGroup } from "./_groups/color-group";
import { FormatGroup } from "./_groups/format-group";
import { HistoryGroup } from "./_groups/history-group";
import { InsertGroup } from "./_groups/insert-group";
import { ParagraphGroup } from "./_groups/paragraph-group";
import { SearchPrintGroup } from "./_groups/search-print-group";
import { TextStyleGroup } from "./_groups/text-style-group";
import { useEditorToolbarState } from "./_hooks/use-editor-toolbar-state";

interface EditorToolbarProps {
  editor: Editor | null;
  docTitle: string;
  onSave: () => void | Promise<void>;
}

export function EditorToolbar({ editor, docTitle, onSave: _onSave }: EditorToolbarProps) {
  const contextValue = useEditorToolbarState(editor, docTitle);

  return (
    <TooltipProvider>
      <ToolbarContext.Provider value={contextValue}>
        <div className="flex flex-wrap justify-center items-center gap-2 border-b bg-background px-4 py-2">
          <HistoryGroup />
          <div className="h-6 w-px bg-border" />
          <FormatGroup />
          <div className="h-6 w-px bg-border" />
          <TextStyleGroup />
          <div className="h-6 w-px bg-border" />
          <ColorGroup />
          <div className="h-6 w-px bg-border" />
          <ParagraphGroup />
          <div className="h-6 w-px bg-border" />
          <InsertGroup />
          <div className="h-6 w-px bg-border" />
          <SearchPrintGroup />
        </div>
      </ToolbarContext.Provider>
    </TooltipProvider>
  );
}
