import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { UsersTab } from '../components/admin/UsersTab';
import { AuditLogTab } from '../components/admin/AuditLogTab';
import { APIKeysTab } from '../components/admin/APIKeysTab';
import { SystemConfigTab } from '../components/admin/SystemConfigTab';
import { useAdmin } from '../hooks/useAdmin';
import { LoadingScreen } from '../components/common/LoadingScreen';
import { useTitle } from '../hooks/useTitle';

export const AdminPanelPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle('Admin');
  const [activeTab, setActiveTab] = useState<'USERS' | 'AUDIT_LOGS' | 'SYSTEM_CONFIG' | 'API_KEYS'>(
    'USERS',
  );
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('viewer');
  const [inviteStatus, setInviteStatus] = useState<'IDLE' | 'SENDING' | 'SUCCESS' | 'ERROR'>(
    'IDLE',
  );
  const [errorMessage, setErrorMessage] = useState('');

  const { users, auditLogs, loading, deleteUser, createUser } = useAdmin();

  const handleInvite = async () => {
    if (!inviteEmail) return;
    setInviteStatus('SENDING');

    // Generate a temporary password (enterprise flow)
    const tempPassword = Math.random().toString(36).slice(-8);
    const username = inviteEmail.split('@')[0];

    const res = await createUser({
      username: username,
      email: inviteEmail,
      password: tempPassword,
      global_role: inviteRole.toLowerCase(),
      is_active: true,
    });

    if (res.ok) {
      setInviteStatus('SUCCESS');
      setTimeout(() => {
        setShowInviteModal(false);
        setInviteStatus('IDLE');
        setInviteEmail('');
      }, 1500);
    } else {
      setInviteStatus('ERROR');
      setErrorMessage(res.error || t('admin.failedToCreateUser'));
    }
  };

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="px-4 py-2 flex justify-between items-center flex-shrink-0 cyber-header-premium relative z-10 border-b border-outline-variant/30">
        <div className="flex items-center gap-4">
          <div className="hidden md:flex w-8 h-8 border border-outline-variant items-center justify-center bg-surface-container-low">
            <span
              className="material-symbols-outlined text-primary/60"
              style={{ fontSize: '16px' }}
            >
              admin_panel_settings
            </span>
          </div>
          <div>
            <div className="flex items-center gap-2 mb-0.5">
              <span className="text-[9px] font-bold tracking-widest text-on-surface-variant opacity-60">
                {t('admin.rootSettings')}
              </span>
              <span className="text-[9px] font-bold tracking-widest text-on-surface-variant opacity-20">
                /
              </span>
              <span className="text-[9px] font-bold tracking-widest text-on-surface-variant opacity-60">
                {t('admin.systemAdministration')}
              </span>
            </div>
            <h1 className="text-title-lg font-bold tracking-tight text-primary uppercase">
              {t('admin.managementConsole')}
            </h1>
          </div>
        </div>
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2">
            <div className="w-1.5 h-1.5 bg-success animate-pulse" />
            <span className="text-[9px] font-bold tracking-widest text-success">
              {t('admin.authServerUp')}
            </span>
          </div>
          <button
            onClick={() => setShowInviteModal(true)}
            className="btn-primary h-8 px-4 flex items-center gap-2 relative group overflow-hidden"
          >
            <div className="absolute inset-0 bg-white/10 translate-x-[-100%] group-hover:translate-x-[100%] transition-none -out" />
            <span className="material-symbols-outlined text-[16px]">person_add</span>
            <span className="tracking-[0.2em]">{t('admin.inviteOperator')}</span>
          </button>
        </div>
      </div>

      {/* Tab Bar */}
      <div className="flex border-b border-outline-variant shrink-0 bg-surface-container-lowest/50 px-4">
        {(['USERS', 'AUDIT_LOGS', 'SYSTEM_CONFIG', 'API_KEYS'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`h-12 px-8 text-label-caps tracking-[0.2em] transition-none relative flex items-center justify-center group ${
              activeTab === tab
                ? 'text-primary'
                : 'text-on-surface-variant opacity-40 hover:opacity-100 hover:bg-white/5'
            }`}
          >
            {t(`admin.tabs.${tab}`)}
            {activeTab === tab && (
              <div className="absolute bottom-0 left-0 w-full h-[2px] bg-primary" />
            )}
            <div
              className={`absolute top-0 left-0 w-[1px] h-full bg-outline-variant/30 ${activeTab === tab ? 'opacity-100' : 'opacity-0'}`}
            />
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto cyber-scrollbar">
        {loading ? (
          <LoadingScreen />
        ) : (
          <div className="p-8">
            {activeTab === 'USERS' && <UsersTab users={users} onDelete={deleteUser} />}
            {activeTab === 'AUDIT_LOGS' && <AuditLogTab logs={auditLogs} />}
            {activeTab === 'SYSTEM_CONFIG' && <SystemConfigTab />}
            {activeTab === 'API_KEYS' && <APIKeysTab />}
          </div>
        )}
      </div>

      {/* Invite User Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/95">
          <div className="cyber-panel w-full max-w-md border-primary flex flex-col animate-in fade-in zoom-in ">
            <div className="px-6 py-4 border-b border-outline-variant flex justify-between items-center bg-surface-container-highest">
              <h2 className="text-label-caps text-primary">{t('admin.inviteNewOperator')}</h2>
              <button
                onClick={() => setShowInviteModal(false)}
                className="text-on-surface-variant hover:text-primary"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>

            {inviteStatus === 'SUCCESS' ? (
              <div className="p-12 flex flex-col items-center justify-center text-center gap-4">
                <div className="w-12 h-12 bg-success flex items-center justify-center">
                  <span className="material-symbols-outlined text-black font-bold">check</span>
                </div>
                <div>
                  <h3 className="text-headline-sm text-primary uppercase">{t('admin.invitationSuccessful')}</h3>
                  <p className="text-label-caps text-on-surface-variant mt-2">
                    {t('admin.userHasBeenProvisioned')}
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-6 space-y-4">
                {inviteStatus === 'ERROR' && (
                  <div className="p-3 border border-error bg-error/5 flex items-center gap-3">
                    <div className="w-1.5 h-1.5 bg-error shrink-0" />
                    <p className="text-label-caps text-error">{errorMessage.toUpperCase()}</p>
                  </div>
                )}

                <div>
                  <label className="text-label-caps text-on-surface-variant mb-2 block">
                    {t('admin.emailAddress')}
                  </label>
                  <input
                    type="email"
                    className="cyber-input w-full p-2 text-mono-data"
                    placeholder="operator@enterprise.com"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    autoFocus
                  />
                </div>
                <div>
                  <label className="text-label-caps text-on-surface-variant mb-2 block">
                    {t('admin.assignedRole')}
                  </label>
                  <select
                    className="cyber-input w-full p-2 text-mono-data bg-surface-container-lowest"
                    value={inviteRole}
                    onChange={(e) => setInviteRole(e.target.value)}
                  >
                    <option value="viewer">{t('admin.viewer')}</option>
                    <option value="manager">{t('admin.manager')}</option>
                    <option value="admin">{t('admin.administrator')}</option>
                  </select>
                </div>
                <div className="pt-4 flex gap-3">
                  <button
                    onClick={() => setShowInviteModal(false)}
                    className="btn-secondary flex-1"
                    disabled={inviteStatus === 'SENDING'}
                  >
                    {t('admin.cancel')}
                  </button>
                  <button
                    onClick={handleInvite}
                    className="btn-primary flex-1 flex items-center justify-center gap-2"
                    disabled={inviteStatus === 'SENDING' || !inviteEmail}
                  >
                    {inviteStatus === 'SENDING' ? (
                      <>
                        <div className="w-3 h-3 border-2 border-black/20 border-t-black animate-spin" />
                        {t('admin.processing')}
                      </>
                    ) : (
                      t('admin.sendInvitation')
                    )}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
