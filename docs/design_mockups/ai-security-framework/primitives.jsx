// Shared visual primitives for AISec Fabric
const { useState, useEffect, useRef, useMemo } = React;

// Logo mark — abstract layered shield/triangle representing 3 planes
function Logo({ size = 24 }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none" style={{display:'block'}}>
      <path d="M16 2 L29 9 V19 L16 30 L3 19 V9 Z" stroke="currentColor" strokeWidth="1.4" fill="none"/>
      <path d="M16 8 L24 12 V19 L16 25 L8 19 V12 Z" stroke="currentColor" strokeWidth="1.2" fill="none" opacity="0.6"/>
      <circle cx="16" cy="16" r="2.2" fill="var(--accent)"/>
    </svg>
  );
}

// Numbered eyebrow with optional title
function Eyebrow({ num, children }) {
  return (
    <div className="eyebrow">
      {num && <span className="mono" style={{color:'var(--accent)'}}>{num}</span>}
      <span>{children}</span>
    </div>
  );
}

// Bordered crosshair corners — technical accent on hero block
function CornerFrame({ children, style }) {
  const c = { position:'absolute', width:10, height:10, border:'1px solid var(--accent)' };
  return (
    <div style={{position:'relative', ...style}}>
      <div style={{...c, top:-1, left:-1, borderRight:'none', borderBottom:'none'}}/>
      <div style={{...c, top:-1, right:-1, borderLeft:'none', borderBottom:'none'}}/>
      <div style={{...c, bottom:-1, left:-1, borderRight:'none', borderTop:'none'}}/>
      <div style={{...c, bottom:-1, right:-1, borderLeft:'none', borderTop:'none'}}/>
      {children}
    </div>
  );
}

// Top navigation bar
function NavBar() {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    window.addEventListener('scroll', onScroll);
    return () => window.removeEventListener('scroll', onScroll);
  }, []);
  return (
    <nav style={{
      position:'sticky', top:0, zIndex:50,
      borderBottom: scrolled ? '1px solid var(--border-soft)' : '1px solid transparent',
      background: scrolled ? 'oklch(0.135 0.012 245 / 0.85)' : 'transparent',
      backdropFilter: scrolled ? 'blur(12px)' : 'none',
      transition:'all .2s ease',
    }}>
      <div className="container" style={{display:'flex', alignItems:'center', height:64, gap:32}}>
        <a href="#" style={{display:'flex', alignItems:'center', gap:10, color:'var(--fg)', textDecoration:'none'}}>
          <Logo size={22}/>
          <span style={{fontSize:15, fontWeight:600, letterSpacing:'-0.01em'}}>AISec Fabric</span>
          <span className="mono" style={{fontSize:10, color:'var(--muted)', border:'1px solid var(--border-soft)', padding:'2px 6px', borderRadius:4, marginLeft:6}}>v1.0</span>
        </a>
        <div style={{display:'flex', gap:24, marginLeft:24, fontSize:14}}>
          {['Платформа', 'Архитектура', 'Дашборд', 'Compliance', 'Документация'].map(l => (
            <a key={l} href="#" style={{color:'var(--fg-2)', textDecoration:'none'}}
               onMouseOver={e=>e.target.style.color='var(--fg)'}
               onMouseOut={e=>e.target.style.color='var(--fg-2)'}>{l}</a>
          ))}
        </div>
        <div style={{flex:1}}/>
        <a href="#" className="btn btn-link mono" style={{fontSize:12, letterSpacing:'0.05em'}}>RU / EN</a>
        <a href="#" className="btn btn-ghost">Войти</a>
        <a href="#contact" className="btn btn-primary">Запросить демо →</a>
      </div>
    </nav>
  );
}

// Footer
function Footer() {
  const cols = [
    {title:'Платформа', items:['AI Gateway','DLP-движок','Shadow AI Discovery','Policy Builder','Asset Graph','Compliance Evidence']},
    {title:'Решения', items:['Для банков','Для телеком','Для госкорпораций','Для fintech','152-ФЗ / 187-ФЗ','КИИ']},
    {title:'Ресурсы', items:['Документация','API Reference','Архитектура','SDK Python / Go','Status','Changelog']},
    {title:'Компания', items:['О нас','Партнёры','Карьера','Новости','Контакты','Безопасность']},
  ];
  return (
    <footer style={{borderTop:'1px solid var(--border-soft)', padding:'80px 0 40px'}}>
      <div className="container">
        <div style={{display:'grid', gridTemplateColumns:'1.4fr repeat(4, 1fr)', gap:48}}>
          <div>
            <div style={{display:'flex', alignItems:'center', gap:10, marginBottom:18}}>
              <Logo size={22}/>
              <span style={{fontSize:15, fontWeight:600}}>AISec Fabric</span>
            </div>
            <p style={{fontSize:13, color:'var(--muted)', lineHeight:1.6, maxWidth:280, margin:0}}>
              AI Security Control Layer для enterprise.
              Видимость, контроль данных, политики и compliance evidence
              для любых AI-систем в вашем периметре.
            </p>
            <div style={{display:'flex', gap:8, marginTop:24, flexWrap:'wrap'}}>
              <span className="tag mono">152-ФЗ</span>
              <span className="tag mono">187-ФЗ КИИ</span>
              <span className="tag mono">ФСТЭК</span>
            </div>
          </div>
          {cols.map(c => (
            <div key={c.title}>
              <div style={{fontSize:12, fontWeight:600, marginBottom:14, letterSpacing:'-0.01em'}}>{c.title}</div>
              <div style={{display:'flex', flexDirection:'column', gap:9}}>
                {c.items.map(i => <a key={i} href="#" style={{fontSize:13, color:'var(--muted)', textDecoration:'none'}}
                  onMouseOver={e=>e.target.style.color='var(--fg)'}
                  onMouseOut={e=>e.target.style.color='var(--muted)'}>{i}</a>)}
              </div>
            </div>
          ))}
        </div>
        <div style={{marginTop:64, paddingTop:24, borderTop:'1px solid var(--border-soft)',
          display:'flex', justifyContent:'space-between', fontSize:12, color:'var(--muted)'}}>
          <div className="mono">© 2026 AISec Fabric. Все права защищены.</div>
          <div style={{display:'flex', gap:24}}>
            <a href="#" style={{color:'var(--muted)', textDecoration:'none'}}>Политика конфиденциальности</a>
            <a href="#" style={{color:'var(--muted)', textDecoration:'none'}}>Условия использования</a>
            <span className="mono" style={{color:'var(--accent)'}}>● status: operational</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

Object.assign(window, { Logo, Eyebrow, CornerFrame, NavBar, Footer });
