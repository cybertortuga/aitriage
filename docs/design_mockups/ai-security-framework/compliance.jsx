// Compliance + Roadmap + Risks/Mitigations + CTA

function Compliance() {
  const items = [
    {code:'152-ФЗ', name:'О персональных данных',
     status:'covered', detail:'Полная DLP-обработка ПДн на лету. Smart redaction, аудит обращений, evidence-пак для проверок РКН.'},
    {code:'187-ФЗ', name:'Безопасность КИИ',
     status:'covered', detail:'Полный offline-режим, air-gapped backup, immutable audit log с HSM-подписью.'},
    {code:'ФСТЭК', name:'Приказы 21, 240 · ГОСТ Р 56939-2024',
     status:'roadmap', detail:'Сертификация запланирована Phase 3 (Q1 2027). Реестр отечественного ПО Phase 2.'},
    {code:'ГОСТ 57580.1', name:'Финансовые операции',
     status:'roadmap', detail:'Mapping контролей под банки. Phase 2. Готовые отчёты для ЦБ РФ.'},
    {code:'ISO 27001', name:'СМИБ',
     status:'covered', detail:'Mapping политик и evidence. Audit-trail соответствует требованиям A.5–A.18.'},
    {code:'NIST AI RMF', name:'+ AI 600-1 · OWASP LLM Top 10',
     status:'partial', detail:'GOVERN, MAP, MEASURE — да. MANAGE Phase 2. OWASP LLM Top 10 — full coverage.'},
    {code:'EU AI Act', name:'Conformity assessment',
     status:'roadmap', detail:'Phase 3. Для экспорта в страны, признающие регуляторику ЕС (MENA enterprise).'},
    {code:'ISO/IEC 42001', name:'AI Management System',
     status:'roadmap', detail:'Контроли в плане Phase 3. Готовый шаблон AIMS для аудитора.'},
  ];

  return (
    <section id="compliance" className="section" style={{background:'oklch(0.155 0.012 245)', position:'relative'}}>
      <div className="container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:64, alignItems:'end', marginBottom:48}}>
          <div>
            <Eyebrow num="05 /">Compliance & регулятор</Eyebrow>
            <h2 className="h2" style={{marginTop:18}}>
              Регулятор как<br/><span style={{color:'var(--accent)'}}>защитный ров.</span>
            </h2>
          </div>
          <p style={{fontSize:15, color:'var(--fg-2)', lineHeight:1.6, maxWidth:480, marginLeft:'auto'}}>
            Иммутабельный подписанный аудит-лог, evidence-паки в один клик,
            CEF-экспорт в MaxPatrol. ГОСТ-крипта через CryptoPro CSP или GoGOST.
            Реестр отечественного ПО и ФСТЭК-сертификация — Phase 2/3.
          </p>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:0,
          border:'1px solid var(--border-soft)', borderRadius:10, overflow:'hidden'}}>
          {items.map((it, i) => {
            const col = i % 4, row = Math.floor(i/4);
            return (
              <div key={it.code} style={{
                padding:'24px 22px',
                background:'var(--bg-elev)',
                borderRight: col < 3 ? '1px solid var(--border-soft)' : 'none',
                borderTop: row > 0 ? '1px solid var(--border-soft)' : 'none',
                position:'relative', minHeight:180,
              }}>
                <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start', marginBottom:10}}>
                  <div className="mono" style={{fontSize:11, color:'var(--accent)', letterSpacing:'0.05em'}}>{it.code}</div>
                  <StatusPill s={it.status}/>
                </div>
                <div style={{fontSize:14, fontWeight:500, marginBottom:10, letterSpacing:'-0.005em'}}>{it.name}</div>
                <div style={{fontSize:12.5, color:'var(--muted)', lineHeight:1.55}}>{it.detail}</div>
              </div>
            );
          })}
        </div>

        {/* Russian-specific */}
        <div style={{marginTop:48, padding:'32px 32px', border:'1px solid var(--border-soft)', borderRadius:10,
          display:'grid', gridTemplateColumns:'1fr 1fr 1fr', gap:48, background:'var(--bg-elev)'}}>
          <div>
            <div className="bracket">RU STACK · OS</div>
            <div style={{fontSize:15, marginTop:10, lineHeight:1.65, color:'var(--fg-2)'}}>
              Astra Linux 1.7+, RED OS, ALT Linux, ROSA. Совместимость с Postgres Pro.
              Установка через Helm на Astra K8s, Deckhouse, OpenShift или single binary
              для air-gapped КИИ-сегментов.
            </div>
          </div>
          <div>
            <div className="bracket">RU CRYPTO</div>
            <div style={{fontSize:15, marginTop:10, lineHeight:1.65, color:'var(--fg-2)'}}>
              ГОСТ TLS через CryptoPro CSP или открытый GoGOST (с патчами Go).
              HSM: Thales Luna, КриптоПро HSM, YubiHSM. Customer-managed keys
              или Vault Transit — никаких managed keys в нашем периметре.
            </div>
          </div>
          <div>
            <div className="bracket">RU LLM PROVIDERS</div>
            <div style={{fontSize:15, marginTop:10, lineHeight:1.65, color:'var(--fg-2)'}}>
              GigaChat, YandexGPT, T-Lite, MTS AI, SaluteSpeech, Yandex SpeechKit.
              Прямые интеграции, без proxy через зарубежные шлюзы.
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function StatusPill({s}) {
  const map = {
    covered: {l:'production', c:'var(--accent)', bg:'oklch(0.91 0.20 130 / 0.12)'},
    partial: {l:'partial', c:'var(--warn)', bg:'oklch(0.85 0.16 75 / 0.14)'},
    roadmap: {l:'roadmap', c:'var(--muted)', bg:'oklch(0.62 0.012 245 / 0.14)'},
  };
  const m = map[s];
  return <span className="mono" style={{
    fontSize:10, padding:'3px 8px', borderRadius:3,
    background:m.bg, color:m.c, letterSpacing:'0.05em', textTransform:'uppercase'
  }}>{m.l}</span>;
}

