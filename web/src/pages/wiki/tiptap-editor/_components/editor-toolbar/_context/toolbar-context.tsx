import { createContext, useContext } from "react";

import type { ToolbarContextValue } from "../../../_shared/const";

export const ToolbarContext = createContext<ToolbarContextValue | null>(null);

export function useToolbarContext() {
  const context = useContext(ToolbarContext);

  if (!context) {
    throw new Error("useToolbarContext must be used within ToolbarContext.Provider");
  }

  return context;
}
