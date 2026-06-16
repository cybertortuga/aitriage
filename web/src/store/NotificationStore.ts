import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import i18n from '../i18n';

export type NotificationSeverity = 'critical' | 'warning' | 'info';

export interface AppNotification {
  id: string;
  timestamp: string;
  title: string;
  message: string;
  severity: NotificationSeverity;
  read: boolean;
  link?: string;
}

interface NotificationState {
  notifications: AppNotification[];
  unreadCount: number;
  addNotification: (notification: Omit<AppNotification, 'id' | 'timestamp' | 'read'>) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationState>()(
  persist(
    (set) => ({
      notifications: [
        {
          id: '1',
          timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
          title: i18n.t('notifications.slaBreachTitle'),
          message: i18n.t('notifications.slaBreachMessage'),
          severity: 'critical',
          read: false,
          link: '/products/finance-core',
        },
        {
          id: '2',
          timestamp: new Date(Date.now() - 1000 * 60 * 120).toISOString(),
          title: i18n.t('notifications.scanCompletedTitle'),
          message: i18n.t('notifications.scanCompletedMessage'),
          severity: 'info',
          read: true,
          link: '/products/edge-gateway',
        },
      ],
      unreadCount: 1,
      addNotification: (notification) => {
        const newNotification: AppNotification = {
          ...notification,
          id: Math.random().toString(36).substring(7),
          timestamp: new Date().toISOString(),
          read: false,
        };
        set((state) => ({
          notifications: [newNotification, ...state.notifications],
          unreadCount: state.unreadCount + 1,
        }));
      },
      markAsRead: (id) => {
        set((state) => {
          const notification = state.notifications.find((n) => n.id === id);
          if (notification && !notification.read) {
            return {
              notifications: state.notifications.map((n) =>
                n.id === id ? { ...n, read: true } : n,
              ),
              unreadCount: state.unreadCount - 1,
            };
          }
          return state;
        });
      },
      markAllAsRead: () => {
        set((state) => ({
          notifications: state.notifications.map((n) => ({ ...n, read: true })),
          unreadCount: 0,
        }));
      },
      clearAll: () => {
        set({ notifications: [], unreadCount: 0 });
      },
    }),
    {
      name: 'aitriage-notifications',
    },
  ),
);
