import React, { useState, useRef, useEffect } from 'react';
import Markdown from 'react-markdown';
import { useCopilotStore } from '../store/CopilotStore';

import { useTranslation } from 'react-i18next';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface AICopilotProps {
  onClose?: () => void;
  isPinned?: boolean;
  onTogglePin?: () => void;
}

const QUICK_ACTIONS = [
  {
    labelKey: 'components.aiCopilot.quickActions.scanProject',
    icon: 'radar',
    promptKey: 'components.aiCopilot.quickActions.scanProjectPrompt',
  },
  {
    labelKey: 'components.aiCopilot.quickActions.threatModel',
    icon: 'security',
    promptKey: 'components.aiCopilot.quickActions.threatModelPrompt',
  },
  {
    labelKey: 'components.aiCopilot.quickActions.checkDeps',
    icon: 'inventory_2',
    promptKey: 'components.aiCopilot.quickActions.checkDepsPrompt',
  },
  {
    labelKey: 'components.aiCopilot.quickActions.bestPractices',
    icon: 'lightbulb',
    promptKey: 'components.aiCopilot.quickActions.bestPracticesPrompt',
  },
];

export const AICopilot: React.FC<AICopilotProps> = ({ onClose, isPinned, onTogglePin }) => {
  const { t } = useTranslation();
  const context = useCopilotStore((state) => state.context);
  const [messages, setMessages] = useState<Message[]>([]);

  useEffect(() => {
    setMessages([
      {
        role: 'assistant',
        content: t('components.aiCopilot.welcomeMessage'),
        timestamp: new Date(),
      },
    ]);
  }, [t]);

  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const sendMessage = async (userMsg: string) => {
    if (!userMsg.trim() || loading) return;

    setInput('');
    const newMessages: Message[] = [
      ...messages,
      { role: 'user', content: userMsg.trim(), timestamp: new Date() },
    ];
    setMessages(newMessages);
    setLoading(true);

    try {
      const apiMessages = [
        {
          role: 'system',
          content: 'You are a concise Security Co-pilot. Keep all answers short, structured, and free of fluff. Avoid long paragraphs; get straight to the point in Russian language.'
        },
        ...newMessages.map((m) => ({
          role: m.role === 'assistant' ? 'assistant' : 'user',
          content: context
            ? `Context: ${JSON.stringify(context)}\n\nUser: ${m.content}`
            : m.content,
        }))
      ];

      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          messages: apiMessages,
        }),
      });

      const data = await res.json();
      if (data.ok) {
        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content: data.content || 'Analysis complete.',
            timestamp: new Date(),
          },
        ]);
      } else {
        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content: `System Alert: ${data.error || 'Failed to connect to AI engine.'}`,
            timestamp: new Date(),
          },
        ]);
      }
    } catch {
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: 'System Alert: Connection interrupted. Please try again.',
          timestamp: new Date(),
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const promptToSubmit = useCopilotStore((state) => state.promptToSubmit);
  const setPromptToSubmit = useCopilotStore((state) => state.setPromptToSubmit);

  useEffect(() => {
    if (promptToSubmit) {
      sendMessage(promptToSubmit);
      setPromptToSubmit(null);
    }
  }, [promptToSubmit]);

  const handleSend = () => sendMessage(input);

  const handleClear = () => {
    setMessages([
      { role: 'assistant', content: t('components.aiCopilot.sessionCleared'), timestamp: new Date() },
    ]);
  };

  const formatTime = (d: Date) => d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

  return (
    <div className="flex flex-col h-full border-l border-outline-variant/40 bg-surface">
      {/* Header */}
      <div className="px-4 py-3 border-b border-outline-variant/30 flex justify-between items-center bg-surface-variant">
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 flex items-center justify-center bg-surface-container-high">
            <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '16px' }}>
              smart_toy
            </span>
          </div>
          <div>
            <span className="text-sm font-semibold text-primary">{t('components.aiCopilot.title')}</span>
            <div className="flex items-center gap-1.5">
              <div className="w-1.5 h-1.5 bg-success" />
              <span className="text-[10px] text-success">{t('components.aiCopilot.statusOnline')}</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={handleClear}
            className="w-8 h-8 flex items-center justify-center text-on-surface-variant/50 hover:text-primary hover:bg-surface-container transition-none"
            title={t('components.aiCopilot.clearChat')}
          >
            <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
              delete_sweep
            </span>
          </button>
          {onTogglePin && (
            <button
              onClick={onTogglePin}
              className={`w-8 h-8 flex items-center justify-center transition-none ${isPinned ? 'text-ai-accent bg-ai-accent/10' : 'text-on-surface-variant/50 hover:text-primary hover:bg-surface-container'}`}
              title={isPinned ? 'Unpin' : 'Pin'}
            >
              <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                {isPinned ? 'keep' : 'keep_off'}
              </span>
            </button>
          )}
          {onClose && (
            <button
              onClick={onClose}
              className="w-8 h-8 flex items-center justify-center text-on-surface-variant/50 hover:text-primary hover:bg-surface-container transition-none"
            >
              <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                close
              </span>
            </button>
          )}
        </div>
      </div>

      {/* Context Banner */}
      {!!context && (
        <div className="px-4 py-2 border-b border-ai-accent/20 flex items-center justify-between bg-surface-variant/40">
          <div className="flex items-center gap-2 overflow-hidden">
            <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '14px' }}>
              link
            </span>
            <span className="text-xs text-ai-accent truncate">
              {(context as any).title || `Finding #${(context as any).id}`}
            </span>
          </div>
          <button
            onClick={() => useCopilotStore.getState().setContext(null)}
            className="text-on-surface-variant/40 hover:text-error transition-none"
          >
            <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
              close
            </span>
          </button>
        </div>
      )}

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto cyber-scrollbar p-4 space-y-4">
        {messages.map((m, i) => (
          <div key={i} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div
              className={`max-w-[88%] px-4 py-3 ${
                m.role === 'user'
                  ? 'bg-primary text-on-primary'
                  : 'text-on-surface bg-surface-container'
              }`}
            >
              {m.role === 'assistant' ? (
                <div>
                  <div className="prose prose-invert prose-sm max-w-none [&_p]:text-[13px] [&_p]:leading-relaxed [&_p]:text-on-surface/90 [&_li]:text-[13px] [&_li]:text-on-surface/80 [&_code]:text-ai-accent [&_code]:bg-ai-accent/10 [&_code]:px-1 [&_code]: [&_strong]:text-primary [&_h2]:text-sm [&_h3]:text-xs">
                    <Markdown>{m.content}</Markdown>
                  </div>
                  {i > 0 && (
                    <div className="mt-2.5 pt-2.5 border-t border-white/5 flex gap-2">
                      <button
                        onClick={() => sendMessage(t('components.aiCopilot.autoFixPrompt', "Сгенерируй краткий промпт для другого ИИ (например, GitHub Copilot или разработчика), чтобы автоматически исправить эту уязвимость."))}
                        className="flex items-center gap-1.5 px-2.5 py-1.5 bg-primary/10 hover:bg-primary/20 border border-primary/20 text-primary text-[10px] font-mono font-medium rounded-lg transition-all cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[12px]">bolt</span>
                        {t('components.aiCopilot.autoFixButton', 'ПРОМПТ ДЛЯ АВТОФИКСА')}
                      </button>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-[13px] leading-relaxed">{m.content}</p>
              )}
              <div
                className={`text-[10px] mt-1.5 ${m.role === 'user' ? 'text-white/40 text-right' : 'text-on-surface-variant/30'}`}
              >
                {formatTime(m.timestamp)}
              </div>
            </div>
          </div>
        ))}
        {loading && (
          <div className="flex justify-start">
            <div className="px-4 py-3 bg-surface-container">
              <div className="flex gap-1.5 py-1">
                <div
                  className="w-2 h-5 bg-primary/40 animate-pulse"
                  style={{ animationDelay: '0ms' }}
                />
                <div
                  className="w-2 h-5 bg-primary/40 animate-pulse"
                  style={{ animationDelay: '150ms' }}
                />
                <div
                  className="w-2 h-5 bg-primary/40 animate-pulse"
                  style={{ animationDelay: '300ms' }}
                />
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Quick Actions */}
      {messages.length <= 1 && !loading && (
        <div className="px-4 pb-3 grid grid-cols-2 gap-2">
          {QUICK_ACTIONS.map((action) => (
            <button
              key={action.labelKey}
              onClick={() => sendMessage(t(action.promptKey))}
              className="flex items-center gap-2 px-3 py-2.5 border border-outline-variant/30 text-left transition-none hover:border-ai-accent/40 hover:bg-ai-accent/5 group"
            >
              <span
                className="material-symbols-outlined text-on-surface-variant/50 group-hover:text-ai-accent transition-none"
                style={{ fontSize: '16px' }}
              >
                {action.icon}
              </span>
              <span className="text-xs text-on-surface-variant group-hover:text-primary transition-none">
                {t(action.labelKey)}
              </span>
            </button>
          ))}
        </div>
      )}

      {/* Input */}
      <div className="p-3 border-t border-outline-variant/30 bg-surface-container-lowest">
        <div className="flex gap-2 items-end">
          <div className="flex-1 relative">
            <input
              className="w-full border border-outline-variant/30 px-4 py-2.5 text-[13px] text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 placeholder:text-on-surface-variant/30 transition-none bg-surface-container"
              placeholder={t('components.aiCopilot.placeholder')}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            />
          </div>
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className="w-10 h-10 flex items-center justify-center shrink-0 transition-none disabled:opacity-20"
            style={{
              background: input.trim() ? 'var(--color-primary)' : 'rgba(255,255,255,0.08)',
              color: input.trim() ? 'var(--color-on-primary)' : 'var(--color-primary)',
            }}
          >
            <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
              send
            </span>
          </button>
        </div>
      </div>
    </div>
  );
};
