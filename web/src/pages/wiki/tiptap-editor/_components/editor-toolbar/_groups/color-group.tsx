import { Highlighter, Type } from "lucide-react";

import { HIGHLIGHT_COLORS, TEXT_COLORS } from "@/pages/wiki/tiptap-editor/_shared/const";

import { useToolbarContext } from "../_context/toolbar-context";
import { ColorPickerPopover } from "../_primitives/color-picker-popover";
import { ToolbarButton } from "../_primitives/toolbar-button";

export function ColorGroup() {
  const { editor, rangeStyle } = useToolbarContext();

  return (
    <>
      <ColorPickerPopover
        color={rangeStyle.textColor}
        colors={TEXT_COLORS}
        onChange={(color) => editor?.chain().focus().setColor(color).run()}
      >
        <div>
          <ToolbarButton icon={<Type className="h-4 w-4" />} title="文字颜色" disabled={!editor} />
        </div>
      </ColorPickerPopover>

      <ColorPickerPopover
        color={rangeStyle.highlight}
        colors={HIGHLIGHT_COLORS}
        onChange={(color) => editor?.chain().focus().toggleHighlight({ color }).run()}
      >
        <div>
          <ToolbarButton
            icon={<Highlighter className="h-4 w-4" />}
            title="高亮颜色"
            disabled={!editor}
          />
        </div>
      </ColorPickerPopover>
    </>
  );
}
