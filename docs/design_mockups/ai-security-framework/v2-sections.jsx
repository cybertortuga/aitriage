// v2 Sections — orbital products, console screenshots, architecture, compliance

const { useState: sS, useEffect: sE, useMemo: sM, useRef: sR } = React;

// ───── Orbital products section (PT-style: 6 corner positions around sphere) ───
const ORBITAL_POSITIONS = [
  // TL — top-left
  {style: {top:0,     left:0},                                 align:'left',   anchor:[0.18, 0.20]},
  // T — top center
  {style: {top:0,     left:'50%', transform:'translateX(-50%)'}, align:'center', anchor:[0.5,  0.04]},
  // TR — top-right
  {style: {top:0,     right:0},                                align:'right',  anchor:[0.82, 0.20]},
  // BR — bottom-right
  {style: {bottom:0,  right:0},                                align:'right',  anchor:[0.82, 0.80]},
  // B — bottom center
  {style: {bottom:0,  left:'50%', transform:'translateX(-50%)'},align:'center', anchor:[0.5,  0.96]},
  // BL — bottom-left
  {style: {bottom:0,  left:0},                                 align:'left',   anchor:[0.18, 0.80]},
];

function V2OrbitalSection() {
  const [active, setActive] = sS(0);
  const sphereSize = 460;

  const [auto, setAuto] = sS(true);
  sE(() => {
    if (!auto) return;
    const id = setInterval(() => setActive(a => (a + 1) % PRODUCTS.length), 5500);
    return () => clearInterval(id);
  }, [auto]);

  const sectionHeight = 760;
  const sectionWidth = 1240;

  return (
    <section id="products" className="v2-section" style={{position:'relative', paddingTop:80, paddingBottom:80}}>
      <div style={{position:'absolute', top:'50%', left:'50%', transform:'translate(-50%, -50%)',
        width:1400, height:1400,
        background:'radial-gradient(circle, rgba(255,13,44,0.10), rgba(255,13,44,0.04) 35%, transparent 60%)',
        pointerEvents:'none'}}/>

      <div className="v2-container" style={{position:'relative'}}>
        <div style={{textAlign:'center', marginBottom:32}}>
          <div className="v2-eyebrow-red" style={{marginBottom:24, justifyContent:'center', display:'inline-flex'}}>Платформа</div>
          <h2 className="v2-h2" style={{maxWidth:880, margin:'0 auto'}}>
            Шесть направлений защиты ИИ.<br/>
            <span style={{color:'var(--v2-muted)'}}>Один контур контроля.</span>
          </h2>
        </div>

        <div
          onMouseLeave={() => setAuto(true)}
          onMouseEnter={() => setAuto(false)}
          style={{
            position:'relative',
            maxWidth: sectionWidth,
            margin:'0 auto',
            height: sectionHeight,
          }}>

          {/* Connector lines from sphere to each card anchor */}
          <svg style={{position:'absolute', inset:0, width:'100%', height:'100%', pointerEvents:'none', zIndex:1}}
            viewBox={`0 0 ${sectionWidth} ${sectionHeight}`} preserveAspectRatio="none">
            {ORBITAL_POSITIONS.map((pos, i) => {
              const cx = sectionWidth / 2, cy = sectionHeight / 2;
              const tx = pos.anchor[0] * sectionWidth;
              const ty = pos.anchor[1] * sectionHeight;
              // Line from sphere edge to card anchor
              const dx = tx - cx, dy = ty - cy;
              const dist = Math.sqrt(dx*dx + dy*dy);
              const ux = dx / dist, uy = dy / dist;
              const x1 = cx + ux * (sphereSize/2 + 6);
              const y1 = cy + uy * (sphereSize/2 + 6);
              const x2 = tx - ux * 40;
              const y2 = ty - uy * 40;
              const isActive = active === i;
              return (
                <g key={i}>
                  <line x1={x1} y1={y1} x2={x2} y2={y2}
                    stroke={isActive ? '#ff0d2c' : '#2a2a2c'}
                    strokeWidth={isActive ? 1.5 : 1}
                    strokeDasharray="4 5"
                    opacity={isActive ? 0.9 : 0.45}/>
                  <circle cx={x2} cy={y2} r={isActive ? 4 : 3} fill={isActive ? '#ff0d2c' : '#555558'}/>
                  {isActive && <PulseDot key={`pd-${i}`} from={{x: x1, y: y1}} to={{x: x2, y: y2}}/>}
                </g>
              );
            })}
          </svg>

          {/* Central sphere */}
          <div style={{position:'absolute', top:'50%', left:'50%', transform:'translate(-50%, -50%)', zIndex:2}}>
            <V2Sphere size={sphereSize}/>
          </div>

          {/* Orbital cards — fixed corner positions */}
          {ORBITAL_POSITIONS.map((pos, i) => (
            <OrbitalSlot key={i}
              pos={pos} idx={i} active={active} setActive={setActive}
              product={PRODUCTS[i]}/>
          ))}
        </div>

        {/* Bottom row of pills */}
        <div style={{display:'flex', justifyContent:'center', gap:8, marginTop:32, flexWrap:'wrap'}}>
          {PRODUCTS.map((p, i) => (
            <button key={p.id} onClick={() => setActive(i)} style={{
              padding:'10px 16px',
              background: active === i ? 'var(--v2-red-soft)' : 'transparent',
              border: `1px solid ${active === i ? 'var(--v2-red-line)' : 'var(--v2-border-soft)'}`,
              borderRadius:8,
              color: active === i ? 'var(--v2-fg)' : 'var(--v2-fg-2)',
              fontSize:13, fontWeight:500,
              cursor:'pointer', transition:'all .15s',
              letterSpacing:'-0.005em', fontFamily:'inherit',
            }}>{p.title}</button>
          ))}
        </div>
      </div>
    </section>
  );
}

