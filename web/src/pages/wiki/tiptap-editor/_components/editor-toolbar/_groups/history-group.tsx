import { Redo2, Undo2 } from "lucide-react";

import { useToolbarContext } from "../_context/toolbar-context";
import { ToolbarButton } from "../_primitives/toolbar-button";

export function HistoryGroup() {
  const { editor } = useToolbarContext();

  return (
    <>
      <ToolbarButton
        icon={<Undo2 className="h-4 w-4" />}
        title="撤销"
        disabled={!editor}
        onClick={() => editor?.chain().focus().undo().run()}
      />
      <ToolbarButton
        icon={<Redo2 className="h-4 w-4" />}
        title="重做"
        disabled={!editor}
        onClick={() => editor?.chain().focus().redo().run()}
      />
    </>
  );
}
