import { Eraser, Lock, Paintbrush } from "lucide-react";

import { useToolbarContext } from "../_context/toolbar-context";
import { useFormatPainter } from "../_hooks/use-format-painter";
import { ToolbarButton } from "../_primitives/toolbar-button";

export function FormatGroup() {
  const { editor } = useToolbarContext();
  const { formatPainter, handleSingleClick, handleDoubleClick } = useFormatPainter(editor);

  return (
    <>
      <ToolbarButton
        icon={
          formatPainter.isLocked ? <Lock className="h-4 w-4" /> : <Paintbrush className="h-4 w-4" />
        }
        title={
          formatPainter.isLocked
            ? "格式刷(已锁定,点击或按ESC取消)"
            : formatPainter.isActive
              ? "格式刷(已激活,点击应用或按ESC取消)"
              : "格式刷"
        }
        isActive={formatPainter.isActive}
        disabled={!editor}
        onClick={handleSingleClick}
        onDoubleClick={handleDoubleClick}
      />
      <ToolbarButton
        icon={<Eraser />}
        title="清除格式"
        disabled={!editor}
        onClick={() => editor?.chain().focus().clearNodes().unsetAllMarks().run()}
      />
    </>
  );
}
