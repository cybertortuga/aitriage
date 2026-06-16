import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ViewModeState {
  mode: 'simple' | 'advanced';
  setMode: (mode: 'simple' | 'advanced') => void;
  toggleMode: () => void;
}

export const useViewModeStore = create<ViewModeState>()(
  persist(
    (set) => ({
      mode: 'simple', // default to simple
      setMode: (mode) => set({ mode }),
      toggleMode: () => set((state) => ({ mode: state.mode === 'advanced' ? 'simple' : 'advanced' })),
    }),
    {
      name: 'aitriage-view-mode',
    }
  )
);
