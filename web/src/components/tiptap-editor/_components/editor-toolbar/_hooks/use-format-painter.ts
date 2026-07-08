import type { Editor } from "@tiptap/react";
import { type MouseEvent as ReactMouseEvent, useCallback, useEffect, useState } from "react";

import type { RangeStyle } from "../../../_shared/const";
import { applyRangeStyle, readRangeStyle } from "../../../_shared/range-style";

export interface FormatPainterState {
  isActive: boolean;
  isLocked: boolean;
  savedStyle: Partial<RangeStyle> | null;
}

const DEFAULT_FORMAT_PAINTER: FormatPainterState = {
  isActive: false,
  isLocked: false,
  savedStyle: null,
};

export function useFormatPainter(editor: Editor | null) {
  const [formatPainter, setFormatPainter] = useState<FormatPainterState>(DEFAULT_FORMAT_PAINTER);

  const captureStyle = useCallback(() => {
    if (!editor) {
      return;
    }

    setFormatPainter((prev) => ({
      ...prev,
      savedStyle: readRangeStyle(editor),
    }));
  }, [editor]);

  const applyCapturedStyle = useCallback(() => {
    if (!formatPainter.savedStyle) {
      return;
    }

    applyRangeStyle(editor, formatPainter.savedStyle);
  }, [editor, formatPainter.savedStyle]);

  const activateFormatPainter = useCallback(() => {
    setFormatPainter((prev) => ({
      ...prev,
      isActive: true,
    }));
  }, []);

  const deactivateFormatPainter = useCallback(() => {
    setFormatPainter((prev) => ({
      ...prev,
      isActive: false,
      isLocked: false,
    }));
  }, []);

  const toggleFormatPainterLock = useCallback(() => {
    setFormatPainter((prev) => ({
      ...prev,
      isLocked: !prev.isLocked,
    }));
  }, []);

  const handleEditorClick = useCallback(() => {
    if (!formatPainter.isActive || !formatPainter.savedStyle) {
      return;
    }

    applyCapturedStyle();

    if (!formatPainter.isLocked) {
      deactivateFormatPainter();
    }
  }, [
    applyCapturedStyle,
    deactivateFormatPainter,
    formatPainter.isActive,
    formatPainter.isLocked,
    formatPainter.savedStyle,
  ]);

  useEffect(() => {
    if (!editor || !formatPainter.isActive) {
      return;
    }

    const editorElement = editor.view?.dom;
    if (!editorElement) {
      return;
    }

    editorElement.addEventListener("click", handleEditorClick);

    return () => {
      editorElement.removeEventListener("click", handleEditorClick);
    };
  }, [editor, formatPainter.isActive, handleEditorClick]);

  useEffect(() => {
    if (!formatPainter.isActive) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        deactivateFormatPainter();
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [deactivateFormatPainter, formatPainter.isActive]);

  const handleSingleClick = useCallback(() => {
    if (formatPainter.isActive) {
      deactivateFormatPainter();
      return;
    }

    captureStyle();
    activateFormatPainter();
  }, [activateFormatPainter, captureStyle, deactivateFormatPainter, formatPainter.isActive]);

  const handleDoubleClick = useCallback(
    (event: ReactMouseEvent) => {
      event.preventDefault();

      if (!formatPainter.isActive) {
        captureStyle();
        activateFormatPainter();
      }

      toggleFormatPainterLock();
    },
    [activateFormatPainter, captureStyle, formatPainter.isActive, toggleFormatPainterLock],
  );

  return {
    formatPainter,
    handleSingleClick,
    handleDoubleClick,
  };
}