// ROADMAP
function Roadmap() {
  const phases = [
    {n:'PHASE 01', t:'0–6 месяцев', name:'MVP. Защита запросов.',
     status:'now',
     items:[
       'AI Gateway · Envoy + Go ext_proc',
       'DLP с 30–50 правилами под РФ',
       'Shadow AI Discovery (non-invasive)',
       'Asset Graph через Apache AGE',
       'Visual Policy Builder',
       'CEF / Syslog / NATS / Kafka bridge',
       'RBAC + OIDC/SAML/AD + SCIM',
       'HA через Patroni + offline-режим',
     ],
     deliverable:'3–5 платных пилотов · 1.5–5 млн ₽ · 2–3 месяца',
    },
    {n:'PHASE 02', t:'6–12 месяцев', name:'Удержание и upsell.',
     status:'next',
     items:[
       'MCP & Skills Registry с manifest + signing',
       'AIBOM в CycloneDX 1.6 ML',
       'RAG Security (ingestion + retrieval DLP)',
       'Local Model Router (vLLM, Ollama, TGI)',
       'Browser Extension (Chromium)',
       'Tetragon-based eBPF discovery',
       'TypeScript SDK + K8s sidecar',
       'CI/CD интеграция (GitLab, GH Actions)',
       'Prompt Vault',
       'Реестр отечественного ПО',
     ],
     deliverable:'8–12 контрактов · ARR 100–200 млн ₽',
    },
    {n:'PHASE 03', t:'12–24 месяца', name:'Дифференциация.',
     status:'later',
     items:[
       'Agent Runtime Security (Agent EDR)',
       'gVisor + Kata sandbox',
       'Human-in-the-loop · kill switch',
       'Red Team Automation (Garak + PyRIT)',
       'EU AI Act conformity',
       'NIST AI RMF + ISO/IEC 42001',
       'Federated air-gapped deployment',
       'Multi-region SaaS · cell-based',
       'ФСТЭК-сертификация → выход в КИИ и гос',
     ],
     deliverable:'ARR 400–800 млн ₽ · СНГ-экспансия',
    },
  ];

  return (
    <section className="section" style={{position:'relative'}}>
      <div className="container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:64, alignItems:'end', marginBottom:64}}>
          <div>
            <Eyebrow num="06 /">Roadmap</Eyebrow>
            <h2 className="h2" style={{marginTop:18}}>
              Phase 1 заморожен.<br/>
              <span style={{color:'var(--muted)'}}>Архитектура определена. Дальше — execution.</span>
            </h2>
          </div>
          <p style={{fontSize:15, color:'var(--fg-2)', lineHeight:1.6, maxWidth:480, marginLeft:'auto'}}>
            Окно для входа в категорию AI Security Gateway сужается — Palo Alto консолидирует
            через Prisma AIRS (Protect AI + Portkey + Koi). У нас 12–18 месяцев чтобы занять
            СНГ-нишу. Phase 1 заморожен и защищаем.
          </p>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:24}}>
          {phases.map((p, i) => (
            <div key={p.n} className="hover-lift" style={{
              padding:'28px 26px',
              background:'var(--bg-elev)',
              border:`1px solid ${p.status==='now' ? 'var(--accent-border)' : 'var(--border-soft)'}`,
              borderRadius:10,
              position:'relative',
            }}>
              {p.status === 'now' && (
                <div style={{position:'absolute', top:-1, right:-1, background:'var(--accent)', color:'var(--accent-ink)',
                  padding:'4px 10px', fontFamily:'var(--font-mono)', fontSize:10, fontWeight:600, letterSpacing:'0.05em',
                  borderRadius:'0 9px 0 6px'}}>NOW · 2026 Q2–Q4</div>
              )}
              <div style={{display:'flex', justifyContent:'space-between', alignItems:'baseline', marginBottom:6}}>
                <span className="mono" style={{fontSize:11, color: p.status==='now' ? 'var(--accent)' : 'var(--muted)', letterSpacing:'0.08em'}}>{p.n}</span>
                <span className="mono" style={{fontSize:11, color:'var(--muted)'}}>{p.t}</span>
              </div>
              <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:20}}>{p.name}</h3>

              <div style={{display:'flex', flexDirection:'column', gap:0, marginBottom:24}}>
                {p.items.map((it,j) => (
                  <div key={j} style={{display:'flex', gap:10, fontSize:13, padding:'8px 0',
                    borderTop: j>0 ? '1px solid var(--border-soft)' : 'none', color:'var(--fg-2)'}}>
                    <span className="mono" style={{color:'var(--muted)', fontSize:11, minWidth:24}}>{String(j+1).padStart(2,'0')}</span>
                    <span style={{flex:1}}>{it}</span>
                  </div>
                ))}
              </div>

              <div style={{paddingTop:16, borderTop:'1px solid var(--border)'}}>
                <div className="mono" style={{fontSize:10, color:'var(--muted)', letterSpacing:'0.08em', marginBottom:6}}>DELIVERABLE</div>
                <div style={{fontSize:13, color:'var(--fg)', lineHeight:1.5}}>{p.deliverable}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// PRINCIPLES — concise restatement
function Principles() {
  const items = [
    {n:'01', t:'Tiered detection всегда',
     d:'Fast path (AC + regex) под 5ms. ML async через NATS с aggressive timeout. Никаких sync-вызовов Python из critical path.'},
    {n:'02', t:'Fail-close только для регулируемых классов',
     d:'ПДн, секреты, гостайна — блок. Low-risk — fail-open + alert. Один сбой ML не кладёт всех клиентов.'},
    {n:'03', t:'Streaming-aware с дня 1',
     d:'Каждая фича проектируется с учётом SSE. Никогда не ломаем JSON. Редакция через [REDACTED:CATEGORY].'},
    {n:'04', t:'Policy simulator обязателен',
     d:'Любое правило проходит dry-run на исторических данных. В block-режим — только после approval CISO.'},
    {n:'05', t:'Stateless gateway',
     d:'Никакого локального состояния в ext_proc. Counters в Valkey. Секреты через Vault. RO root FS, drop ALL caps.'},
    {n:'06', t:'Customer-controlled crypto',
     d:'Customer-managed keys через HSM заказчика (Thales, КриптоПро, YubiHSM) или Vault Transit. Никаких managed keys у нас.'},
    {n:'07', t:'Identity-first',
     d:'Не только люди — service accounts, агенты, API-ключи, MCP-серверы. Всё first-class в графе.'},
    {n:'08', t:'Visual Policy Builder обязателен',
     d:'CISO не пишут Rego. Блочный конструктор поверх OPA — без него DSL не примут.'},
  ];

  return (
    <section className="section-sm" style={{position:'relative', borderTop:'1px solid var(--border-soft)', borderBottom:'1px solid var(--border-soft)'}}>
      <div className="container">
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-end', marginBottom:40, flexWrap:'wrap', gap:24}}>
          <div>
            <Eyebrow num="07 /">Принципы</Eyebrow>
            <h2 className="h2" style={{marginTop:18, fontSize:'clamp(28px, 3vw, 40px)'}}>
              Восемь правил, которые не нарушаются.
            </h2>
          </div>
        </div>
        <div style={{display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:32}}>
          {items.map(p => (
            <div key={p.n} style={{borderTop:'1px solid var(--border)', paddingTop:18}}>
              <div className="mono" style={{fontSize:11, color:'var(--accent)', marginBottom:10, letterSpacing:'0.05em'}}>{p.n}</div>
              <div style={{fontSize:15, fontWeight:500, marginBottom:8, letterSpacing:'-0.01em'}}>{p.t}</div>
              <div style={{fontSize:12.5, color:'var(--muted)', lineHeight:1.55}}>{p.d}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// CTA
function CTA() {
  return (
    <section id="contact" className="section" style={{position:'relative', overflow:'hidden'}}>
      <div style={{position:'absolute', inset:0, background:'radial-gradient(circle at 50% 100%, oklch(0.91 0.20 130 / 0.08), transparent 70%)', pointerEvents:'none'}}/>
      <div className="container" style={{position:'relative'}}>
        <div style={{
          border:'1px solid var(--border)',
          borderRadius:12,
          background:'var(--bg-elev)',
          padding:'56px 56px',
          display:'grid', gridTemplateColumns:'1.4fr 1fr', gap:48, alignItems:'center'
        }}>
          <div>
            <Eyebrow>Pilot package · 1.5–5 млн ₽ · 2–3 месяца</Eyebrow>
            <h2 className="h2" style={{marginTop:18, marginBottom:20, fontSize:'clamp(32px, 4vw, 48px)'}}>
              Установка за 1–2 дня.<br/>
              <span style={{color:'var(--muted)'}}>Helm chart или single binary.</span>
            </h2>
            <p style={{fontSize:15, color:'var(--fg-2)', lineHeight:1.6, marginBottom:32, maxWidth:520}}>
              SSO/AD с первого дня. 10 популярных провайдеров. 30+ DLP-правил под РФ.
              Visual Policy Builder. SIEM-экспорт в MaxPatrol. CISO Dashboard с цифрами
              «что нашли и что предотвратили». Отчёт за пилот с ROI-расчётом.
            </p>
            <div style={{display:'flex', gap:12, flexWrap:'wrap'}}>
              <a href="#" className="btn btn-primary">Запросить пилот →</a>
              <a href="#" className="btn btn-ghost">Скачать архитектурный whitepaper</a>
            </div>
          </div>
          <div>
            <div className="card-flat" style={{padding:'24px 26px', background:'var(--bg)'}}>
              <div className="mono" style={{fontSize:11, color:'var(--muted)', letterSpacing:'0.1em', marginBottom:18}}>ЧЕК-ЛИСТ ПЕРВОГО ПИЛОТА</div>
              {[
                'Установка ≤ 2 дней (Helm / Compose)',
                'SSO/AD интеграция',
                '10+ провайдеров: GigaChat, YGPT, GPT, Claude…',
                'RU DLP: 30+ правил, customer словари',
                'Visual policy builder + dry-run',
                'CEF SIEM-экспорт (MaxPatrol)',
                'CISO Dashboard',
                'Smart redaction в SSE без breaking JSON',
                'HA story для prod (Patroni + Envoy a/a)',
                'Отчёт за период с ROI',
              ].map((it, i) => (
                <div key={i} style={{display:'flex', alignItems:'center', gap:10, padding:'7px 0', borderTop: i>0?'1px solid var(--border-soft)':'none'}}>
                  <span style={{width:16, height:16, borderRadius:3, background:'oklch(0.91 0.20 130 / 0.15)',
                    color:'var(--accent)', display:'inline-flex', alignItems:'center', justifyContent:'center', fontSize:11}}>✓</span>
                  <span style={{fontSize:13, color:'var(--fg-2)'}}>{it}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

// Logos marquee — placeholder for customer/partner integrations
function IntegrationStrip() {
  const items = [
    'envoy proxy 1.39',
    'postgres 18.4 + apache age',
    'clickhouse 25.x',
    'nats jetstream 2.10',
    'qdrant 1.x',
    'opa rego 0.69',
    'vault 1.18',
    'authentik 2025.x',
    'sigstore + cosign',
    'opentelemetry',
    'kubernetes 1.30',
    'helm 3.x',
  ];
  const all = [...items, ...items];
  return (
    <div style={{padding:'28px 0', borderTop:'1px solid var(--border-soft)', borderBottom:'1px solid var(--border-soft)', background:'var(--bg)'}}>
      <div className="container" style={{display:'flex', alignItems:'center', gap:32}}>
        <span className="mono" style={{fontSize:11, color:'var(--muted)', letterSpacing:'0.12em', flexShrink:0}}>BUILT ON</span>
        <div className="marquee">
          <div className="marquee-track">
            {all.map((s,i) => (
              <span key={i} className="mono" style={{fontSize:13, color:'var(--fg-2)', letterSpacing:'-0.005em', whiteSpace:'nowrap'}}>
                <span style={{color:'var(--accent)'}}>+ </span>{s}
              </span>
            ))}
          </div>
          <div className="marquee-track">
            {all.map((s,i) => (
              <span key={`b${i}`} className="mono" style={{fontSize:13, color:'var(--fg-2)', letterSpacing:'-0.005em', whiteSpace:'nowrap'}}>
                <span style={{color:'var(--accent)'}}>+ </span>{s}
              </span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

Object.assign(window, { Compliance, Roadmap, Principles, CTA, IntegrationStrip });
