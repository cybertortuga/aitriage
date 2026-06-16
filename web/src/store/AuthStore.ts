import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  ok?: boolean;
  id: number;
  username: string;
  email?: string;
  full_name?: string;
  global_role: string;
  is_admin?: boolean;
  avatar_url?: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (user: User, token: string) => void;
  logout: () => void;
  hasRole: (roles: string[]) => boolean;
  setUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      login: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      hasRole: (roles) => {
        const user = get().user;
        if (!user) return false;
        return roles.includes(user.global_role);
      },
      setUser: (user) => set({ user, isAuthenticated: true }),
    }),
    {
      name: 'auth-storage', // name of the item in the storage (must be unique)
    },
  ),
);
