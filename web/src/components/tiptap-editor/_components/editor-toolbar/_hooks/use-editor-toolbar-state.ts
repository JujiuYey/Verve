import type { Editor } from "@tiptap/react";
import { useEffect, useMemo, useState } from "react";

import type { ToolbarContextValue } from "../../../_shared/const";
import { DEFAULT_RANGE_STYLE } from "../../../_shared/const";
import { readRangeStyle } from "../../../_shared/range-style";

export function useEditorToolbarState(
  editor: Editor | null,
  docTitle: string,
): ToolbarContextValue {
  const [styleVersion, setStyleVersion] = useState(0);

  const rangeStyle = useMemo(() => {
    void styleVersion;
    return editor ? readRangeStyle(editor) : DEFAULT_RANGE_STYLE;
  }, [editor, styleVersion]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    const updateStyle = () => {
      setStyleVersion((current) => current + 1);
    };

    editor.on?.("selectionUpdate", updateStyle);
    editor.on?.("transaction", updateStyle);

    return () => {
      editor.off?.("selectionUpdate", updateStyle);
      editor.off?.("transaction", updateStyle);
    };
  }, [editor]);

  return {
    editor,
    rangeStyle,
    docTitle,
  };
}
