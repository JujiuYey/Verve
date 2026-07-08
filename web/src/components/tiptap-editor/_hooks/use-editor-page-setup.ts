import { useCallback, useState } from "react";

import type { PaperSize } from "../_shared/const";
import {
  DEFAULT_MARGINS,
  DEFAULT_PAPER_SIZE,
  parsePaperSize,
  ZOOM_DEFAULT,
  ZOOM_MAX,
  ZOOM_MIN,
  ZOOM_STEP,
} from "../_shared/const";

export interface PageMargins {
  top: number;
  bottom: number;
  left: number;
  right: number;
}

export interface EditorPageSetup {
  zoom: number;
  margins: PageMargins;
  paperSize: PaperSize;
  zoomIn: () => void;
  zoomOut: () => void;
  zoomReset: () => void;
  setZoom: (value: number) => void;
  setMargins: (margins: Partial<PageMargins>) => void;
  resetMargins: () => void;
  setPaperSize: (value: string) => void;
}

function clampZoom(v: number) {
  return Math.round(Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, v)) * 100) / 100;
}

export function useEditorPageSetup(): EditorPageSetup {
  const [zoom, setZoomRaw] = useState(ZOOM_DEFAULT);
  const [margins, setMarginsRaw] = useState<PageMargins>({ ...DEFAULT_MARGINS });
  const [paperSize, setPaperSizeRaw] = useState<PaperSize>({ ...DEFAULT_PAPER_SIZE });

  const zoomIn = useCallback(() => setZoomRaw((z) => clampZoom(z + ZOOM_STEP)), []);
  const zoomOut = useCallback(() => setZoomRaw((z) => clampZoom(z - ZOOM_STEP)), []);
  const zoomReset = useCallback(() => setZoomRaw(ZOOM_DEFAULT), []);
  const setZoom = useCallback((v: number) => setZoomRaw(clampZoom(v)), []);

  const setMargins = useCallback(
    (partial: Partial<PageMargins>) => setMarginsRaw((prev) => ({ ...prev, ...partial })),
    [],
  );
  const resetMargins = useCallback(() => setMarginsRaw({ ...DEFAULT_MARGINS }), []);

  const setPaperSize = useCallback((value: string) => {
    setPaperSizeRaw(parsePaperSize(value));
  }, []);

  return {
    zoom,
    margins,
    paperSize,
    zoomIn,
    zoomOut,
    zoomReset,
    setZoom,
    setMargins,
    resetMargins,
    setPaperSize,
  };
}
