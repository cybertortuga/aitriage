import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useTitle } from '../hooks/useTitle';

type Message = {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
};

export const AIChatPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('pages.aiChat.pageTitle'));
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'system',
      content: t('pages.aiChat.systemInitMessage'),
      timestamp: new Date(),
    },
  ]);
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const [loading, setLoading] = useState(false);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const newUserMsg: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date(),
    };

    const updatedMessages = [...messages, newUserMsg];
    setMessages(updatedMessages);
    setInput('');
    setLoading(true);

    fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages: updatedMessages
          .filter((m) => m.role !== 'system')
          .map((m) => ({ role: m.role, content: m.content })),
      }),
    })
      .then((res) => res.json())
      .then((data) => {
        if (data.ok) {
          setMessages((prev) => [
            ...prev,
            {
              id: (Date.now() + 1).toString(),
              role: 'assistant',
              content: data.content || t('pages.aiChat.analysisComplete'),
              timestamp: new Date(),
            },
          ]);
        } else {
          setMessages((prev) => [
            ...prev,
            {
              id: (Date.now() + 1).toString(),
              role: 'assistant',
              content: t('pages.aiChat.systemAlertError', { error: data.error || t('pages.aiChat.failedToConnect') }),
              timestamp: new Date(),
            },
          ]);
        }
      })
      .catch(() => {
        setMessages((prev) => [
          ...prev,
          {
            id: (Date.now() + 1).toString(),
            role: 'assistant',
            content: t('pages.aiChat.systemAlertConnectionInterrupted'),
            timestamp: new Date(),
          },
        ]);
      })
      .finally(() => setLoading(false));
  };

  return (
    <div className="flex flex-col min-h-full bg-v2-bg">
      {/* Page Header */}
      <div className="px-8 py-6 border-b border-v2-border-soft flex justify-between items-center shrink-0 bg-v2-surface">
        <div>
          <p className="text-[10px] font-bold tracking-widest text-v2-muted mb-1 uppercase">
            {t('pages.aiChat.headerSubtitle')}
          </p>
          <h1 className="text-2xl font-bold tracking-tight text-white uppercase">
            {t('pages.aiChat.headerTitle')}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <div className="w-2 h-2 bg-v2-red animate-pulse rounded-full" />
          <span className="text-[11px] font-bold tracking-widest text-v2-red uppercase">
            {t('pages.aiChat.modelName')}
          </span>
        </div>
      </div>

      <div className="flex-1 flex flex-col overflow-hidden bg-v2-bg">
        {/* Chat Messages */}
        <div className="flex-1 overflow-y-auto cyber-scrollbar p-8 flex flex-col gap-6">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex flex-col max-w-4xl w-full ${msg.role === 'user' ? 'ml-auto items-end' : 'mr-auto items-start'}`}
            >
              <div className="flex items-center gap-3 mb-2">
                <span
                  className={`text-[11px] font-bold tracking-widest uppercase ${
                    msg.role === 'user'
                      ? 'text-white'
                      : msg.role === 'system'
                        ? 'text-v2-muted'
                        : 'text-v2-red'
                  }`}
                >
                  {msg.role === 'user'
                    ? t('pages.aiChat.roleOperator')
                    : msg.role === 'system'
                      ? t('pages.aiChat.roleSystem')
                      : t('pages.aiChat.roleAgent')}
                </span>
                <span className="font-mono text-[10px] text-v2-muted">
                  {msg.timestamp.toLocaleTimeString([], { hour12: false })}
                </span>
              </div>
              <div
                className={`p-5 rounded-xl font-mono text-[13px] whitespace-pre-wrap leading-relaxed ${
                  msg.role === 'user'
                    ? 'bg-v2-surface border border-v2-border-soft text-white'
                    : msg.role === 'system'
                      ? 'border border-dashed border-v2-border-soft text-v2-muted bg-transparent'
                      : 'bg-v2-red-soft border border-v2-red-line text-white'
                }`}
              >
                {msg.content}
              </div>
            </div>
          ))}
          {loading && (
            <div className="mr-auto items-start flex flex-col w-full max-w-4xl">
              <span className="text-[11px] font-bold tracking-widest text-v2-red uppercase mb-2">
                {t('pages.aiChat.roleAgent')}
              </span>
              <div className="p-5 rounded-xl border border-v2-red-line bg-v2-red-soft">
                <div className="flex gap-2">
                  {[0, 1, 2].map((i) => (
                    <div
                      key={i}
                      className="w-2.5 h-2.5 bg-v2-red rounded-full animate-bounce"
                      style={{ animationDelay: `${i * 150}ms` }}
                    />
                  ))}
                </div>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="px-8 py-6 border-t border-v2-border-soft bg-v2-surface shrink-0">
          <form onSubmit={handleSend} className="flex gap-4 max-w-5xl mx-auto w-full">
            <div className="flex-1 relative">
              <div className="flex items-center v2-input px-5 py-4 gap-4 bg-v2-bg rounded-xl focus-within:border-v2-red transition-colors h-14">
                <span
                  className="material-symbols-outlined text-v2-muted"
                  style={{ fontSize: '20px' }}
                >
                  psychology
                </span>
                <input
                  type="text"
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder={t('pages.aiChat.inputPlaceholder')}
                  className="bg-transparent border-none outline-none text-white flex-1 font-mono text-[13px] placeholder:text-v2-muted h-full"
                  autoFocus
                />
              </div>
              <div className="absolute right-4 top-1/2 -translate-y-1/2 flex gap-3">
                <button type="button" className="text-v2-muted hover:text-white transition-colors">
                  <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                    attach_file
                  </span>
                </button>
                <button type="button" className="text-v2-muted hover:text-white transition-colors">
                  <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                    mic
                  </span>
                </button>
              </div>
            </div>
            <button
              type="submit"
              disabled={!input.trim() || loading}
              className="v2-btn v2-btn-red px-8 h-14 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span className="material-symbols-outlined" style={{ fontSize: '20px' }}>
                send
              </span>
              <span className="text-[11px] font-bold tracking-[0.2em] uppercase">{t('pages.aiChat.executeButton')}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
