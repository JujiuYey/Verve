import { useEffect } from "react";

import { applyTheme, type ThemeMode } from "@/lib/theme";
import { useAppStore } from "@/stores/app";

function resolveTheme(theme: "system" | ThemeMode): ThemeMode {
  if (theme !== "system") {
    return theme;
  }

  if (typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }

  return "light";
}

export function useThemeSync() {
  const theme = useAppStore((s) => s.settings.theme);

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    const syncTheme = () => {
      applyTheme(resolveTheme(theme));
    };

    syncTheme();

    if (theme !== "system") {
      return;
    }

    media.addEventListener("change", syncTheme);
    return () => media.removeEventListener("change", syncTheme);
  }, [theme]);
}

export function initializeTheme() {
  applyTheme(resolveTheme(useAppStore.getState().settings.theme));
}
