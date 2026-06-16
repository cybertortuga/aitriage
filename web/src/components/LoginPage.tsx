import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/AuthStore';
import { useTranslation } from 'react-i18next';

// Animated ribbon background — three layers rotating + breathing
const V2RibbonBg: React.FC = () => {
  const layers = 18;
  const colors = ['var(--accent-color)', 'var(--accent-color-hover)', 'var(--accent-color-line)'];

  return (
    <svg
      viewBox="0 0 800 800"
      preserveAspectRatio="xMidYMid slice"
      className="fixed inset-0 w-full h-full pointer-events-none z-0"
      style={{ willChange: 'transform' }}
    >
      <defs>
        <radialGradient id="ribGlow" cx="55%" cy="50%" r="55%">
          <stop offset="0%" stopColor="var(--accent-color)" stopOpacity="0.35" />
          <stop offset="50%" stopColor="var(--accent-color-hover)" stopOpacity="0.15" />
          <stop offset="100%" stopColor="#000" stopOpacity="0" />
        </radialGradient>
      </defs>
      <rect width="800" height="800" fill="url(#ribGlow)" className="v2-ribbon-layer-3" />

      {/* Layer 1 — slow forward rotation */}
      <g className="v2-ribbon-layer-1" style={{ mixBlendMode: 'screen' }}>
        {Array.from({ length: layers }).map((_, i) => {
          const t = i / layers;
          const r = 90 + i * 38;
          const cx = 480 + Math.sin(i * 0.4) * 30;
          const cy = 440 + Math.cos(i * 0.35) * 26;
          const sw = 2.5 - t * 1.6;
          const opacity = (1 - t) * 0.7;
          const rx = r * (1 + Math.sin(i * 0.7) * 0.18);
          const ry = r * (1 + Math.cos(i * 0.5) * 0.12);
          const rot = i * 14;
          return (
            <ellipse
              key={`a${i}`}
              cx={cx}
              cy={cy}
              rx={rx}
              ry={ry}
              fill="none"
              stroke={colors[i % colors.length]}
              strokeWidth={sw}
              opacity={opacity}
              transform={`rotate(${rot} ${cx} ${cy})`}
            />
          );
        })}
      </g>

      {/* Layer 2 — slow reverse rotation, offset radii */}
      <g className="v2-ribbon-layer-2" style={{ mixBlendMode: 'screen' }}>
        {Array.from({ length: layers }).map((_, i) => {
          const t = i / layers;
          const r = 130 + i * 36;
          const cx = 480 + Math.sin(i * 0.5 + 1.3) * 25;
          const cy = 440 + Math.cos(i * 0.45 + 0.7) * 22;
          const sw = 2.0 - t * 1.2;
          const opacity = (1 - t) * 0.5;
          const rx = r * (1 + Math.cos(i * 0.6) * 0.2);
          const ry = r * (1 + Math.sin(i * 0.55) * 0.14);
          const rot = i * -11;
          return (
            <ellipse
              key={`b${i}`}
              cx={cx}
              cy={cy}
              rx={rx}
              ry={ry}
              fill="none"
              stroke={colors[(i + 1) % colors.length]}
              strokeWidth={sw}
              opacity={opacity}
              transform={`rotate(${rot} ${cx} ${cy})`}
            />
          );
        })}
      </g>
    </svg>
  );
};