function OrbitalSlot({ pos, idx, active, setActive, product }) {
  const isActive = active === idx;
  const cardWidth = isActive ? 320 : 240;

  // For center-top/bottom positions, content always centers; otherwise alignment by side
  const align = pos.align;

  return (
    <div
      onMouseEnter={() => setActive(idx)}
      onClick={() => setActive(idx)}
      style={{
        position:'absolute',
        ...pos.style,
        width: cardWidth,
        cursor:'pointer',
        zIndex: isActive ? 5 : 3,
        transition:'width .3s ease',
      }}>
      {isActive ? (
        <div className="v2-fade-in" style={{
          padding:'22px 24px',
          background:'rgba(15,15,16,0.95)',
          border:'1px solid var(--v2-red-line)',
          borderRadius:14,
          textAlign:'left',
          boxShadow:'0 20px 60px -20px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,13,44,0.06)',
          backdropFilter:'blur(16px)',
        }}>
          <div style={{
            width:48, height:48,
            background:'#0a0a0c',
            border:'1px solid var(--v2-red-line)',
            borderRadius:10,
            display:'flex', alignItems:'center', justifyContent:'center',
            color:'var(--v2-red)',
            marginBottom:14,
          }}>
            {React.cloneElement(SECTION_ICONS[product.icon], {width:30, height:30})}
          </div>
          <h3 style={{fontSize:17, fontWeight:600, letterSpacing:'-0.015em', margin:'0 0 8px'}}>
            {product.title}
          </h3>
          <p style={{fontSize:13, color:'var(--v2-fg-2)', lineHeight:1.5, margin:'0 0 14px'}}>
            {product.desc}
          </p>
          <div className="v2-mono" style={{fontSize:10, color:'var(--v2-muted)',
            letterSpacing:'0.12em', marginBottom:8}}>ПРОДУКТЫ</div>
          <div style={{display:'flex', flexDirection:'column', gap:4}}>
            {product.products.map(prod => (
              <div key={prod} style={{fontSize:13, color:'var(--v2-fg)', display:'flex', alignItems:'center', gap:8}}>
                <span style={{color:'var(--v2-red)'}}>›</span>{prod}
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div style={{
          display:'flex', flexDirection:'column',
          alignItems: align === 'right' ? 'flex-end' : align === 'center' ? 'center' : 'flex-start',
          gap:12,
        }}>
          <div style={{
            width:56, height:56,
            background:'#0f0f10',
            border:'1px solid var(--v2-border-soft)',
            borderRadius:12,
            display:'flex', alignItems:'center', justifyContent:'center',
            color:'var(--v2-fg-2)',
            transition:'border-color .2s, color .2s',
          }}>
            {SECTION_ICONS[product.icon]}
          </div>
          <div style={{textAlign: align}}>
            <div style={{
              fontSize:16, fontWeight:600, letterSpacing:'-0.015em',
              color:'var(--v2-fg)',
              marginBottom: 4, lineHeight:1.2,
            }}>
              {product.title}
            </div>
            <div style={{fontSize:13, color:'var(--v2-muted)'}}>
              {product.code}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function PulseDot({ from, to }) {
  return (
    <circle r="3.5" fill="#ff0d2c" opacity="0.9">
      <animate attributeName="cx" from={from.x} to={to.x} dur="1.8s" repeatCount="indefinite"/>
      <animate attributeName="cy" from={from.y} to={to.y} dur="1.8s" repeatCount="indefinite"/>
      <animate attributeName="opacity" values="0;1;1;0" dur="1.8s" repeatCount="indefinite"/>
    </circle>
  );
}

// ───── Console screenshots gallery ─────────────────────────────────────────
function V2ConsolePreview() {
  const tabs = [
    {id:'overview', label:'Командный центр', sub:'KPI · трафик · риски', img:'screenshots/dash-overview.png'},
    {id:'shadow', label:'Shadow AI', sub:'5 источников · инциденты', img:'screenshots/dash-shadow.png'},
    {id:'policies', label:'Visual Policy Builder', sub:'OPA Rego без кода', img:'screenshots/dash-policies.png'},
    {id:'graph', label:'AI Asset Graph', sub:'identity · models · MCP', img:'screenshots/dash-graph.png'},
    {id:'audit', label:'Audit · WORM', sub:'evidence-паки в один клик', img:'screenshots/dash-audit.png'},
  ];
  const [tab, setTab] = sS('overview');
  const cur = tabs.find(t => t.id === tab);

  return (
    <section className="v2-section" style={{paddingTop:80, paddingBottom:120,
      background:'#08080a', borderTop:'1px solid var(--v2-border-soft)', borderBottom:'1px solid var(--v2-border-soft)'}}>
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:64, alignItems:'end', marginBottom:64, flexWrap:'wrap'}}>
          <div>
            <div className="v2-eyebrow-red" style={{marginBottom:24}}>CISO Console</div>
            <h2 className="v2-h2">
              То, что видит безопасность<br/>
              <span style={{color:'var(--v2-muted)'}}>в 09:00 утра в понедельник.</span>
            </h2>
          </div>
          <p className="v2-lead" style={{maxWidth:520, marginLeft:'auto'}}>
            Командный центр с метриками, картой теневого ИИ, визуальным
            конструктором политик, графом активов и подписанным WORM-аудитом.
            Все экраны — реальные интерфейсы из продукта.
          </p>
        </div>

        {/* Tabs */}
        <div style={{display:'flex', flexWrap:'wrap', gap:8, marginBottom:24}}>
          {tabs.map(t => (
            <button key={t.id} onClick={() => setTab(t.id)} style={{
              padding:'14px 22px',
              background: tab === t.id ? '#0f0f10' : 'transparent',
              border: `1px solid ${tab === t.id ? 'var(--v2-red-line)' : 'var(--v2-border-soft)'}`,
              borderRadius:12,
              cursor:'pointer',
              textAlign:'left',
              flex:'1 1 200px',
              fontFamily:'inherit',
              transition:'all .15s',
            }}>
              <div style={{fontSize:14, fontWeight:600, color:'var(--v2-fg)', marginBottom:4}}>{t.label}</div>
              <div className="v2-mono" style={{fontSize:11, color: tab === t.id ? 'var(--v2-red)' : 'var(--v2-muted)'}}>{t.sub}</div>
            </button>
          ))}
        </div>

        {/* Screenshot area */}
        <div key={tab} className="v2-fade-in" style={{
          position:'relative',
          border:'1px solid var(--v2-border-soft)',
          borderRadius:16,
          overflow:'hidden',
          background:'#000',
          boxShadow:'0 30px 80px -20px rgba(0,0,0,0.6), 0 0 0 1px rgba(255,13,44,0.05)',
        }}>
          {/* Glow */}
          <div style={{position:'absolute', inset:-1, background:'linear-gradient(180deg, rgba(255,13,44,0.15), transparent 30%)', pointerEvents:'none'}}/>
          {/* Window chrome */}
          <div style={{display:'flex', alignItems:'center', height:38, padding:'0 14px',
            borderBottom:'1px solid #1a1a1c', gap:14, background:'#0a0a0c'}}>
            <div style={{display:'flex', gap:6}}>
              {['#ff5f57','#febc2e','#28c840'].map((c,i) =>
                <span key={i} style={{width:11, height:11, borderRadius:'50%', background:c, opacity:0.85}}/>)}
            </div>
            <div className="v2-mono" style={{flex:1, textAlign:'center', fontSize:11.5, color:'var(--v2-muted)'}}>
              fabric.console · tenant=enterprise-prod · role=CISO
            </div>
            <span className="v2-tag v2-tag-red" style={{fontSize:10, padding:'2px 8px'}}>● live</span>
          </div>
          <img src={cur.img} alt={cur.label} style={{
            width:'100%', display:'block',
            // Crop top header area (the v1 page title that's included in the screenshot)
            objectFit:'cover', objectPosition:'center 60%',
            maxHeight: 720,
          }}/>
        </div>

        {/* Inline detail strip */}
        <div style={{marginTop:32, display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:24}}>
          {[
            {n:'01', t:'Все запросы — в одной хронологии',
             d:'Каждое обращение к LLM проходит через Fabric. Identity, модель, классификация, действие, p99 — на одном экране.'},
            {n:'02', t:'Без чтения кода компонентов',
             d:'CISO видит «что нашли и что предотвратили». Конструктор политик собирает Rego под капотом — не нужно учить DSL.'},
            {n:'03', t:'Подписанные evidence-паки',
             d:'PDF для проверок РКН, ФСТЭК и ЦБ РФ собирается за один клик из неизменяемого аудит-лога. С подписью HSM.'},
          ].map(it => (
            <div key={it.n} style={{paddingTop:20, borderTop:'1px solid var(--v2-border-soft)'}}>
              <div className="v2-mono" style={{fontSize:11, color:'var(--v2-red)', marginBottom:10, letterSpacing:'0.05em'}}>{it.n}</div>
              <div style={{fontSize:15, fontWeight:600, letterSpacing:'-0.01em', marginBottom:6}}>{it.t}</div>
              <p style={{fontSize:13, color:'var(--v2-muted)', lineHeight:1.55, margin:0}}>{it.d}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ───── How it works — high-level flow without exposing stack ────────────────
function V2HowItWorks() {
  const stages = [
    {n:'01', t:'Пользователь обращается к LLM',
     d:'Сотрудник, агент или service account отправляет запрос к любой языковой модели — внешней или внутренней.'},
    {n:'02', t:'AISec Fabric перехватывает',
     d:'Каждый запрос проходит через единый защитный слой. Streaming-aware, без потери UX и без раскрытия задержек пользователю.'},
    {n:'03', t:'Классификация и контроль',
     d:'PII, секреты, конфиденциальные данные обнаруживаются и редактируются. Промпт-инъекции и jailbreak-попытки блокируются.'},
    {n:'04', t:'Политики и approval-цепочки',
     d:'Регулируемые операции уходят на approval CISO. Risk-score рассчитывается по identity, отделу и контексту запроса.'},
    {n:'05', t:'Подписанный аудит',
     d:'Каждое решение сохраняется в неизменяемый WORM-журнал. Evidence-pack для проверки собирается за один клик.'},
  ];
  const [active, setActive] = sS(2);

  return (
    <section className="v2-section">
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:64, alignItems:'end', marginBottom:64}}>
          <div>
            <div className="v2-eyebrow-red" style={{marginBottom:24}}>Как это работает</div>
            <h2 className="v2-h2">Один контур.<br/>Пять моментов истины.</h2>
          </div>
          <p className="v2-lead" style={{maxWidth:480, marginLeft:'auto'}}>
            Защита AI-периметра проектируется как полноценный security control,
            не как обёртка над одной LLM. Каждый запрос проходит пять стадий —
            от identity до подписи решения в журнале.
          </p>
        </div>

        <div style={{display:'flex', flexDirection:'column', gap:8}}>
          {stages.map((s, i) => {
            const isActive = active === i;
            return (
              <div key={s.n}
                onMouseEnter={() => setActive(i)}
                style={{
                  display:'grid',
                  gridTemplateColumns: '120px 1fr 1.4fr 80px',
                  gap:32, alignItems:'center',
                  padding:'24px 28px',
                  borderRadius:14,
                  border:`1px solid ${isActive ? 'var(--v2-red-line)' : 'var(--v2-border-soft)'}`,
                  background: isActive ? 'rgba(255,13,44,0.04)' : 'transparent',
                  cursor:'pointer',
                  transition:'all .25s',
                }}>
                <div className="v2-mono" style={{fontSize:13, color: isActive ? 'var(--v2-red)' : 'var(--v2-muted)', letterSpacing:'0.08em'}}>
                  STAGE {s.n}
                </div>
                <div style={{fontSize:19, fontWeight:600, letterSpacing:'-0.018em'}}>
                  {s.t}
                </div>
                <p style={{fontSize:14, color:'var(--v2-fg-2)', margin:0, lineHeight:1.55}}>{s.d}</p>
                <div style={{textAlign:'right', color: isActive ? 'var(--v2-red)' : 'var(--v2-muted)', fontSize:22}}>
                  {isActive ? '●' : '○'}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

// ───── Compliance strip — quick stamps ─────────────────────────────────────
function V2ComplianceStrip() {
  const items = [
    {code:'152-ФЗ', name:'Персональные данные', status:'production'},
    {code:'187-ФЗ', name:'Безопасность КИИ', status:'production'},
    {code:'ФСТЭК', name:'Приказы 21, 240', status:'roadmap'},
    {code:'ГОСТ 57580.1', name:'Финансовые операции', status:'production'},
    {code:'ISO 27001', name:'СМИБ', status:'production'},
    {code:'NIST AI RMF', name:'+ OWASP LLM Top 10', status:'production'},
  ];

  return (
    <section className="v2-section" style={{paddingTop:96, paddingBottom:96, background:'#08080a',
      borderTop:'1px solid var(--v2-border-soft)', borderBottom:'1px solid var(--v2-border-soft)'}}>
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1.6fr', gap:64, alignItems:'start', marginBottom:48}}>
          <div>
            <div className="v2-eyebrow-red" style={{marginBottom:24}}>Compliance</div>
            <h2 className="v2-h2" style={{fontSize:'clamp(32px, 3.6vw, 48px)'}}>
              Регулятор как<br/>
              <span style={{color:'var(--v2-red)'}}>защитный ров.</span>
            </h2>
          </div>
          <p className="v2-lead" style={{maxWidth:560}}>
            Неизменяемый подписанный аудит-лог, evidence-паки в один клик,
            CEF-экспорт в SIEM. ГОСТ-крипта, российские OS, HSM
            заказчика — без managed-ключей в нашем периметре. Реестр
            отечественного ПО и ФСТЭК-сертификация в Phase 2/3.
          </p>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:0,
          border:'1px solid var(--v2-border-soft)', borderRadius:14, overflow:'hidden'}}>
          {items.map((it, i) => {
            const col = i % 3, row = Math.floor(i/3);
            return (
              <div key={it.code} style={{
                padding:'28px 28px',
                borderRight: col < 2 ? '1px solid var(--v2-border-soft)' : 'none',
                borderTop: row > 0 ? '1px solid var(--v2-border-soft)' : 'none',
                background:'#0a0a0c',
                display:'flex', alignItems:'center', justifyContent:'space-between',
              }}>
                <div>
                  <div className="v2-mono" style={{fontSize:13, color:'var(--v2-red)', marginBottom:4, letterSpacing:'0.05em'}}>{it.code}</div>
                  <div style={{fontSize:15, color:'var(--v2-fg-2)'}}>{it.name}</div>
                </div>
                <span className={it.status === 'production' ? 'v2-tag v2-tag-red' : 'v2-tag'}
                  style={{fontSize:10, padding:'3px 8px'}}>{it.status}</span>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

Object.assign(window, { V2OrbitalSection, V2ConsolePreview, V2HowItWorks, V2ComplianceStrip });
