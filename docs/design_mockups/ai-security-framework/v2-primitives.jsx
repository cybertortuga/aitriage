// v2 shared primitives (PT-style)
const { useState: uS, useEffect: uE, useRef: uR, useMemo: uM } = React;

function V2Logo() {
  return (
    <a href="landing-v2.html" className="v2-logo">
      <span className="v2-logo-mark">ai</span>
      <span style={{display:'flex', flexDirection:'column', lineHeight:1.1}}>
        <span style={{fontSize:16, fontWeight:700, letterSpacing:'-0.01em'}}>AISec Fabric</span>
        <span style={{fontSize:11, color:'var(--v2-muted)', fontWeight:500, letterSpacing:0}}>AI Security Framework</span>
      </span>
    </a>
  );
}

function V2Nav({ active }) {
  const [scrolled, setScrolled] = uS(false);
  uE(() => {
    const onS = () => setScrolled(window.scrollY > 16);
    window.addEventListener('scroll', onS);
    return () => window.removeEventListener('scroll', onS);
  }, []);
  const links = [
    {l:'Платформа', h:'#products', k:'products'},
    {l:'Сервисы', h:'#services', k:'services'},
    {l:'Как работает', h:'#how', k:'how'},
    {l:'Compliance', h:'#compliance', k:'compliance'},
    {l:'Исследования', h:'#research', k:'research'},
    {l:'Контакты', h:'#contact', k:'contact'},
  ];
  return (
    <nav className={`v2-nav ${scrolled ? 'scrolled' : ''}`}>
      <div className="v2-container v2-nav-inner">
        <V2Logo/>
        <div className="v2-nav-links">
          {links.map(l => (
            <a key={l.k} href={l.h} className={`v2-nav-link ${active === l.k ? 'active' : ''}`}>{l.l}</a>
          ))}
          <a href="#" className="v2-nav-link" style={{display:'inline-flex', alignItems:'center', gap:6}}>
            Инвесторам <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 17L17 7M17 7H8M17 7V16"/></svg>
          </a>
        </div>
        <div style={{display:'flex', gap:8, alignItems:'center'}}>
          <button className="v2-icon-btn" title="Язык" style={{width:'auto', padding:'0 14px', gap:6, display:'inline-flex'}}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6"><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3c2.5 3 2.5 15 0 18M12 3c-2.5 3-2.5 15 0 18"/></svg>
            <span style={{fontSize:14, fontWeight:500}}>Ру</span>
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
          </button>
          <button className="v2-icon-btn" title="Поддержка">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/></svg>
          </button>
          <button className="v2-icon-btn" title="Поиск">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
          </button>
        </div>
      </div>
    </nav>
  );
}

