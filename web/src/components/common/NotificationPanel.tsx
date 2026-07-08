import React from 'react';
import { useNotificationStore } from '../../store/NotificationStore';
import type { AppNotification } from '../../store/NotificationStore';
import { formatDistanceToNow } from 'date-fns';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';

interface NotificationPanelProps {
  onClose: () => void;
}

export const NotificationPanel: React.FC<NotificationPanelProps> = ({ onClose }) => {
  const { notifications, markAsRead, markAllAsRead, clearAll } = useNotificationStore();
  const navigate = useNavigate();
  const { t } = useTranslation('components');

  const handleNotificationClick = (n: AppNotification) => {
    markAsRead(n.id);
    if (n.link) navigate(n.link);
    onClose();
  };

  const getSeverityStyles = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'border-l-4 border-severity-critical bg-severity-critical/5';
      case 'warning':
        return 'border-l-4 border-severity-high bg-severity-high/5';
      default:
        return 'border-l-4 border-primary-fixed-dim bg-primary-fixed-dim/5';
    }
  };

  return (
    <motion.div 
      initial={{ opacity: 0, y: -20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, scale: 0.95 }}
      transition={{ type: "spring", stiffness: 300, damping: 25 }}
      className="fixed top-16 right-8 w-96 bg-surface-container-high border border-outline-variant z-[100] flex flex-col font-code"
    >
      <div className="p-4 border-b border-outline-variant flex justify-between items-center bg-surface-container-highest">
        <h3 className="text-[10px] font-black uppercase tracking-[0.2em] italic">
          :: {t('NotificationPanel.title', 'NOTIFICATION_FEED')}
        </h3>
        <div className="flex gap-4">
          <button
            onClick={markAllAsRead}
            className="text-[9px] font-black uppercase opacity-40 hover:opacity-100 transition-none"
          >
            {t('NotificationPanel.ackAll', 'ACK_ALL')}
          </button>
          <button
            onClick={onClose}
            className="text-[10px] font-black uppercase opacity-60 hover:text-error"
          >
            ✕
          </button>
        </div>
      </div>

      <div className="max-h-[400px] overflow-y-auto cyber-scrollbar">
        {notifications.length === 0 ? (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="p-8 text-center text-[10px] font-black uppercase opacity-20 italic"
          >
            {t('NotificationPanel.noPending', 'NO_PENDING_ALERTS')}
          </motion.div>
        ) : (
          <AnimatePresence>
            {notifications.map((n) => (
              <motion.div
                key={n.id}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                layout
                onClick={() => handleNotificationClick(n)}
                className={`p-4 border-b border-outline-variant/30 cursor-pointer group transition-all duration-300 hover:bg-surface-container-highest ${getSeverityStyles(n.severity)} ${!n.read ? 'opacity-100' : 'opacity-50'}`}
              >
                <div className="flex justify-between items-start mb-1">
                  <span
                    className={`text-[9px] font-black uppercase tracking-widest ${
                      n.severity === 'critical'
                        ? 'text-severity-critical'
                        : n.severity === 'warning'
                          ? 'text-severity-high'
                          : 'text-primary-fixed-dim'
                    }`}
                  >
                    {n.title}
                  </span>
                  <span className="text-[8px] opacity-40 font-mono">
                    {formatDistanceToNow(new Date(n.timestamp))} {t('NotificationPanel.ago', 'ago')}
                  </span>
                </div>
                <p className="text-[10px] leading-relaxed text-on-surface mb-2 font-medium">
                  {n.message}
                </p>
                {!n.read && (
                  <div className="flex items-center gap-2">
                    <div className="w-1.5 h-1.5 bg-primary-fixed-dim animate-pulse" />
                    <span className="text-[8px] font-black uppercase tracking-tighter text-primary-fixed-dim">
                      {t('NotificationPanel.newLog', 'NEW_LOG_ENTRY')}
                    </span>
                  </div>
                )}
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>

      <div className="p-3 border-t border-outline-variant bg-surface-container-low flex justify-center">
        <button
          onClick={clearAll}
          className="text-[9px] font-black uppercase opacity-30 hover:opacity-100 hover:text-error transition-colors"
        >
          // {t('NotificationPanel.purgeHistory', 'PURGE_HISTORY')}
        </button>
      </div>
    </motion.div>
  );
};
