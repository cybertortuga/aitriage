import React, { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import Markdown from 'react-markdown';
import { useTranslation } from 'react-i18next';
import type { Finding } from '../types';

/* ── Types ── */
type ChatSession = {
  id: number;
  title: string;
  created_at: string;
  updated_at: string;
};

type Message = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
};

type DBMessage = {
  id: number;
  session_id: number;
  role: string;
  content: string;
  created_at: string;
};

/* ── API helpers ── */
const api = {
  async listSessions(): Promise<ChatSession[]> {
    const res = await fetch('/api/chat/sessions');
    const data = await res.json();
    return data.ok ? (data.sessions || []) : [];
  },
  async createSession(title: string): Promise<number> {
    const res = await fetch('/api/chat/sessions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title }),
    });
    const data = await res.json();
    return data.id;
  },
  async deleteSession(id: number): Promise<void> {
    await fetch(`/api/chat/sessions/${id}`, { method: 'DELETE' });
  },
  async renameSession(id: number, title: string): Promise<void> {
    await fetch(`/api/chat/sessions/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title }),
    });
  },
  async getMessages(sessionId: number): Promise<DBMessage[]> {
    const res = await fetch(`/api/chat/messages?session_id=${sessionId}`);
    const data = await res.json();
    return data.ok ? (data.messages || []) : [];
  },
  async addMessage(sessionId: number, role: string, content: string): Promise<void> {
    await fetch('/api/chat/messages', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ session_id: sessionId, role, content }),
    });
  },
};

/* ── Quick actions ── */
const getQuickActions = (t: any) => [
  { label: t('simpleAIChat.quickActions.issues.label'), icon: 'search', prompt: t('simpleAIChat.quickActions.issues.prompt') },
  { label: t('simpleAIChat.quickActions.improve.label'), icon: 'lightbulb', prompt: t('simpleAIChat.quickActions.improve.prompt') },
  { label: t('simpleAIChat.quickActions.critical.label'), icon: 'priority_high', prompt: t('simpleAIChat.quickActions.critical.prompt') },
  { label: t('simpleAIChat.quickActions.report.label'), icon: 'description', prompt: t('simpleAIChat.quickActions.report.prompt') },
];

/* ── Component ── */
interface SimpleAIChatPageProps {
  contextFinding?: Finding | null;
  initialPrompt?: string | null;
  onClearContext?: () => void;
  onClearInitialPrompt?: () => void;
}

export const SimpleAIChatPage: React.FC<SimpleAIChatPageProps> = ({ 
  contextFinding, 
  initialPrompt,
  onClearContext,
  onClearInitialPrompt
}) => {
  const { t } = useTranslation('pages');

  // Session state
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<number | null>(null);
  const [sessionsLoading, setSessionsLoading] = useState(true);

  // Chat state
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  // UI state
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editTitle, setEditTitle] = useState('');

  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Load sessions on mount
  useEffect(() => {
    loadSessions();
  }, []);

  // Auto-scroll
  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [messages, loading]);

  // Handle context finding
  useEffect(() => {
    if (contextFinding) {
      const contextMsg = `${t('simpleAIChat.context.explainPrefix')}\n\n**${contextFinding.title}**\n${t('simpleAIChat.context.severity')} ${contextFinding.severity}\n${t('simpleAIChat.context.file')} ${contextFinding.file_path || t('simpleAIChat.context.notSpecified')}\n${contextFinding.description ? `\n${t('simpleAIChat.context.description')} ${contextFinding.description}` : ''}`;
      handleNewChatWithMessage(contextMsg);
      if (onClearContext) onClearContext();
    }
  }, [contextFinding, t, onClearContext]);

  // Handle initial prompt
  useEffect(() => {
    if (initialPrompt) {
      handleNewChatWithMessage(initialPrompt);
      if (onClearInitialPrompt) onClearInitialPrompt();
    }
  }, [initialPrompt, onClearInitialPrompt]);

  const loadSessions = async () => {
    setSessionsLoading(true);
    try {
      const s = await api.listSessions();
      setSessions(s);
    } catch { /* ignore */ }
    setSessionsLoading(false);
  };

  const loadMessages = async (sessionId: number) => {
    try {
      const dbMsgs = await api.getMessages(sessionId);
      setMessages(dbMsgs.filter(m => m.role !== 'system').map(m => ({
        id: m.id.toString(),
        role: m.role as 'user' | 'assistant',
        content: m.content,
        timestamp: new Date(m.created_at),
      })));
    } catch { /* ignore */ }
  };

  const selectSession = async (sessionId: number) => {
    setActiveSessionId(sessionId);
    await loadMessages(sessionId);
  };

  const handleNewChat = () => {
    setActiveSessionId(null);
    setMessages([]);
    if (onClearContext) onClearContext();
    inputRef.current?.focus();
  };

  const handleNewChatWithMessage = async (msg: string) => {
    // Create session, then send
    const title = msg.slice(0, 50).replace(/\n/g, ' ') + (msg.length > 50 ? '...' : '');
    const sessionId = await api.createSession(title);
    setActiveSessionId(sessionId);
    setMessages([]);
    await loadSessions();
    await sendMessage(msg, sessionId);
  };

  const handleDeleteSession = async (id: number) => {
    await api.deleteSession(id);
    if (activeSessionId === id) {
      setActiveSessionId(null);
      setMessages([]);
    }
    await loadSessions();
  };

  const handleRenameSession = async (id: number) => {
    if (editTitle.trim()) {
      await api.renameSession(id, editTitle.trim());
      setEditingId(null);
      await loadSessions();
    }
  };

  const sendMessage = async (userMsg: string, forceSessionId?: number) => {
    if (!userMsg.trim() || loading) return;
    setInput('');

    let sessionId = forceSessionId ?? activeSessionId;

    // Create session if none active
    if (!sessionId) {
      const title = userMsg.slice(0, 50).replace(/\n/g, ' ') + (userMsg.length > 50 ? '...' : '');
      sessionId = await api.createSession(title);
      setActiveSessionId(sessionId);
      await loadSessions();
    }

    // Save user message to DB
    await api.addMessage(sessionId, 'user', userMsg.trim());

    const newUserMsg: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: userMsg.trim(),
      timestamp: new Date(),
    };
    const updatedMessages = [...messages, newUserMsg];
    setMessages(updatedMessages);
    setLoading(true);

    try {
      const apiMessages = [
        ...updatedMessages.map(m => ({ role: m.role === 'assistant' ? 'assistant' : 'user', content: m.content })),
      ];
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: apiMessages }),
      });
      const data = await res.json();
      const reply = data.ok ? (data.content || t('simpleAIChat.chat.done')) : `${t('simpleAIChat.chat.errorPrefix')} ${data.error || t('simpleAIChat.chat.couldNotConnect')}`;

      // Save assistant reply to DB
      await api.addMessage(sessionId, 'assistant', reply);

      setMessages(prev => [...prev, {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: reply,
        timestamp: new Date(),
      }]);

      // Auto-rename session with first user message
      if (updatedMessages.filter(m => m.role === 'user').length === 1) {
        const shortTitle = userMsg.slice(0, 60).replace(/\n/g, ' ');
        await api.renameSession(sessionId, shortTitle);
        await loadSessions();
      }
    } catch {
      const errMsg = t('simpleAIChat.chat.connectionLost');
      await api.addMessage(sessionId, 'assistant', errMsg);
      setMessages(prev => [...prev, {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: errMsg,
        timestamp: new Date(),
      }]);
    } finally {
      setLoading(false);
    }
  };

  const handleSend = (e: React.FormEvent) => { e.preventDefault(); sendMessage(input); };
  const formatTime = (d: Date) => d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  const isEmpty = messages.length === 0 && !activeSessionId;

  return (
    <div className="flex h-full overflow-hidden">
      {/* ── Sidebar: Chat History ── */}
      <motion.div
        animate={{ width: sidebarOpen ? 260 : 0 }}
        transition={{ type: "spring", stiffness: 300, damping: 30 }}
        className="shrink-0 border-r border-[rgba(255,255,255,0.06)] flex flex-col bg-background overflow-hidden h-full"
      >
        <div className={`w-[260px] h-full flex flex-col shrink-0 transition-opacity duration-200 ${sidebarOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}>
          {/* New Chat Button */}
          <div className="p-3">
            <button
              onClick={handleNewChat}
              className="w-full flex items-center gap-2 px-3 py-2.5 rounded-lg border border-[rgba(255,255,255,0.08)] text-[13px] text-[#a1a1aa] hover:text-[#f4f4f5] hover:bg-[rgba(255,255,255,0.03)] transition-colors"
            >
              <span className="material-symbols-outlined text-[16px]">add</span>
              {t('simpleAIChat.ui.newChat')}
            </button>
          </div>

          {/* Sessions List */}
          <div className="flex-1 overflow-y-auto px-2 pb-2" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
            {sessionsLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="w-4 h-4 border-2 border-[#27272a] border-t-[#52525b] rounded-full animate-spin" />
              </div>
            ) : sessions.length === 0 ? (
              <div className="text-center py-8 px-4">
                <p className="text-[12px] text-[#3f3f46]">{t('simpleAIChat.ui.noConversations')}</p>
              </div>
            ) : (
              <motion.div 
                initial="hidden" 
                animate="visible" 
                variants={{ visible: { transition: { staggerChildren: 0.05 } } }}
              >
                <AnimatePresence initial={false}>
                  {sessions.map(s => (
                    <motion.div
                      key={s.id}
                      variants={{ hidden: { opacity: 0, x: -10 }, visible: { opacity: 1, x: 0 } }}
                      exit={{ opacity: 0, x: -20, height: 0, marginTop: 0, marginBottom: 0, overflow: 'hidden' }}
                      transition={{ duration: 0.2 }}
                      className={`group relative rounded-lg mb-0.5 overflow-hidden ${activeSessionId === s.id ? 'bg-[rgba(255,255,255,0.05)]' : 'hover:bg-[rgba(255,255,255,0.02)]'}`}
                    >
                      {editingId === s.id ? (
                        <form onSubmit={(e) => { e.preventDefault(); handleRenameSession(s.id); }} className="px-3 py-2">
                          <input
                            value={editTitle}
                            onChange={e => setEditTitle(e.target.value)}
                            onBlur={() => handleRenameSession(s.id)}
                            autoFocus
                            className="w-full bg-surface-bright border border-[rgba(255,255,255,0.1)] rounded px-2 py-1 text-[12px] text-[#f4f4f5] outline-none"
                          />
                        </form>
                      ) : (
                        <button
                          onClick={() => selectSession(s.id)}
                          className="w-full text-left px-3 py-2.5 flex items-center gap-2"
                        >
                          <span className="material-symbols-outlined text-[14px] text-[#3f3f46]">chat_bubble</span>
                          <span className="text-[13px] text-[#a1a1aa] truncate flex-1">{s.title}</span>
                        </button>
                      )}
                      {/* Actions */}
                      {editingId !== s.id && (
                        <div className="absolute right-1 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-0.5 bg-surface-bright rounded-md border border-[rgba(255,255,255,0.06)] px-0.5">
                          <button
                            onClick={(e) => { e.stopPropagation(); setEditingId(s.id); setEditTitle(s.title); }}
                            className="p-1 text-[#52525b] hover:text-[#a1a1aa] transition-colors"
                            title={t('simpleAIChat.ui.rename')}
                          >
                            <span className="material-symbols-outlined text-[14px]">edit</span>
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); handleDeleteSession(s.id); }}
                            className="p-1 text-[#52525b] hover:text-[#ef4444] transition-colors"
                            title={t('simpleAIChat.ui.delete')}
                          >
                            <span className="material-symbols-outlined text-[14px]">delete</span>
                          </button>
                        </div>
                      )}
                    </motion.div>
                  ))}
                </AnimatePresence>
              </motion.div>
            )}
          </div>
        </div>
      </motion.div>

      {/* ── Main Chat Area ── */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top bar */}
        <div className="shrink-0 h-10 border-b border-[rgba(255,255,255,0.06)] flex items-center px-3 gap-2">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-1 text-[#52525b] hover:text-[#a1a1aa] transition-colors"
            title={sidebarOpen ? t('simpleAIChat.ui.hideSidebar') : t('simpleAIChat.ui.showSidebar')}
          >
            <span className="material-symbols-outlined text-[18px]">{sidebarOpen ? 'left_panel_close' : 'left_panel_open'}</span>
          </button>
          {activeSessionId && (
            <span className="text-[12px] text-[#52525b] truncate">
              {sessions.find(s => s.id === activeSessionId)?.title || t('simpleAIChat.ui.chat')}
            </span>
          )}
        </div>

        {/* Context Banner */}
        {contextFinding && (
          <div className="px-6 py-2.5 border-b border-[rgba(255,255,255,0.06)] flex items-center justify-between shrink-0">
            <div className="flex items-center gap-2 overflow-hidden">
              <span className="material-symbols-outlined text-[#52525b] text-[14px]">link</span>
              <span className="text-[12px] text-[#71717a] truncate">
                {t('simpleAIChat.ui.discussing')}<strong className="text-[#a1a1aa]">{contextFinding.title}</strong>
              </span>
            </div>
            {onClearContext && (
              <button onClick={onClearContext} className="text-[#3f3f46] hover:text-[#ef4444] transition-colors p-1">
                <span className="material-symbols-outlined text-[14px]">close</span>
              </button>
            )}
          </div>
        )}

        {/* Messages */}
        <div ref={scrollRef} className="flex-1 overflow-y-auto" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
          <AnimatePresence mode="wait">
            {isEmpty ? (
              <motion.div 
                key="empty"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                transition={{ duration: 0.2 }}
                className="flex flex-col items-center justify-center h-full px-6 py-12"
              >
                <div className="w-10 h-10 rounded-lg bg-surface-bright flex items-center justify-center mb-4">
                  <span className="material-symbols-outlined text-[#52525b] text-[24px]">smart_toy</span>
                </div>
                <h2 className="text-[15px] font-medium text-[#f4f4f5] mb-1">{t('simpleAIChat.ui.aiAssistant')}</h2>
                <p className="text-[13px] text-[#3f3f46] text-center max-w-md mb-8">
                  {t('simpleAIChat.ui.askQuestion')}
                </p>
                <motion.div 
                  initial="hidden" animate="visible"
                  variants={{ visible: { transition: { staggerChildren: 0.1, delayChildren: 0.2 } } }}
                  className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-w-lg w-full"
                >
                  {getQuickActions(t).map((action: any) => (
                    <motion.button
                      variants={{ hidden: { opacity: 0, y: 10 }, visible: { opacity: 1, y: 0 } }}
                      key={action.label}
                      onClick={() => sendMessage(action.prompt)}
                      className="flex items-center gap-3 p-3 rounded-lg border border-[rgba(255,255,255,0.06)] text-left hover:bg-[rgba(255,255,255,0.02)] transition-colors"
                    >
                      <span className="material-symbols-outlined text-[#3f3f46] text-[18px]">{action.icon}</span>
                      <span className="text-[13px] text-[#71717a]">{action.label}</span>
                    </motion.button>
                  ))}
                </motion.div>
              </motion.div>
            ) : (
              <motion.div
                key="chat-messages"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="max-w-3xl mx-auto px-6 py-6 space-y-5"
              >
                <AnimatePresence initial={false}>
                  {messages.map(msg => (
                    <motion.div 
                      key={msg.id} 
                      layout="position"
                      initial={{ opacity: 0, y: 12 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ type: "spring", stiffness: 350, damping: 30 }}
                      className={`flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}
                    >
                      <div className={`w-7 h-7 rounded-md flex items-center justify-center shrink-0 mt-0.5 ${
                        msg.role === 'user' ? 'bg-surface-container-highest' : 'bg-surface-bright'
                      }`}>
                        <span className="material-symbols-outlined text-[#71717a] text-[14px]">
                          {msg.role === 'user' ? 'person' : 'smart_toy'}
                        </span>
                      </div>
                      <div className={`max-w-[85%] flex flex-col ${msg.role === 'user' ? 'items-end' : 'items-start'}`}>
                        <div className={`px-4 py-2.5 rounded-lg text-[13px] leading-relaxed ${
                          msg.role === 'user'
                            ? 'bg-surface-container-highest text-[#f4f4f5] rounded-tr-sm'
                            : 'text-[#a1a1aa] rounded-tl-sm'
                        }`}>
                          {msg.role === 'assistant' ? (
                            <div className="simple-chat-markdown prose prose-invert prose-sm max-w-none [&_p]:text-[13px] [&_p]:leading-relaxed [&_p]:text-[#a1a1aa] [&_p]:m-0 [&_li]:text-[13px] [&_strong]:text-[#f4f4f5] [&_code]:text-[var(--accent-color)] [&_code]:bg-surface-bright [&_code]:px-1 [&_code]:rounded">
                              <Markdown>{msg.content}</Markdown>
                            </div>
                          ) : <p>{msg.content}</p>}
                        </div>
                        <span className="text-[10px] text-[#27272a] mt-1 px-1">{formatTime(msg.timestamp)}</span>
                      </div>
                    </motion.div>
                  ))}
                  {loading && (
                    <motion.div 
                      key="loading"
                      layout="position"
                      initial={{ opacity: 0, y: 12 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -12 }}
                      transition={{ duration: 0.2 }}
                      className="flex gap-3"
                    >
                      <div className="w-7 h-7 rounded-md flex items-center justify-center bg-surface-bright">
                        <span className="material-symbols-outlined text-[#71717a] text-[14px]">smart_toy</span>
                      </div>
                      <div className="flex gap-1 items-center px-2 py-3">
                        {[0, 1, 2].map(i => (
                          <div key={i} className="w-1.5 h-1.5 rounded-full bg-[#3f3f46] animate-pulse" style={{ animationDelay: `${i * 150}ms` }} />
                        ))}
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Input */}
        <div className="shrink-0 border-t border-[rgba(255,255,255,0.06)] px-6 py-4">
          <form onSubmit={handleSend} className="max-w-3xl mx-auto flex gap-2">
            <input
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              placeholder={t('simpleAIChat.ui.inputPlaceholder')}
              className="flex-1 bg-surface-bright border border-[rgba(255,255,255,0.06)] rounded-lg px-4 py-2.5 text-[13px] text-[#f4f4f5] placeholder:text-[#3f3f46] outline-none focus:border-[rgba(255,255,255,0.12)] transition-colors"
              autoFocus
            />
            <button
              type="submit"
              disabled={!input.trim() || loading}
              className="w-9 h-9 rounded-lg bg-[#f4f4f5] text-[var(--bg-color)] flex items-center justify-center disabled:opacity-20 transition-opacity shrink-0"
            >
              <span className="material-symbols-outlined text-[18px]">{loading ? 'hourglass_top' : 'arrow_upward'}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
