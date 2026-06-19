import React, { useEffect, useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { useAuthStore } from '../store/AuthStore';
import api from '../services/api';
import { useCopilotStore } from '../store/CopilotStore';
import { useViewModeStore } from '../store/ViewModeStore';
import { AICopilot } from './AICopilot';
import { SimpleDashboardPage } from '../pages/SimpleDashboardPage';
import { SimpleAIChatPage } from '../pages/SimpleAIChatPage';
import { RepositoriesPage } from '../pages/RepositoriesPage';
import { TriagedPage } from '../pages/TriagedPage';
import { RunwayReportsPage } from '../pages/RunwayReportsPage';
import { FAQPage } from '../pages/FAQPage';
import type { Finding } from '../types';

type SimpleTab = 'overview' | 'repositories' | 'triaged' | 'reports' | 'chat' | 'faq';

const TABS: SimpleTab[] = ['overview', 'repositories', 'triaged', 'reports', 'chat', 'faq'];

const slideVariants = {
  enter: (direction: number) => ({
    x: direction > 0 ? 30 : direction < 0 ? -30 : 0,
    opacity: 0,
  }),
  center: {
    x: 0,
    opacity: 1,
  },
  exit: (direction: number) => ({
    x: direction > 0 ? -30 : direction < 0 ? 30 : 0,
    opacity: 0,
  }),
};

export const Layout: React.FC = () => {
  const { t } = useTranslation('components');
  const { user, setUser, logout } = useAuthStore();
  const viewMode = useViewModeStore((state) => state.mode);
  const navigate = useNavigate();
  const location = useLocation();
  const isCopilotOpen = useCopilotStore((state) => state.isOpen);
  const isCopilotPinned = useCopilotStore((state) => state.isPinned);

  const [tabState, setTabState] = useState<{ current: SimpleTab; direction: number }>({
    current: 'overview',
    direction: 0,
  });
  const simpleTab = tabState.current;
  const setSimpleTab = (nextTab: SimpleTab) => {
    const currentIndex = TABS.indexOf(tabState.current);
    const nextIndex = TABS.indexOf(nextTab);
    setTabState({
      current: nextTab,
      direction: nextIndex > currentIndex ? 1 : -1,
    });
  };

  const [chatContextFinding, setChatContextFinding] = useState<Finding | null>(null);
  const [chatInitialPrompt, setChatInitialPrompt] = useState<string | null>(null);

  const [isSidebarPinned, setIsSidebarPinned] = useState(() => {
    try {
      return localStorage.getItem('sidebar_pinned') === 'true';
    } catch {
      return false;
    }
  });

  const toggleSidebarPin = () => {
    setIsSidebarPinned((prev) => {
      const next = !prev;
      try {
        localStorage.setItem('sidebar_pinned', String(next));
      } catch {}
      return next;
    });
  };

  const handleNavigateToChat = (findingOrPrompt?: Finding | string) => {
    if (findingOrPrompt) {
      if (typeof findingOrPrompt === 'string') {
        setChatInitialPrompt(findingOrPrompt);
        setChatContextFinding(null);
      } else {
        setChatContextFinding(findingOrPrompt);
        setChatInitialPrompt(null);
      }
    } else {
      setChatContextFinding(null);
      setChatInitialPrompt(null);
    }
    setSimpleTab('chat');
  };

  // When entering simple mode, redirect to root so URL is clean
  useEffect(() => {
    if (viewMode === 'simple' && location.pathname !== '/') {
      navigate('/', { replace: true });
    }
  }, [viewMode]);

  useEffect(() => {
    api
      .get('/me')
      .then((res) => {
        if (res.data.ok) {
          setUser(res.data);
        } else {
          setUser({
            ok: true,
            id: 1,
            username: 'admin',
            global_role: 'superadmin',
            is_admin: true,
          });
        }
      })
      .catch(() => {
        setUser({ ok: true, id: 1, username: 'admin', global_role: 'superadmin', is_admin: true });
      });
  }, [setUser]);

  const handleLogout = () => {
    document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    logout();
    navigate('/login');
  };

  if (!user) {
    return (
      <div className="h-screen flex flex-col items-center justify-center bg-background gap-6">
        <div className="flex gap-1.5">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className="w-2 h-6 bg-primary animate-pulse"
              style={{ animationDelay: `${i * 0.15}s` }}
            />
          ))}
        </div>
        <span className="text-label-caps text-on-surface-variant/40 tracking-[0.25em] text-xs">
          {t('initializing_system')}
        </span>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden bg-background relative noise">
      {/* Background — mode-specific */}
      {viewMode === 'advanced' ? (
        <div className="absolute inset-0 overflow-hidden pointer-events-none z-0">
          <div className="ambient-bg-motion" />
          <div className="luxury-glow-orb-1" />
          <div className="luxury-glow-orb-2" />
          <div className="luxury-glow-orb-3" />
          <div className="absolute inset-0 grid-bg opacity-25" />
        </div>
      ) : (
        <div className="absolute inset-0 overflow-hidden pointer-events-none z-0">
          <div className="absolute inset-0 simple-bg-gradient" />
          <div className="ambient-bg-motion" />
        </div>
      )}

      <Header />

      {/* Simple Mode Tab Bar */}
      <div
        className={`relative z-20 border-b border-[rgba(255,255,255,0.06)] bg-background overflow-hidden transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] ${
          viewMode === 'simple'
            ? 'max-h-14 opacity-100 translate-y-0'
            : 'max-h-0 opacity-0 -translate-y-2 pointer-events-none'
        }`}
      >
        <div className="flex items-center gap-0 px-6">
          {([
            { id: 'overview' as SimpleTab, label: t('tab_overview') },
            { id: 'repositories' as SimpleTab, label: t('tab_repositories') },
            { id: 'triaged' as SimpleTab, label: t('tab_triaged') },
            { id: 'reports' as SimpleTab, label: t('tab_reports') },
            { id: 'chat' as SimpleTab, label: t('tab_ai_assistant') },
            { id: 'faq' as SimpleTab, label: t('tab_faq') },
          ]).map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSimpleTab(tab.id)}
              className={`relative px-4 py-2.5 text-[13px] font-medium transition-colors ${
                simpleTab === tab.id
                  ? 'text-[#f4f4f5]'
                  : 'text-[#52525b] hover:text-[#a1a1aa]'
              }`}
            >
              {tab.label}
              {simpleTab === tab.id && (
                <motion.div layoutId="simple-tab-indicator" className="absolute bottom-0 left-4 right-4 h-[1px] bg-[#f4f4f5]" />
              )}
            </button>
          ))}
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden relative z-10">
        <Sidebar
          onLogout={handleLogout}
          isPinned={isSidebarPinned}
          onTogglePin={toggleSidebarPin}
          isVisible={viewMode === 'advanced'}
        />

        {/* AI Copilot — advanced mode only */}
        {viewMode === 'advanced' && isCopilotOpen && (
          <aside
            className={`
 ${isCopilotPinned ? 'relative' : `absolute left-16 top-0 bottom-0 z-50`} 
 w-[400px] border-r border-outline-variant animate-in slide-in-from-left bg-surface shrink-0
 `}
            style={isSidebarPinned && !isCopilotPinned ? { left: '256px' } : undefined}
          >
            <AICopilot
              onClose={() => useCopilotStore.getState().setIsOpen(false)}
              isPinned={isCopilotPinned}
              onTogglePin={() =>
                useCopilotStore.getState().setIsPinned(!useCopilotStore.getState().isPinned)
              }
            />
          </aside>
        )}

        <div className="flex-1 min-w-0 flex overflow-hidden relative">
          <main className="flex-1 h-full bg-transparent relative overflow-hidden">
            <AnimatePresence mode="wait" custom={tabState.direction}>
              {viewMode === 'advanced' ? (
                <motion.div
                  key="advanced"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="h-full"
                >
                  <Outlet />
                </motion.div>
              ) : (
                <motion.div
                  key={`simple-${simpleTab}`}
                  custom={tabState.direction}
                  variants={slideVariants}
                  initial="enter"
                  animate="center"
                  exit="exit"
                  transition={{
                    x: { type: "spring", stiffness: 380, damping: 35 },
                    opacity: { duration: 0.15 }
                  }}
                  className="h-full w-full absolute inset-0"
                >
                  {simpleTab === 'overview' ? (
                    <SimpleDashboardPage onNavigateToChat={handleNavigateToChat} />
                  ) : simpleTab === 'repositories' ? (
                    <RepositoriesPage />
                  ) : simpleTab === 'triaged' ? (
                    <TriagedPage />
                  ) : simpleTab === 'reports' ? (
                    <RunwayReportsPage />
                  ) : simpleTab === 'chat' ? (
                    <SimpleAIChatPage
                      contextFinding={chatContextFinding}
                      initialPrompt={chatInitialPrompt}
                      onClearContext={() => setChatContextFinding(null)}
                      onClearInitialPrompt={() => setChatInitialPrompt(null)}
                    />
                  ) : (
                    <FAQPage />
                  )}
                </motion.div>
              )}
            </AnimatePresence>
          </main>
        </div>


      </div>
    </div>
  );
};
