import React, { useState, useEffect } from 'react';
import api from '../services/api';
import { formatDistanceToNow } from 'date-fns';
import { useTranslation } from 'react-i18next';

interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
  read: boolean;
  created_at: string;
}

export const NotificationDropdown: React.FC = () => {
  const { t } = useTranslation('components');
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchNotifications = async () => {
    setLoading(true);
    try {
      const res = await api.get('/notifications');
      if (res.data) {
        setNotifications(res.data);
      }
    } catch (err) {
      console.error('Failed to fetch notifications', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchNotifications();
    }
  }, [isOpen]);

  const markAsRead = async (id: string) => {
    try {
      await api.post(`/notifications/${id}/read`);
      setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)));
    } catch (err) {
      console.error('Failed to mark as read', err);
    }
  };

  const unreadCount = notifications.filter((n) => !n.read).length;

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="hover:bg-surface-container-high transition-none flex items-center justify-center h-10 w-10 opacity-80 hover:opacity-100 relative"
      >
        <span
          className="material-symbols-outlined"
          style={{
            fontVariationSettings: unreadCount > 0 ? "'FILL' 1" : "'FILL' 0",
            fontSize: '24px',
          }}
        >
          notifications
        </span>
        {unreadCount > 0 && (
          <span className="absolute top-2 right-2 w-2 h-2 bg-error border border-black" />
        )}
      </button>

      {isOpen && (
        <>
          <div className="fixed inset-0 z-[60]" onClick={() => setIsOpen(false)} />
          <div className="absolute right-0 mt-2 w-80 bg-surface-container border border-outline-variant z-[70]">
            <div className="px-4 py-3 border-b border-outline-variant flex justify-between items-center bg-surface-bright">
              <span className="text-label-caps font-label-caps text-primary tracking-widest text-[10px]">
                {t('components.notificationDropdown.operationalAlerts')}
              </span>
              {unreadCount > 0 && (
                <span className="text-[9px] font-mono-data text-error bg-error/10 px-2 py-0.5 border border-error/20">
                  {unreadCount} {t('components.notificationDropdown.new')}
                </span>
              )}
            </div>

            <div className="max-h-[400px] overflow-y-auto cyber-scrollbar">
              {loading && notifications.length === 0 ? (
                <div className="p-8 text-center animate-pulse">
                  <span className="text-label-caps font-label-caps text-on-surface-variant text-[10px] tracking-widest">
                    {t('components.notificationDropdown.synchronizing')}
                  </span>
                </div>
              ) : notifications.length === 0 ? (
                <div className="p-8 text-center opacity-30 italic">
                  <span className="text-label-caps font-label-caps text-on-surface-variant text-[10px] tracking-widest">
                    {t('components.notificationDropdown.noAlerts')}
                  </span>
                </div>
              ) : (
                <div className="flex flex-col">
                  {notifications.map((n) => (
                    <div
                      key={n.id}
                      className={`p-4 border-b border-outline-variant/30 hover:bg-surface-container-high transition-none cursor-pointer group ${!n.read ? 'bg-primary/5' : ''}`}
                      onClick={() => !n.read && markAsRead(n.id)}
                    >
                      <div className="flex justify-between items-start mb-1">
                        <span
                          className={`text-[10px] font-bold tracking-tight uppercase ${
                            n.type === 'error'
                              ? 'text-error'
                              : n.type === 'warning'
                                ? 'text-amber-500'
                                : n.type === 'success'
                                  ? 'text-green-500'
                                  : 'text-primary'
                          }`}
                        >
                          {n.title}
                        </span>
                        <span className="text-[9px] font-mono-data text-on-surface-variant opacity-40">
                          {formatDistanceToNow(new Date(n.created_at), { addSuffix: true })}
                        </span>
                      </div>
                      <p className="text-[11px] text-on-surface-variant leading-relaxed line-clamp-2">
                        {n.message}
                      </p>
                      {!n.read && (
                        <div className="mt-2 flex items-center gap-1 text-[8px] font-black text-primary opacity-0 group-hover:opacity-100 transition-none">
                          {t('components.notificationDropdown.markAsResolved')}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="p-2 bg-surface-bright border-t border-outline-variant">
              <button className="w-full py-2 text-[9px] font-black uppercase tracking-[0.3em] text-on-surface-variant hover:text-primary transition-none">
                {t('components.notificationDropdown.viewAllLogs')}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