function V2Footer() {
  const cols = [
    {t:'Продукты', items:['AI Gateway','Shadow AI Discovery','AI DLP','AI Asset Graph','Policy & Compliance','Audit & Evidence']},
    {t:'Сервисы', items:['Пилот за 1.5 млн','AI Red Team','Внедрение','Поддержка 24/7','Обучение CISO','Threat Hunting']},
    {t:'Технологии', items:['Архитектура','Документация','API Reference','SDK Python / Go','Whitepapers','Changelog']},
    {t:'Компания', items:['О нас','Партнёры','Карьера','Новости','Контакты','Инвесторам']},
  ];
  return (
    <footer className="v2-footer">
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1.4fr repeat(4, 1fr)', gap:48}}>
          <div>
            <V2Logo/>
            <p style={{fontSize:14, color:'var(--v2-muted)', lineHeight:1.6, maxWidth:280, marginTop:24, marginBottom:24}}>
              AI Security Framework для enterprise. Видимость, контроль, политики
              и compliance evidence для любых AI-систем в вашем периметре.
            </p>
            <div style={{display:'flex', gap:8, flexWrap:'wrap'}}>
              <span className="v2-tag">152-ФЗ</span>
              <span className="v2-tag">187-ФЗ КИИ</span>
              <span className="v2-tag">ФСТЭК</span>
            </div>
          </div>
          {cols.map(c => (
            <div key={c.t}>
              <div style={{fontSize:13, fontWeight:600, marginBottom:16, color:'var(--v2-fg)'}}>{c.t}</div>
              <div style={{display:'flex', flexDirection:'column', gap:10}}>
                {c.items.map(i =>
                  <a key={i} href="#" style={{fontSize:14, color:'var(--v2-muted)', textDecoration:'none'}}
                    onMouseOver={e=>e.target.style.color='var(--v2-fg)'}
                    onMouseOut={e=>e.target.style.color='var(--v2-muted)'}>{i}</a>
                )}
              </div>
            </div>
          ))}
        </div>
        <div style={{marginTop:72, paddingTop:24, borderTop:'1px solid var(--v2-border-soft)',
          display:'flex', justifyContent:'space-between', fontSize:13, color:'var(--v2-muted)', flexWrap:'wrap', gap:16}}>
          <div>© 2026 AISec Fabric. Все права защищены.</div>
          <div style={{display:'flex', gap:24}}>
            <a href="#" style={{color:'var(--v2-muted)', textDecoration:'none'}}>Политика конфиденциальности</a>
            <a href="#" style={{color:'var(--v2-muted)', textDecoration:'none'}}>Условия использования</a>
            <span style={{color:'var(--v2-red)'}}>● ALL SYSTEMS OPERATIONAL</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

// "Вас взломали?" floating CTA — top right
function V2AlertCorner() {
  return (
    <div style={{position:'absolute', top:24, right:40, zIndex:5}}>
      <a href="#" className="v2-alert-pill">
        Вас атаковали?
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 2L1 21h22L12 2z"/><path d="M12 9v4M12 17h.01"/>
        </svg>
      </a>
    </div>
  );
}

// Animated red ribbon background — three layers rotating + breathing
function V2RibbonBg({ variant = 0 }) {
  const layers = 18;
  const colors = ['#ff0d2c', '#cc0a23', '#990818'];

  return (
    <svg viewBox="0 0 800 800" preserveAspectRatio="xMidYMid slice"
      style={{position:'absolute', inset:0, width:'100%', height:'100%', pointerEvents:'none', willChange:'transform'}}>
      <defs>
        <radialGradient id={`ribGlow-${variant}`} cx="55%" cy="50%" r="55%">
          <stop offset="0%" stopColor="#ff0d2c" stopOpacity="0.5"/>
          <stop offset="50%" stopColor="#660510" stopOpacity="0.25"/>
          <stop offset="100%" stopColor="#000" stopOpacity="0"/>
        </radialGradient>
      </defs>
      <rect width="800" height="800" fill={`url(#ribGlow-${variant})`} className="v2-ribbon-layer-3"/>

      {/* Layer 1 — slow forward rotation */}
      <g className="v2-ribbon-layer-1" style={{mixBlendMode:'screen'}}>
        {Array.from({length: layers}).map((_, i) => {
          const t = i / layers;
          const r = 90 + i * 38 + (variant * 18);
          const cx = 480 + Math.sin(i*0.4 + variant) * 30;
          const cy = 440 + Math.cos(i*0.35 + variant) * 26;
          const sw = 2.5 - t*1.6;
          const opacity = (1 - t) * 0.75;
          const rx = r * (1 + Math.sin(i*0.7) * 0.18);
          const ry = r * (1 + Math.cos(i*0.5) * 0.12);
          const rot = i * 14 + variant*22;
          return (
            <ellipse key={`a${i}`} cx={cx} cy={cy} rx={rx} ry={ry}
              fill="none"
              stroke={colors[i % colors.length]}
              strokeWidth={sw}
              opacity={opacity}
              transform={`rotate(${rot} ${cx} ${cy})`}/>
          );
        })}
      </g>

      {/* Layer 2 — slow reverse rotation, offset radii */}
      <g className="v2-ribbon-layer-2" style={{mixBlendMode:'screen'}}>
        {Array.from({length: layers}).map((_, i) => {
          const t = i / layers;
          const r = 130 + i * 36 + (variant * 12);
          const cx = 480 + Math.sin(i*0.5 + variant + 1.3) * 25;
          const cy = 440 + Math.cos(i*0.45 + variant + 0.7) * 22;
          const sw = 2.0 - t*1.2;
          const opacity = (1 - t) * 0.55;
          const rx = r * (1 + Math.cos(i*0.6) * 0.20);
          const ry = r * (1 + Math.sin(i*0.55) * 0.14);
          const rot = i * -11 + variant*-18;
          return (
            <ellipse key={`b${i}`} cx={cx} cy={cy} rx={rx} ry={ry}
              fill="none"
              stroke={colors[(i+1) % colors.length]}
              strokeWidth={sw}
              opacity={opacity}
              transform={`rotate(${rot} ${cx} ${cy})`}/>
          );
        })}
      </g>
    </svg>
  );
}

// ─── Shared product data ───────────────────────────────────────────────
const PRODUCTS = [
  {
    id: 'gateway',
    code: 'AISec Gateway',
    title: 'Защита AI-трафика',
    desc: 'Перехват и контроль каждого запроса к внешним и внутренним LLM. Streaming-aware, fail-close для регулируемых классов.',
    products: ['AISec Gateway','AISec Edge Proxy','AISec Cloud Gateway'],
    icon: 'shield',
  },
  {
    id: 'shadow',
    code: 'Shadow AI Discovery',
    title: 'Обнаружение теневого ИИ',
    desc: 'Карта всех LLM, агентов, MCP-серверов и API-ключей в вашем периметре. Без агентов на машинах.',
    products: ['AISec Discovery','AISec eBPF Sensor'],
    icon: 'radar',
  },
  {
    id: 'dlp',
    code: 'AI DLP',
    title: 'Защита данных в LLM',
    desc: 'Smart redaction PII, секретов и конфиденциальных данных в потоке. Готовые правила под российский периметр.',
    products: ['AISec DLP','AISec Prompt Vault'],
    icon: 'lock',
  },
  {
    id: 'graph',
    code: 'AI Asset Graph',
    title: 'Граф ИИ-активов',
    desc: 'Люди, модели, агенты, MCP-серверы, RAG-индексы, API-ключи — first-class сущности с связями.',
    products: ['AISec Asset Graph','AISec NHI Registry'],
    icon: 'graph',
  },
  {
    id: 'policy',
    code: 'Policy & Compliance',
    title: 'Политики и compliance',
    desc: 'Visual Policy Builder для CISO. Dry-run на исторических данных перед block-режимом.',
    products: ['AISec Policy Engine','AISec Compliance Suite'],
    icon: 'doc',
  },
  {
    id: 'audit',
    code: 'Audit & Evidence',
    title: 'Аудит и evidence-паки',
    desc: 'Подписанный иммутабельный аудит-лог. Evidence pack PDF для ФСТЭК, 152-ФЗ и КИИ в один клик.',
    products: ['AISec Audit','AISec SIEM Bridge'],
    icon: 'eye',
  },
];

const SECTION_ICONS = {
  shield: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <path d="M24 6L8 12v12c0 10 16 18 16 18s16-8 16-18V12L24 6z"/>
    <circle cx="24" cy="24" r="4" fill="currentColor" opacity="0.4"/>
    <path d="M24 16v4M20 24h8" opacity="0.6"/>
  </svg>,
  radar: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <circle cx="24" cy="24" r="18"/>
    <circle cx="24" cy="24" r="12"/>
    <circle cx="24" cy="24" r="6"/>
    <line x1="24" y1="24" x2="36" y2="14"/>
    <circle cx="36" cy="14" r="2" fill="currentColor"/>
  </svg>,
  lock: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <rect x="10" y="22" width="28" height="20" rx="2"/>
    <path d="M16 22V14a8 8 0 0116 0v8"/>
    <circle cx="24" cy="32" r="3" fill="currentColor" opacity="0.4"/>
  </svg>,
  graph: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <circle cx="12" cy="12" r="4"/><circle cx="36" cy="14" r="4"/>
    <circle cx="14" cy="36" r="4"/><circle cx="38" cy="34" r="4"/>
    <circle cx="24" cy="24" r="4" fill="currentColor" opacity="0.3"/>
    <path d="M14 14l8 8m4 0l8-8m-8 8l-8 10m8-10l10 8"/>
  </svg>,
  doc: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <path d="M14 6h16l8 8v28H14V6z"/>
    <path d="M30 6v8h8"/>
    <path d="M19 22h14M19 28h14M19 34h10"/>
  </svg>,
  eye: <svg width="42" height="42" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.2">
    <path d="M4 24s8-12 20-12 20 12 20 12-8 12-20 12S4 24 4 24z"/>
    <circle cx="24" cy="24" r="6"/>
    <circle cx="24" cy="24" r="2" fill="currentColor"/>
  </svg>,
};

// Wireframe sphere — JS 3D projection isolated via React.memo + ref-based DOM updates
const V2Sphere = React.memo(function V2Sphere({ size = 400 }) {
  // Precompute points once
  const points = uM(() => {
    const N = 160;
    const pts = [];
    for (let i = 0; i < N; i++) {
      const phi = Math.acos(1 - 2 * (i + 0.5) / N);
      const theta = Math.PI * (1 + Math.sqrt(5)) * i;
      pts.push([
        Math.sin(phi) * Math.cos(theta),
        Math.sin(phi) * Math.sin(theta),
        Math.cos(phi),
      ]);
    }
    return pts;
  }, []);

  // Compute edges between neighbours
  const edges = uM(() => {
    const e = [];
    const thr = 0.32;
    for (let i = 0; i < points.length; i++) {
      for (let j = i+1; j < points.length; j++) {
        const dx = points[i][0] - points[j][0];
        const dy = points[i][1] - points[j][1];
        const dz = points[i][2] - points[j][2];
        const d = Math.sqrt(dx*dx + dy*dy + dz*dz);
        if (d < thr) e.push([i, j]);
      }
    }
    return e;
  }, [points]);

  const r = size / 2 - 8;
  const cx = size / 2, cy = size / 2;

  // Initial projection (rot=0) so first paint already shows a proper sphere
  const initProj = points.map(([x, y, z]) => {
    const depth = (z + 1) / 2;
    return {
      cx: cx + x * r,
      cy: cy + y * r,
      r: 1 + depth * 1.6,
      opacity: 0.22 + depth * 0.7,
      z,
    };
  });

  // Refs to dot circles + edge lines so we can update them directly without React re-renders
  const dotRefs = uR([]);
  const edgeRefs = uR([]);
  const rafRef = uR(0);
  const lastFrame = uR(0);

  uE(() => {
    let rot = 0;
    let raf;
    const animate = (ts) => {
      // Throttle to ~40fps
      if (ts - lastFrame.current > 25) {
        rot += 0.012;
        const cosR = Math.cos(rot), sinR = Math.sin(rot);
        const cosT = Math.cos(rot * 0.3), sinT = Math.sin(rot * 0.3);

        // Project all points
        const projected = points.map(([x, y, z]) => {
          // Y rotation
          const x1 = x * cosR - z * sinR;
          const z1 = x * sinR + z * cosR;
          // Small X-tilt for depth
          const y1 = y * cosT - z1 * sinT;
          const z2 = y * sinT + z1 * cosT;
          return [cx + x1 * r, cy + y1 * r, z2];
        });

        // Update dot positions + sizes/opacities
        for (let i = 0; i < projected.length; i++) {
          const [px, py, pz] = projected[i];
          const depth = (pz + 1) / 2;
          const el = dotRefs.current[i];
          if (el) {
            el.setAttribute('cx', px);
            el.setAttribute('cy', py);
            el.setAttribute('r', 1 + depth * 1.6);
            el.setAttribute('opacity', 0.22 + depth * 0.7);
          }
        }
        // Update edges
        for (let i = 0; i < edges.length; i++) {
          const [a, b] = edges[i];
          const [ax, ay, az] = projected[a];
          const [bx, by, bz] = projected[b];
          const dA = (az + 1) / 2, dB = (bz + 1) / 2;
          const minDepth = Math.min(dA, dB);
          const el = edgeRefs.current[i];
          if (el) {
            el.setAttribute('x1', ax);
            el.setAttribute('y1', ay);
            el.setAttribute('x2', bx);
            el.setAttribute('y2', by);
            el.setAttribute('opacity', 0.04 + minDepth * 0.20);
          }
        }
        lastFrame.current = ts;
      }
      raf = requestAnimationFrame(animate);
    };
    raf = requestAnimationFrame(animate);
    rafRef.current = raf;
    return () => cancelAnimationFrame(raf);
  }, [points, edges, cx, cy, r]);

  return (
    <div className="v2-sphere-wrap" style={{display:'inline-block', position:'relative'}}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}
        style={{filter:'drop-shadow(0 0 80px rgba(255,13,44,0.4))'}}>
        <defs>
          <radialGradient id="sphereGlowMemo" cx="50%" cy="50%" r="50%">
            <stop offset="60%" stopColor="rgba(255,13,44,0)" />
            <stop offset="95%" stopColor="rgba(255,13,44,0.5)" />
            <stop offset="100%" stopColor="rgba(255,13,44,0)" />
          </radialGradient>
        </defs>
        {/* Outer ring */}
        <circle cx={cx} cy={cy} r={r+2} fill="none" stroke="url(#sphereGlowMemo)" strokeWidth="6"/>
        <circle cx={cx} cy={cy} r={r} fill="none" stroke="rgba(255,13,44,0.25)" strokeWidth="1"/>

        {/* Edges (animated via refs) */}
        <g style={{mixBlendMode:'screen'}}>
          {edges.map(([a, b], k) => {
            const A = initProj[a], B = initProj[b];
            const minDepth = Math.min((A.z+1)/2, (B.z+1)/2);
            return <line key={`e${k}`}
              ref={el => (edgeRefs.current[k] = el)}
              x1={A.cx} y1={A.cy} x2={B.cx} y2={B.cy}
              stroke="#ff0d2c" strokeWidth="0.6" opacity={0.04 + minDepth * 0.20}/>;
          })}
        </g>

        {/* Dots (animated via refs) */}
        {initProj.map((p, i) => (
          <circle key={`d${i}`}
            ref={el => (dotRefs.current[i] = el)}
            cx={p.cx} cy={p.cy} r={p.r}
            fill="#ff0d2c" opacity={p.opacity}/>
        ))}

        {/* Static center logo on top */}
        <g>
          <rect x={cx-24} y={cy-24} width="48" height="48" rx="7" fill="#ff0d2c"/>
          <text x={cx} y={cy+8} textAnchor="middle" fill="#fff"
            fontFamily="Onest, sans-serif" fontSize="20" fontWeight="700" letterSpacing="-0.04em">ai</text>
        </g>
      </svg>
    </div>
  );
});

function V2OrbitCard({ p, idx, active, setActive, align }) {
  const isActive = active === idx;
  const isLeft = align === 'right';
  return (
    <div
      onMouseEnter={() => setActive(idx)}
      onClick={() => setActive(idx)}
      style={{
        cursor:'pointer',
        width: 260,
        textAlign: isLeft ? 'right' : 'left',
        transition:'all .25s',
        position:'relative',
      }}>
      <div style={{
        width:56, height:56,
        background:'#0f0f10',
        border:`1px solid ${isActive ? 'var(--v2-red)' : 'var(--v2-border-soft)'}`,
        borderRadius:12,
        display:'flex', alignItems:'center', justifyContent:'center',
        color: isActive ? 'var(--v2-red)' : 'var(--v2-fg-2)',
        marginBottom:12,
        marginLeft: isLeft ? 'auto' : 0,
        transition:'border-color .2s, color .2s, box-shadow .2s',
        boxShadow: isActive ? '0 0 0 6px rgba(255,13,44,0.08)' : 'none',
      }}>
        {SECTION_ICONS[p.icon]}
      </div>

      <div style={{
        fontSize:17, fontWeight:600, letterSpacing:'-0.015em',
        color: isActive ? 'var(--v2-fg)' : 'var(--v2-fg-2)',
        marginBottom: 6, lineHeight:1.2,
      }}>
        {p.title}
      </div>
      <div style={{fontSize:13, color:'var(--v2-muted)', lineHeight:1.5}}>
        {p.code}
      </div>
    </div>
  );
}

Object.assign(window, { V2Logo, V2Nav, V2Footer, V2AlertCorner, V2RibbonBg, V2Sphere, V2OrbitCard, SECTION_ICONS, PRODUCTS });
