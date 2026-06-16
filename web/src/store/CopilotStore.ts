import { create } from 'zustand';

interface CopilotState {
  isOpen: boolean;
  isPinned: boolean;
  context: unknown | null;
  promptToSubmit: string | null;
  setIsOpen: (isOpen: boolean) => void;
  setIsPinned: (isPinned: boolean) => void;
  setContext: (context: unknown) => void;
  setPromptToSubmit: (prompt: string | null) => void;
  toggle: () => void;
}

export const useCopilotStore = create<CopilotState>((set) => ({
  isOpen: false,
  isPinned: localStorage.getItem('copilotPinned') === 'true',
  context: null,
  promptToSubmit: null,
  setIsOpen: (isOpen) => set({ isOpen }),
  setIsPinned: (isPinned) => {
    localStorage.setItem('copilotPinned', String(isPinned));
    set({ isPinned });
  },
  setContext: (context) => set({ context }),
  setPromptToSubmit: (promptToSubmit) => set({ promptToSubmit }),
  toggle: () => set((state) => ({ isOpen: !state.isOpen })),
}));
