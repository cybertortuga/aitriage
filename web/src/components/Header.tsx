import React from 'react';
import { useCopilotStore } from '../store/CopilotStore';
import { useViewModeStore } from '../store/ViewModeStore';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import api from '../services/api';
import Modal from './common/Modal';

const PALETTES = [
  { id: 'violet', name: 'Obsidian Violet', hex: '#8b5cf6' },
  { id: 'cyan', name: 'Aurora Cyan', hex: '#06b6d4' },
  { id: 'emerald', name: 'Emerald Secure', hex: '#10b981' },
  { id: 'amber', name: 'Magma Orange', hex: '#f97316' },
  { id: 'rose', name: 'Sakura Pink', hex: '#f472b6' },
  { id: 'crimson', name: 'Bordeaux Red', hex: '#e11d48' },
  { id: 'cobalt', name: 'Royal Cobalt', hex: '#2563eb' },
  { id: 'bronze', name: 'Imperial Gold', hex: '#eab308' },
  { id: 'teal', name: 'Acid Lime', hex: '#84cc16' },
  { id: 'white', name: 'Quartz White', hex: '#f8fafc' },
];

export const Header: React.FC = () => {
  const [scrolled, setScrolled] = React.useState(false);
  const [showSettings, setShowSettings] = React.useState(false);
  const [settingsTab, setSettingsTab] = React.useState('theme');
  const [enableAnalytics, setEnableAnalytics] = React.useState(() => localStorage.getItem('aitriage_analytics') !== 'false');
  const [autoScan, setAutoScan] = React.useState(() => localStorage.getItem('aitriage_autoscan') === 'true');
  const { t, i18n } = useTranslation();
  
  const [showRebuildModal, setShowRebuildModal] = React.useState(false);
  const [rebuildStatus, setRebuildStatus] = React.useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  
  const { mode, toggleMode } = useViewModeStore();
  const isCopilotOpen = useCopilotStore((state) => state.isOpen);
  const [accent, setAccent] = React.useState(() => {
    try {
      return localStorage.getItem('aitriage_accent') || 'white';
    } catch {
      return 'white';
    }
  });

  React.useEffect(() => {
    try {
      localStorage.setItem('aitriage_accent', accent);
    } catch {}
    document.documentElement.setAttribute('data-accent', accent);
  }, [accent]);

  React.useEffect(() => {
    const handleAccentChange = () => {
      try {
        setAccent(localStorage.getItem('aitriage_accent') || 'white');
      } catch {}
    };
    window.addEventListener('aitriage_accent_change', handleAccentChange);
    return () => window.removeEventListener('aitriage_accent_change', handleAccentChange);
  }, []);

  React.useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 16);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <header
      className={`sticky top-0 z-50 transition-all duration-400 ease-out border-b ${
        scrolled 
          ? 'luxury-glass border-outline shadow-md' 
          : 'bg-background/40 border-transparent'
      }`}
    >
      <div className="flex justify-between w-full h-14 items-center px-6">
        {/* Left Section: Logo & Search */}
        <div className="flex items-center gap-6">
          <a href="/" className="flex items-center gap-3 text-on-background no-underline group">
            <div 
              className="w-7 h-7 rounded-md bg-primary text-on-primary grid place-items-center font-display font-bold text-sm tracking-tight transition-all duration-300"
              style={{ boxShadow: '0 0 12px var(--accent-color-line)' }}
            >
              AI
            </div>
            <div className="flex flex-col leading-[1.1]">
              <span className="text-sm font-sans font-semibold tracking-wide text-on-background">
                AITriage
              </span>
              <span className="text-label-caps text-on-surface-variant">
                {t('components.security_platform')}
              </span>
            </div>
          </a>

          {mode === 'advanced' && (
            <div className="hidden md:flex items-center bg-surface border border-outline rounded-md h-8 px-3 w-64 ml-4 transition-all duration-300 ease-out focus-within:border-primary focus-within:shadow-[0_0_0_1px_rgba(139,92,246,0.2)]">
              <span className="material-symbols-outlined text-on-surface-variant mr-2 text-[16px]">search</span>
              <input
                className="bg-transparent border-none focus:ring-0 focus:outline-none text-xs text-on-surface placeholder:text-on-surface-variant w-full h-full p-0 font-mono"
                placeholder={t('components.search')}
                type="text"
              />
            </div>
          )}
        </div>

        {/* Right Section: Status, Notifications, Copilot, Mode Toggle */}
        <div className="flex items-center gap-4">
          <button
            onClick={() => {
              document.body.classList.add('lang-changing');
              setTimeout(() => {
                i18n.changeLanguage(i18n.language.startsWith('ru') ? 'en' : 'ru');
                setTimeout(() => {
                  document.body.classList.remove('lang-changing');
                }, 80);
              }, 220);
            }}
            className="flex items-center justify-center w-8 h-8 rounded-md bg-surface text-on-surface border border-outline hover:border-outline-variant hover:bg-surface-bright transition-all duration-300 hover:-translate-y-[1px] relative overflow-hidden"
            title={t('components.switch_language')}
          >
            <AnimatePresence mode="wait" initial={false}>
              <motion.span
                key={i18n.language}
                initial={{ y: 8, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                exit={{ y: -8, opacity: 0 }}
                transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
                className="text-[12px] font-bold uppercase font-sans tracking-wider absolute"
              >
                {i18n.language.startsWith('ru') ? 'RU' : 'EN'}
              </motion.span>
            </AnimatePresence>
          </button>

          <button
            onClick={() => setShowSettings(true)}
            className="flex items-center justify-center w-8 h-8 rounded-md bg-surface text-on-surface border border-outline hover:border-outline-variant hover:bg-surface-bright transition-all duration-300 hover:-translate-y-[1px]"
            title={t('components.theme_settings')}
          >
            <span className="material-symbols-outlined text-[18px]">settings</span>
          </button>

          <button
            onClick={() => {
              setRebuildStatus('idle');
              setShowRebuildModal(true);
            }}
            className={`flex items-center justify-center w-8 h-8 rounded-md bg-surface text-on-surface border border-[var(--accent-color-line)] hover:border-[var(--accent-color)] hover:bg-[var(--accent-color-soft)] text-[var(--accent-color)] transition-all duration-300 hover:-translate-y-[1px] ${
              rebuildStatus === 'loading' ? 'pointer-events-none opacity-50' : ''
            }`}
            title={t('components.header.rebuild_title', 'Rebuild Container')}
          >
            <span className={`material-symbols-outlined text-[18px] ${rebuildStatus === 'loading' ? 'animate-spin' : ''}`}>
              restart_alt
            </span>
          </button>

          {mode === 'advanced' && (
            <button
              onClick={() => useCopilotStore.getState().toggle()}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-all duration-300 ease-out hover:-translate-y-[1px] ${
                isCopilotOpen
                  ? 'bg-primary text-on-primary'
                  : 'bg-surface text-on-surface border border-outline hover:border-outline-variant hover:bg-surface-bright'
              }`}
              style={isCopilotOpen ? { boxShadow: '0 0 12px var(--accent-color-line)' } : undefined}
            >
              <span className="material-symbols-outlined text-[16px]">smart_toy</span>
              <span>{t('components.copilot')}</span>
            </button>
          )}

          <button
            onClick={toggleMode}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-all duration-300 ease-out hover:-translate-y-[1px] bg-surface text-on-surface border border-outline hover:border-outline-variant hover:bg-surface-bright`}
            title={mode === 'simple' ? t('components.switch_advanced') : t('components.switch_simple')}
          >
            <span className="material-symbols-outlined text-[16px]">
              {mode === 'simple' ? 'dashboard_customize' : 'web_asset'}
            </span>
            <span>{mode === 'simple' ? t('components.advanced') : t('components.simple')}</span>
          </button>
        </div>
      </div>

      <AnimatePresence>
        {showSettings && (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm"
          >
            <motion.div 
              initial={{ opacity: 0, scale: 0.98, y: 15 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.98, y: 15 }}
              transition={{ type: "spring", stiffness: 350, damping: 30 }}
              className="w-[850px] h-[650px] bg-[#09090b]/90 backdrop-blur-3xl border border-[rgba(255,255,255,0.08)] rounded-2xl flex flex-col overflow-hidden shadow-[0_0_100px_-20px_var(--accent-color-soft)] relative"
            >
              {/* Subtle top glow */}
              <div className="absolute top-0 left-1/4 right-1/4 h-[1px] bg-gradient-to-r from-transparent via-[var(--accent-color-line)] to-transparent" />
              
              <div className="px-8 py-6 flex justify-between items-center relative z-10">
                <h2 className="text-[12px] uppercase tracking-[0.2em] text-[#f4f4f5] flex items-center gap-3 font-semibold">
                  <span className="material-symbols-outlined text-[18px] text-[var(--accent-color)]">settings</span>
                  {t('components.system_settings')}
                </h2>
                <button
                  onClick={() => setShowSettings(false)}
                  className="w-8 h-8 rounded-full flex items-center justify-center text-[#71717a] hover:text-[#f4f4f5] hover:bg-[rgba(255,255,255,0.05)] transition-all duration-300"
                >
                  <span className="material-symbols-outlined text-[20px]">close</span>
                </button>
              </div>
              
              <div className="flex flex-1 overflow-hidden border-t border-[rgba(255,255,255,0.04)] relative z-10">
                <div className="w-56 border-r border-[rgba(255,255,255,0.04)] bg-[rgba(0,0,0,0.2)] p-6 flex flex-col gap-3">
                  <button
                    onClick={() => setSettingsTab('theme')}
                    className={`text-left px-4 py-2.5 rounded-lg text-[11px] font-bold tracking-widest uppercase transition-all duration-300 ${
                      settingsTab === 'theme' ? 'bg-[var(--accent-color-soft)] text-[var(--accent-color)] border border-[var(--accent-color-line)] shadow-[inset_2px_0_0_0_var(--accent-color)]' : 'text-[#71717a] hover:text-[#f4f4f5] hover:bg-[rgba(255,255,255,0.03)] border border-transparent'
                    }`}
                  >
                    {t('components.theme')}
                  </button>
                  <button
                    onClick={() => setSettingsTab('database')}
                    className={`text-left px-4 py-2.5 rounded-lg text-[11px] font-bold tracking-widest uppercase transition-all duration-300 ${
                      settingsTab === 'database' ? 'bg-[var(--accent-color-soft)] text-[var(--accent-color)] border border-[var(--accent-color-line)] shadow-[inset_2px_0_0_0_var(--accent-color)]' : 'text-[#71717a] hover:text-[#f4f4f5] hover:bg-[rgba(255,255,255,0.03)] border border-transparent'
                    }`}
                  >
                    {t('components.database')}
                  </button>
                  <button
                    onClick={() => setSettingsTab('general')}
                    className={`text-left px-4 py-2.5 rounded-lg text-[11px] font-bold tracking-widest uppercase transition-all duration-300 ${
                      settingsTab === 'general' ? 'bg-[var(--accent-color-soft)] text-[var(--accent-color)] border border-[var(--accent-color-line)] shadow-[inset_2px_0_0_0_var(--accent-color)]' : 'text-[#71717a] hover:text-[#f4f4f5] hover:bg-[rgba(255,255,255,0.03)] border border-transparent'
                    }`}
                  >
                    {t('components.general')}
                  </button>
                </div>

                <div className="flex-1 p-8 overflow-y-auto cyber-scrollbar">
                  <AnimatePresence mode="wait">
                    {settingsTab === 'theme' && (
                      <motion.div
                        key="theme"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -10 }}
                        transition={{ duration: 0.2 }}
                        className="space-y-6"
                      >
                      <label className="text-[10px] text-[#71717a] tracking-[0.2em] font-semibold uppercase">{t('components.select_accent_palette')}</label>
                      <div className="grid grid-cols-2 gap-4">
                        {PALETTES.map((p) => {
                          const isActive = accent === p.id;
                          return (
                            <button
                              key={p.id}
                              onClick={() => {
                                setAccent(p.id);
                                try {
                                  localStorage.setItem('aitriage_accent', p.id);
                                } catch {}
                                document.documentElement.setAttribute('data-accent', p.id);
                                window.dispatchEvent(new Event('aitriage_accent_change'));
                              }}
                              className={`group flex items-center gap-4 p-4 rounded-xl transition-all duration-300 text-left relative overflow-hidden ${
                                isActive
                                  ? 'bg-[rgba(255,255,255,0.03)] border border-[var(--accent-color-line)] shadow-[0_0_20px_var(--accent-color-soft)]'
                                  : 'bg-[rgba(255,255,255,0.01)] border border-[rgba(255,255,255,0.04)] hover:bg-[rgba(255,255,255,0.03)] hover:border-[rgba(255,255,255,0.1)]'
                              }`}
                            >
                              {isActive && <div className="absolute inset-0 bg-gradient-to-r from-[var(--accent-color-soft)] to-transparent opacity-50" />}
                              <span
                                className="w-4 h-4 rounded-full shrink-0 relative z-10 shadow-lg"
                                style={{ backgroundColor: p.hex, boxShadow: isActive ? `0 0 12px ${p.hex}` : 'none' }}
                              />
                              <span className={`text-[11px] font-mono tracking-widest truncate relative z-10 transition-colors ${isActive ? 'text-[#f4f4f5] font-bold' : 'text-[#a1a1aa] group-hover:text-[#f4f4f5]'}`}>
                                {p.name.toUpperCase()}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                      </motion.div>
                    )}

                    {settingsTab === 'database' && (
                      <motion.div
                        key="database"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -10 }}
                        transition={{ duration: 0.2 }}
                        className="space-y-8"
                      >
                      <div>
                        <h3 className="text-[12px] font-bold text-[#ef4444] uppercase tracking-widest mb-2 flex items-center gap-2">
                          <span className="material-symbols-outlined text-[16px]">warning</span>
                          {t('components.danger_zone')}
                        </h3>
                        <p className="text-[11px] text-[#71717a] mb-6 font-mono">
                          {t('components.danger_zone_desc')}
                        </p>
                        
                        <div className="space-y-4">
                          <div className="p-5 border border-[#ef4444]/20 rounded-xl bg-gradient-to-r from-[#ef4444]/[0.02] to-transparent flex justify-between items-center group hover:border-[#ef4444]/40 transition-colors">
                            <div>
                              <div className="text-[13px] font-bold text-[#f4f4f5]">{t('components.clear_findings_cache')}</div>
                              <div className="text-[11px] text-[#a1a1aa] mt-1">{t('components.clear_cache_desc')}</div>
                            </div>
                            <button 
                              onClick={async () => {
                                if (confirm(t('components.confirm_clear_cache'))) {
                                  try {
                                    await api.post('/admin/clear-cache');
                                    window.location.reload();
                                  } catch (e) {
                                    alert(t('components.clear_cache_failed'));
                                  }
                                }
                              }}
                              className="px-5 py-2.5 bg-[rgba(239,68,68,0.1)] text-[#ef4444] border border-[#ef4444]/30 rounded-lg text-[11px] font-bold uppercase tracking-wider hover:bg-[#ef4444]/20 transition-all shadow-[0_0_15px_rgba(239,68,68,0)] hover:shadow-[0_0_15px_rgba(239,68,68,0.2)]"
                            >
                              {t('components.clear_cache_btn')}
                            </button>
                          </div>

                          <div className="p-5 border border-[#ef4444]/30 rounded-xl bg-gradient-to-r from-[#ef4444]/10 to-[#ef4444]/[0.02] flex justify-between items-center group shadow-[0_0_20px_rgba(239,68,68,0.05)]">
                            <div>
                              <div className="text-[13px] font-bold text-[#f4f4f5]">{t('components.purge_all_data')}</div>
                              <div className="text-[11px] text-[#a1a1aa] mt-1">{t('components.purge_all_data_desc')}</div>
                            </div>
                            <button 
                              onClick={async () => {
                                if (confirm(t('components.confirm_purge_data'))) {
                                  try {
                                    await api.post('/admin/purge');
                                    window.location.reload();
                                  } catch (e) {
                                    alert(t('components.purge_data_failed'));
                                  }
                                }
                              }}
                              className="px-5 py-2.5 bg-[#ef4444] text-white rounded-lg text-[11px] font-bold uppercase tracking-wider hover:bg-[#dc2626] transition-all shadow-[0_0_15px_rgba(239,68,68,0.3)] hover:shadow-[0_0_25px_rgba(239,68,68,0.5)]"
                            >
                              {t('components.purge_database_btn')}
                            </button>
                          </div>
                        </div>
                      </div>
                      </motion.div>
                    )}

                    {settingsTab === 'general' && (
                      <motion.div
                        key="general"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -10 }}
                        transition={{ duration: 0.2 }}
                        className="space-y-8"
                      >
                      <div>
                        <h3 className="text-[10px] text-[#71717a] tracking-[0.2em] font-semibold uppercase mb-4">{t('components.application_settings')}</h3>
                        <div className="space-y-4">
                          <div className="flex items-center justify-between p-5 border border-[rgba(255,255,255,0.06)] rounded-xl bg-[rgba(255,255,255,0.01)] hover:bg-[rgba(255,255,255,0.02)] transition-colors">
                            <div>
                              <div className="text-[13px] font-bold text-[#f4f4f5]">{t('components.enable_analytics')}</div>
                              <div className="text-[11px] text-[#a1a1aa] mt-1">{t('components.enable_analytics_desc')}</div>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input 
                                type="checkbox" 
                                className="sr-only peer" 
                                checked={enableAnalytics}
                                onChange={(e) => {
                                  setEnableAnalytics(e.target.checked);
                                  localStorage.setItem('aitriage_analytics', String(e.target.checked));
                                }}
                              />
                              <div className="w-12 h-6 bg-[rgba(255,255,255,0.1)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-6 peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-[#f4f4f5] after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[var(--accent-color)]"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between p-5 border border-[rgba(255,255,255,0.06)] rounded-xl bg-[rgba(255,255,255,0.01)] hover:bg-[rgba(255,255,255,0.02)] transition-colors">
                            <div>
                              <div className="text-[13px] font-bold text-[#f4f4f5]">{t('components.auto_scan_startup')}</div>
                              <div className="text-[11px] text-[#a1a1aa] mt-1">{t('components.auto_scan_startup_desc')}</div>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input 
                                type="checkbox" 
                                className="sr-only peer" 
                                checked={autoScan}
                                onChange={(e) => {
                                  setAutoScan(e.target.checked);
                                  localStorage.setItem('aitriage_autoscan', String(e.target.checked));
                                }}
                              />
                              <div className="w-12 h-6 bg-[rgba(255,255,255,0.1)] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-6 peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-[#f4f4f5] after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[var(--accent-color)]"></div>
                            </label>
                          </div>
                        </div>
                      </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              </div>
              
              <div className="p-6 flex justify-end relative z-10 border-t border-[rgba(255,255,255,0.04)]">
                <button
                  onClick={() => setShowSettings(false)}
                  className="px-8 py-3 bg-[var(--accent-color)] text-[var(--accent-color-on-text)] hover:bg-[var(--accent-color-hover)] rounded-lg text-[12px] font-bold uppercase tracking-widest shadow-[0_0_15px_var(--accent-color-line)] hover:shadow-[0_0_25px_var(--accent-color-soft)] transition-all duration-300"
                >
                  {t('components.apply_close')}
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <Modal
        isOpen={showRebuildModal}
        onClose={() => rebuildStatus !== 'loading' && setShowRebuildModal(false)}
        title={t('components.header.rebuild_title', 'Rebuild Container')}
        maxWidth="max-w-md"
      >
        <div className="flex flex-col gap-6 text-center py-4">
          {rebuildStatus === 'idle' && (
            <>
              <div className="mx-auto w-12 h-12 rounded-full bg-[var(--accent-color-soft)] border border-[var(--accent-color-line)] flex items-center justify-center text-[var(--accent-color)] mb-2">
                <span className="material-symbols-outlined text-[24px]">restart_alt</span>
              </div>
              <p className="text-sm text-v2-fg-2 leading-relaxed">
                {t('components.header.rebuild_confirm_text', 'Are you sure you want to rebuild the container? This will stop and compile the environment.')}
              </p>
              <div className="flex justify-center gap-3 mt-4">
                <button
                  onClick={() => setShowRebuildModal(false)}
                  className="v2-btn v2-btn-ghost px-5 py-2 font-mono text-[11px] uppercase tracking-wider"
                >
                  {t('components.header.rebuild_cancel', 'Cancel')}
                </button>
                <button
                  onClick={async () => {
                    setRebuildStatus('loading');
                    try {
                      await api.post('/admin/rebuild');
                      setRebuildStatus('success');
                    } catch (e) {
                      console.error(e);
                      setRebuildStatus('error');
                    }
                  }}
                  className="v2-btn v2-btn-red px-5 py-2 font-mono text-[11px] uppercase tracking-wider"
                >
                  {t('components.header.rebuild_confirm', 'Rebuild')}
                </button>
              </div>
            </>
          )}

          {rebuildStatus === 'loading' && (
            <>
              <div className="mx-auto w-12 h-12 rounded-full bg-[var(--accent-color-soft)] border border-[var(--accent-color-line)] flex items-center justify-center text-[var(--accent-color)] mb-2 animate-spin">
                <span className="material-symbols-outlined text-[24px]">sync</span>
              </div>
              <p className="text-sm text-white font-mono uppercase tracking-widest animate-pulse">
                {t('components.header.rebuilding', 'Rebuilding Environment...')}
              </p>
            </>
          )}

          {rebuildStatus === 'success' && (
            <>
              <div className="mx-auto w-12 h-12 rounded-full bg-success/10 border border-success/30 flex items-center justify-center text-success mb-2">
                <span className="material-symbols-outlined text-[24px]">check_circle</span>
              </div>
              <p className="text-sm text-v2-fg-2">
                {t('components.header.rebuild_success', 'Rebuild initiated successfully.')}
              </p>
              <div className="flex justify-center mt-4">
                <button
                  onClick={() => setShowRebuildModal(false)}
                  className="v2-btn v2-btn-ghost px-5 py-2 font-mono text-[11px] uppercase tracking-wider"
                >
                  {t('components.header.rebuild_close', 'Close')}
                </button>
              </div>
            </>
          )}

          {rebuildStatus === 'error' && (
            <>
              <div className="mx-auto w-12 h-12 rounded-full bg-error/10 border border-error/30 flex items-center justify-center text-error mb-2">
                <span className="material-symbols-outlined text-[24px]">error</span>
              </div>
              <p className="text-sm text-v2-fg-2">
                {t('components.header.rebuild_error', 'Failed to initiate rebuild or connection lost.')}
              </p>
              <div className="flex justify-center gap-3 mt-4">
                <button
                  onClick={() => setRebuildStatus('idle')}
                  className="v2-btn v2-btn-ghost px-5 py-2 font-mono text-[11px] uppercase tracking-wider"
                >
                  {t('components.header.rebuild_retry', 'Retry')}
                </button>
                <button
                  onClick={() => setShowRebuildModal(false)}
                  className="v2-btn v2-btn-ghost px-5 py-2 font-mono text-[11px] uppercase tracking-wider"
                >
                  {t('components.header.rebuild_close', 'Close')}
                </button>
              </div>
            </>
          )}
        </div>
      </Modal>
    </header>
  );
};