export const LoginPage: React.FC = () => {
  const { t } = useTranslation('components');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuthStore();
  const navigate = useNavigate();

  const [accent, setAccent] = useState(() => {
    try {
      return localStorage.getItem('aitriage_accent') || 'white';
    } catch {
      return 'white';
    }
  });

  useEffect(() => {
    try {
      localStorage.setItem('aitriage_accent', accent);
    } catch {}
    document.documentElement.setAttribute('data-accent', accent);
  }, [accent]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      const data = await resp.json();
      if (data.ok) {
        login(
          {
            id: data.user_id,
            username: data.username,
            email: data.email || '',
            full_name: data.full_name || '',
            global_role: data.role,
          },
          data.token,
        );
        navigate('/');
      } else {
        setError(data.error || 'INVALID_CREDENTIALS');
      }
    } catch {
      setError('AUTH_SERVER_UNREACHABLE');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-[#0a0a0b] font-sans text-body-sm selection:bg-primary-container selection:text-white relative overflow-hidden">
      {/* ── Background animation layers ── */}
      <V2RibbonBg />
      <div className="fixed inset-0 bg-gradient-to-r from-black via-black/80 to-transparent pointer-events-none -z-10" />
      <div className="fixed inset-0 bg-gradient-to-t from-black via-transparent to-black pointer-events-none -z-10" />

      <div className="w-full max-w-md px-8 relative z-10">
        <div className="luxury-glass p-12 rounded-lg bg-[#111113]/75 border border-white/5 backdrop-blur-xl shadow-2xl">
          {/* Header */}
          <div className="mb-10 border-b border-outline-variant/30 pb-6">
            <div className="flex items-center gap-3 mb-2">
              <div 
                className="w-8 h-8 rounded-lg flex items-center justify-center font-extrabold text-white text-base select-none transition-all duration-300"
                style={{ backgroundColor: 'var(--accent-color)', boxShadow: '0 0 12px var(--accent-color-line)' }}
              >
                ai
              </div>
              <span className="text-xl font-bold tracking-tight text-white font-sans">
                {t('login.brand_name')}
              </span>
            </div>
            <div className="text-[10px] font-mono text-on-surface-variant opacity-60 uppercase tracking-[0.2em]">
              {t('login.secure_access_portal')}
            </div>
          </div>

          {/* Form */}
          <form onSubmit={handleLogin} className="flex flex-col gap-6">
            <div className="flex flex-col gap-2">
              <label className="text-[10px] font-mono text-on-surface-variant uppercase tracking-widest">
                {t('login.operator_identity')}
              </label>
              <input
                type="text"
                required
                placeholder={t('login.username_placeholder')}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="cyber-input w-full py-3 px-4 text-mono-data font-mono bg-[#0a0a0b] border border-white/5 rounded-lg focus:outline-none focus:border-primary/45 focus:ring-1 focus:ring-primary/20 transition-all duration-300 ease-out"
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-[10px] font-mono text-on-surface-variant uppercase tracking-widest">
                {t('login.access_token')}
              </label>
              <input
                type="password"
                required
                placeholder="••••••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="cyber-input w-full py-3 px-4 text-mono-data font-mono bg-[#0a0a0b] border border-white/5 rounded-lg focus:outline-none focus:border-primary/45 focus:ring-1 focus:ring-primary/20 transition-all duration-300 ease-out"
              />
            </div>

            {error && (
              <div className="p-3 border border-[#ef4444]/30 bg-[#ef4444]/5 rounded-lg flex items-center gap-3 transition-all duration-300 ease-out">
                <div className="w-1.5 h-1.5 bg-[#ef4444] rounded-full shrink-0 shadow-[0_0_8px_rgba(239,68,68,0.4)]" />
                <span className="text-[10px] font-mono text-[#ef4444] uppercase tracking-widest">
                  {t(`login.errors.${error.toLowerCase()}`, error)}
                </span>
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="btn-primary w-full py-3 mt-2 flex items-center justify-center gap-3 rounded-lg font-sans font-semibold tracking-wider active:scale-[0.98] transition-all duration-300 ease-out disabled:opacity-40 z-10 cursor-pointer"
              style={{ boxShadow: '0 0 12px var(--accent-color-soft)' }}
            >
              {isLoading ? (
                <div className="flex gap-1 items-center justify-center">
                  {[0, 1, 2].map((i) => (
                    <div
                      key={i}
                      className="w-1 h-3 bg-white animate-pulse"
                      style={{ animationDelay: `${i * 0.15}s` }}
                    />
                  ))}
                  <span className="ml-2">{t('login.verifying')}</span>
                </div>
              ) : (
                t('login.initialize_session')
              )}
            </button>
          </form>

          {/* Footer */}
          <div className="mt-10 pt-6 border-t border-outline-variant/30 flex justify-between items-center">
            <span className="text-[10px] font-mono text-on-surface-variant opacity-40">v1.4.2</span>

            {/* Accent Picker */}
            <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg border border-outline-variant/30 bg-[#0a0a0b]">
              <button
                type="button"
                onClick={() => setAccent('white')}
                className={`w-3 h-3 rounded-full transition-all hover:scale-110 border ${accent === 'white' ? 'border-white scale-105' : 'border-transparent'}`}
                style={{ backgroundColor: '#f8fafc' }}
                title={t('login.accent_white')}
              />
              <button
                type="button"
                onClick={() => setAccent('violet')}
                className={`w-3 h-3 rounded-full transition-all hover:scale-110 border ${accent === 'violet' ? 'border-white scale-105' : 'border-transparent'}`}
                style={{ backgroundColor: '#8b5cf6' }}
                title={t('login.accent_violet')}
              />
              <button
                type="button"
                onClick={() => setAccent('cyan')}
                className={`w-3 h-3 rounded-full bg-[#06b6d4] transition-all hover:scale-110 border ${accent === 'cyan' ? 'border-white scale-105' : 'border-transparent'}`}
                title={t('login.accent_cyan')}
              />
              <button
                type="button"
                onClick={() => setAccent('amber')}
                className={`w-3 h-3 rounded-full bg-[#f59e0b] transition-all hover:scale-110 border ${accent === 'amber' ? 'border-white scale-105' : 'border-transparent'}`}
                title={t('login.accent_amber')}
              />
            </div>

            <span className="text-[10px] font-mono text-on-surface-variant opacity-40">
              {t('login.standby')}
            </span>
          </div>
        </div>

        <p className="mt-6 text-center text-[9px] font-mono text-on-surface-variant opacity-30 tracking-[0.2em] uppercase">
          {t('login.protected_by')}
        </p>
      </div>
    </div>
  );
};
