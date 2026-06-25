import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AppSettings {
  autoSave: boolean;
  theme: "system" | "light" | "dark";
}

interface AppState {
  settings: AppSettings;
  updateSettings: (settings: Partial<AppSettings>) => void;
  resetSettings: () => void;
}

const defaultSettings: AppSettings = {
  autoSave: true,
  theme: "system",
};

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      settings: { ...defaultSettings },
      updateSettings: (partial) =>
        set((state) => ({
          settings: { ...state.settings, ...partial },
        })),
      resetSettings: () => set({ settings: { ...defaultSettings } }),
    }),
    {
      name: "app-setting",
    },
  ),
);
