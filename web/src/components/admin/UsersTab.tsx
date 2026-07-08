import React from 'react';
import { useTranslation } from 'react-i18next';
import type { User } from '../../hooks/useAdmin';

interface UsersTabProps {
  users: User[];
  onDelete?: (id: number) => void;
}

export const UsersTab: React.FC<UsersTabProps> = ({ users, onDelete }) => {
  const { t } = useTranslation('components');
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-surface-container-low p-6 border border-outline-variant relative overflow-hidden">
        <div className="absolute top-0 left-0 w-1 h-full bg-primary/40" />
        <div>
          <h3 className="text-label-caps text-on-surface-variant tracking-[0.3em] italic">
            {t('components.usersTab.title')}
          </h3>
          <p className="text-label-xs text-on-surface-variant opacity-40 mt-1 uppercase">
            {t('components.usersTab.description')}
          </p>
        </div>
        <div className="flex gap-4 items-center">
          <div className="text-right mr-4 hidden sm:block">
            <div className="text-[10px] font-black text-on-surface tracking-widest">
              {users.length}
            </div>
            <div className="text-[8px] text-on-surface-variant opacity-40 uppercase">
              {t('components.usersTab.totalIdentities')}
            </div>
          </div>
          <button className="btn-mechanical-primary px-6 py-2 text-[10px] font-black uppercase tracking-widest flex items-center gap-2">
            <span className="material-symbols-outlined text-[16px]">add</span>[ {t('components.usersTab.addOperator')} ]
          </button>
        </div>
      </div>

      <div className="cyber-panel flex flex-col overflow-hidden">
        <div className="cyber-grid-header flex items-center text-label-caps font-label-caps text-on-surface-variant tracking-widest shrink-0 border-b border-outline-variant bg-surface-container-low/50">
          <div className="flex-1 py-4 px-6">{t('components.usersTab.identityHandle')}</div>
          <div className="w-48 py-4 px-6 shrink-0">{t('components.usersTab.roleTier')}</div>
          <div className="w-40 py-4 px-6 shrink-0">{t('components.usersTab.status')}</div>
          <div className="w-48 py-4 px-6 shrink-0 text-right">{t('components.usersTab.operations')}</div>
        </div>

        <div className="divide-y divide-outline-variant/30">
          {users.map((user) => (
            <div
              key={user.id}
              className="cyber-grid-row flex items-center group hover:bg-white/5 transition-none"
            >
              <div className="flex-1 py-4 px-6 flex items-center gap-4">
                <div className="w-10 h-10 border border-outline-variant bg-surface-container-lowest flex items-center justify-center text-primary/40 group-hover:text-primary group-hover:border-primary/40 transition-none">
                  <span className="material-symbols-outlined text-[20px]">person</span>
                </div>
                <div>
                  <div className="text-mono-data font-mono-data text-on-surface group-hover:text-primary transition-none">
                    {user.username}
                  </div>
                  <div className="text-[9px] font-mono opacity-40 lowercase tracking-tight">
                    {user.email || t('components.usersTab.noEmail')}
                  </div>
                </div>
              </div>
              <div className="w-48 py-4 px-6 shrink-0">
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-primary/40" />
                  <span className="text-mono-data font-mono-data text-primary-fixed-dim uppercase tracking-tighter">
                    {user.global_role}
                  </span>
                </div>
              </div>
              <div className="w-40 py-4 px-6 shrink-0 flex items-center gap-2">
                <div
                  className={`skeuo-led ${user.is_active ? 'text-success bg-success' : 'text-error bg-error animate-pulse'}`}
                />
                <span
                  className={`text-label-caps font-label-caps tracking-widest text-[10px] ${
                    user.is_active ? 'text-success' : 'text-error'
                  }`}
                >
                  {user.is_active ? t('components.usersTab.active') : t('components.usersTab.suspended')}
                </span>
              </div>
              <div className="w-48 py-4 px-6 text-right shrink-0 flex justify-end gap-3 opacity-0 group-hover:opacity-100 transition-none">
                <button className="text-[10px] font-black text-on-surface-variant hover:text-primary transition-none tracking-widest">
                  [ {t('components.usersTab.manage')} ]
                </button>
                {!user.is_admin && user.global_role !== 'superadmin' && (
                  <button
                    onClick={() => onDelete?.(user.id)}
                    className="text-[10px] font-black text-on-surface-variant hover:text-error transition-none tracking-widest"
                  >
                    [ {t('components.usersTab.terminate')} ]
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
