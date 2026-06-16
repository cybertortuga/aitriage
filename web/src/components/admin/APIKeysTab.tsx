import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

export const APIKeysTab: React.FC = () => {
  const { t } = useTranslation('components');
  const [keys, setKeys] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchKeys = async () => {
    try {
      const res = await fetch('/api/admin/keys');
      const data = await res.json();
      if (data.ok) setKeys(data.keys || []);
    } catch (e) {
      console.error(t('APIKeysTab.failedToFetchKeys'), e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeys();
  }, []);

  const handleGenerate = async () => {
    const name = prompt(t('APIKeysTab.enterKeyName'));
    if (!name) return;

    try {
      const res = await fetch('/api/admin/keys/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      });
      const data = await res.json();
      if (data.ok) {
        // Use navigator.clipboard for a modern experience
        await navigator.clipboard.writeText(data.token);
        alert(`${t('APIKeysTab.keyGeneratedAlert')}\n\n${data.token}\n\n${t('APIKeysTab.tokenNotShownAgain')}`);
        fetchKeys();
      }
    } catch (e) {
      alert(t('components.apiKeysTab.failedToGenerateKey'));
    }
  };

  const handleRevoke = async (id: number) => {
    if (!confirm(t('APIKeysTab.confirmRevoke'))) return;

    try {
      const res = await fetch(`/api/admin/keys/${id}`, { method: 'DELETE' });
      const data = await res.json();
      if (data.ok) fetchKeys();
    } catch (e) {
      alert(t('components.apiKeysTab.failedToRevokeKey'));
    }
  };

  const copyToClipboard = async (text: string) => {
    await navigator.clipboard.writeText(text);
    // Simple visual feedback could be added here
  };

  return (
    <div className="flex flex-col gap-6 max-w-5xl">
      <div className="flex justify-between items-center bg-surface-container-low p-6 border border-outline-variant">
        <div>
          <h3 className="text-headline-sm font-headline-sm text-primary tracking-tighter">
            {t('APIKeysTab.title')}
          </h3>
          <p className="text-body-sm font-body-sm text-on-surface-variant opacity-60 max-w-md mt-1">
            {t('APIKeysTab.description')}
          </p>
        </div>
        <button onClick={handleGenerate} className="btn-primary h-10 px-6 flex items-center gap-2">
          <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
            add
          </span>
          {t('APIKeysTab.generateKey')}
        </button>
      </div>

      <div className="cyber-panel flex flex-col overflow-hidden">
        <div className="cyber-grid-header flex items-center text-label-caps font-label-caps text-on-surface-variant tracking-widest shrink-0 border-b border-outline-variant">
          <div className="flex-1 py-4 px-6">{t('APIKeysTab.tableHeader.keyName')}</div>
          <div className="w-64 py-4 px-6 shrink-0">{t('APIKeysTab.tableHeader.tokenPrefix')}</div>
          <div className="w-40 py-4 px-6 shrink-0">{t('APIKeysTab.tableHeader.createdAt')}</div>
          <div className="w-32 py-4 px-6 shrink-0">{t('APIKeysTab.tableHeader.status')}</div>
          <div className="w-32 py-4 px-6 shrink-0 text-right">{t('APIKeysTab.tableHeader.operations')}</div>
        </div>

        <div className="divide-y divide-outline-variant/30">
          {loading ? (
            <div className="py-20 text-center animate-pulse">
              <span className="text-label-caps text-on-surface-variant tracking-[0.3em] opacity-40">
                {t('APIKeysTab.syncingKeys')}
              </span>
            </div>
          ) : keys.length === 0 ? (
            <div className="py-20 text-center">
              <span className="text-label-caps text-on-surface-variant tracking-widest opacity-20">
                {t('APIKeysTab.noKeysFound')}
              </span>
            </div>
          ) : (
            keys.map((k, i) => (
              <div
                key={i}
                className="cyber-grid-row flex items-center group hover:bg-white/5 transition-none"
              >
                <div className="flex-1 py-4 px-6 text-mono-data font-mono-data text-on-surface flex items-center gap-3">
                  <div className="w-1.5 h-1.5 bg-primary/20" />
                  {k.name}
                </div>
                <div className="w-64 py-4 px-6 text-mono-data font-mono-data text-on-surface-variant shrink-0 flex items-center gap-2">
                  <code className="bg-surface-container/50 px-2 py-0.5 border border-outline-variant/30">
                    {k.prefix}••••••••
                  </code>
                  <button
                    onClick={() => copyToClipboard(k.prefix)}
                    className="opacity-0 group-hover:opacity-100 text-on-surface-variant hover:text-primary transition-none"
                    title={t('APIKeysTab.copyPrefix')}
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
                      content_copy
                    </span>
                  </button>
                </div>
                <div className="w-40 py-4 px-6 text-mono-data font-mono-data text-on-surface-variant shrink-0 opacity-60">
                  {new Date(k.created_at).toLocaleDateString()}
                </div>
                <div className="w-32 py-4 px-6 shrink-0 flex items-center gap-2">
                  <div
                    className={`skeuo-led ${k.status === 'ACTIVE' ? 'text-success bg-success' : 'text-error bg-error animate-pulse'}`}
                  />
                  <span
                    className={`text-label-caps font-label-caps tracking-widest text-[10px] ${k.status === 'ACTIVE' ? 'text-green-500' : 'text-error'}`}
                  >
                    {k.status}
                  </span>
                </div>
                <div className="w-32 py-4 px-6 text-right shrink-0 opacity-0 group-hover:opacity-100 transition-none">
                  {k.status === 'ACTIVE' && (
                    <button
                      onClick={() => handleRevoke(k.id)}
                      className="text-error hover:underline text-label-caps font-label-caps tracking-widest text-[10px]"
                    >
                      {t('APIKeysTab.revokeAccess')}
                    </button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="p-5 border border-outline-variant bg-surface-container-lowest/50 flex items-start gap-4">
        <span
          className="material-symbols-outlined text-primary mt-0.5"
          style={{ fontSize: '20px' }}
        >
          verified_user
        </span>
        <div>
          <p className="text-label-caps font-label-caps text-primary tracking-widest mb-1">
            {t('APIKeysTab.securityNotice.title')}
          </p>
          <p className="text-body-sm font-body-sm text-on-surface-variant leading-relaxed opacity-70">
            {t('APIKeysTab.securityNotice.descPart1')} <strong>{t('APIKeysTab.securityNotice.complianceLevel')}</strong>{t('APIKeysTab.securityNotice.descPart2')}
          </p>
        </div>
      </div>
    </div>
  );
};
