import { useEffect, type ReactNode } from "react";

import { initializeTheme, useThemeSync } from "@/hooks/use-theme";

export function ThemeProvider({ children }: { children: ReactNode }) {
  useEffect(() => {
    initializeTheme();
  }, []);

  useThemeSync();

  return <>{children}</>;
}
