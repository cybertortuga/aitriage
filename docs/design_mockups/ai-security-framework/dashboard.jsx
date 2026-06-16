// Live Dashboard Mock — interactive CISO console preview
const { useState: useStateD, useEffect: useEffectD, useMemo: useMemoD, useRef: useRefD } = React;

function LiveDashboard() {
  const [tab, setTab] = useStateD('overview');
  return (
    <section id="dash" className="section" style={{position:'relative'}}>
      <div className="container">
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-end', marginBottom:48, gap:32, flexWrap:'wrap'}}>
          <div>
            <Eyebrow num="04 /">CISO Console · превью</Eyebrow>
            <h2 className="h2" style={{marginTop:18, maxWidth:720}}>
              То, что видит безопасность<br/>в 09:00 утра в понедельник.
            </h2>
          </div>
          <span className="tag live">live demo · click around</span>
        </div>

        {/* App frame */}
        <div style={{border:'1px solid var(--border)', borderRadius:12, overflow:'hidden',
          background:'var(--bg-elev)', boxShadow:'0 20px 60px -20px rgba(0,0,0,0.6)'}}>
          {/* Window chrome */}
          <div style={{display:'flex', alignItems:'center', height:42, padding:'0 14px',
            borderBottom:'1px solid var(--border-soft)', gap:14}}>
            <div style={{display:'flex', gap:6}}>
              {['#ff5f57','#febc2e','#28c840'].map((c,i) =>
                <span key={i} style={{width:11, height:11, borderRadius:'50%', background:c, opacity:0.85}}/>)}
            </div>
            <div className="mono" style={{flex:1, textAlign:'center', fontSize:11.5, color:'var(--muted)'}}>
              fabric.console · tenant=sberbank-prod · region=ru-central1 · role=CISO
            </div>
            <div className="tag mono live" style={{fontSize:10, padding:'2px 8px'}}>3 live alerts</div>
          </div>

          <div style={{display:'grid', gridTemplateColumns:'220px 1fr', minHeight:680}}>
            <Sidebar tab={tab} setTab={setTab}/>
            <div style={{padding:'28px 32px', borderLeft:'1px solid var(--border-soft)', overflow:'hidden'}}>
              {tab === 'overview' && <Overview/>}
              {tab === 'shadow' && <ShadowAI/>}
              {tab === 'policies' && <Policies/>}
              {tab === 'graph' && <AssetGraph/>}
              {tab === 'audit' && <AuditLog/>}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function Sidebar({ tab, setTab }) {
  const groups = [
    {label:'OVERVIEW', items:[
      {id:'overview', n:'Командный центр', icon:'◇'},
      {id:'shadow', n:'Shadow AI', icon:'⊘', badge:7},
    ]},
    {label:'CONTROL', items:[
      {id:'policies', n:'Политики', icon:'⊞'},
      {id:'graph', n:'Asset Graph', icon:'⊜'},
    ]},
    {label:'EVIDENCE', items:[
      {id:'audit', n:'Audit Log', icon:'≡'},
    ]},
  ];
  return (
    <aside style={{padding:'20px 14px', background:'var(--bg)', position:'relative'}}>
      <div style={{display:'flex', alignItems:'center', gap:8, padding:'8px 10px', marginBottom:18}}>
        <Logo size={18}/>
        <span style={{fontWeight:600, fontSize:13}}>AISec Fabric</span>
      </div>
      {groups.map(g => (
        <div key={g.label} style={{marginBottom:18}}>
          <div className="mono" style={{fontSize:10, color:'var(--muted)',
            letterSpacing:'0.12em', padding:'8px 10px'}}>{g.label}</div>
          {g.items.map(it => (
            <button key={it.id} onClick={() => setTab(it.id)} style={{
              display:'flex', alignItems:'center', gap:10, width:'100%',
              padding:'9px 10px', background: tab===it.id ? 'var(--surface)' : 'transparent',
              border:'none', borderRadius:6, cursor:'pointer', color: tab===it.id ? 'var(--fg)' : 'var(--fg-2)',
              fontSize:13, textAlign:'left',
              borderLeft: tab===it.id ? '2px solid var(--accent)' : '2px solid transparent',
              marginLeft:-2,
            }}>
              <span style={{color: tab===it.id ? 'var(--accent)' : 'var(--muted)', width:14}}>{it.icon}</span>
              <span style={{flex:1}}>{it.n}</span>
              {it.badge && <span className="mono" style={{
                fontSize:10, padding:'2px 6px', borderRadius:10,
                background:'oklch(0.72 0.20 25 / 0.18)', color:'var(--danger)'
              }}>{it.badge}</span>}
            </button>
          ))}
        </div>
      ))}
      <div style={{position:'absolute', bottom:24, left:14, right:14, width:'192px'}}>
        <div className="card-flat" style={{padding:'12px 14px'}}>
          <div className="mono" style={{fontSize:10, color:'var(--muted)', marginBottom:6}}>QUOTA · MAY</div>
          <div style={{display:'flex', justifyContent:'space-between', fontSize:12, marginBottom:6}}>
            <span>Tokens</span><span className="mono">2.4M / 5M</span>
          </div>
          <div style={{height:3, background:'var(--surface)', borderRadius:2, overflow:'hidden'}}>
            <div style={{width:'48%', height:'100%', background:'var(--accent)'}}/>
          </div>
        </div>
      </div>
    </aside>
  );
}

// === OVERVIEW ===
function Overview() {
  return (
    <div>
      {/* Top stats row */}
      <div style={{display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:14, marginBottom:24}}>
        <KPI label="Запросов к LLM сегодня" value="142,847" delta="+12.4%" color="var(--accent)"/>
        <KPI label="Заблокировано (PII / secrets)" value="318" delta="14 secrets" color="var(--danger)"/>
        <KPI label="Cache hit ratio" value="71.2%" delta="–840 ML calls" color="var(--info)"/>
        <KPI label="Shadow AI инцидентов" value="7" delta="3 critical" color="var(--warn)"/>
      </div>

      {/* Traffic chart */}
      <div className="card" style={{padding:0}}>
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', padding:'18px 22px', borderBottom:'1px solid var(--border-soft)', gap:18}}>
          <div style={{minWidth:0, flex:1}}>
            <div style={{fontSize:14, fontWeight:500, whiteSpace:'nowrap', overflow:'hidden', textOverflow:'ellipsis'}}>Трафик AI Gateway · последние 24 часа</div>
            <div className="mono" style={{fontSize:11, color:'var(--muted)', marginTop:4}}>requests/min · по провайдерам</div>
          </div>
          <div style={{display:'flex', gap:14, fontSize:11, flexShrink:0, flexWrap:'wrap', justifyContent:'flex-end'}}>
            <LegendDot c="var(--accent)" l="GigaChat"/>
            <LegendDot c="var(--info)" l="YandexGPT"/>
            <LegendDot c="oklch(0.72 0.16 290)" l="OpenAI"/>
            <LegendDot c="var(--warn)" l="Claude"/>
            <LegendDot c="var(--danger)" l="blocked"/>
          </div>
        </div>
        <div style={{padding:'14px 14px 18px'}}>
          <TrafficChart/>
        </div>
      </div>

      {/* Two columns */}
      <div style={{display:'grid', gridTemplateColumns:'1.4fr 1fr', gap:14, marginTop:14}}>
        <RecentBlocks/>
        <RiskBoard/>
      </div>
    </div>
  );
}

function KPI({label, value, delta, color}) {
  return (
    <div className="card" style={{padding:'16px 18px'}}>
      <div style={{fontSize:11.5, color:'var(--muted)', marginBottom:8}}>{label}</div>
      <div style={{display:'flex', alignItems:'baseline', gap:10}}>
        <div style={{fontSize:26, letterSpacing:'-0.02em', fontWeight:500}}>{value}</div>
      </div>
      <div className="mono" style={{fontSize:11, color, marginTop:6}}>{delta}</div>
    </div>
  );
}

function LegendDot({c, l}) {
  return (
    <span style={{display:'inline-flex', alignItems:'center', gap:6, color:'var(--fg-2)'}}>
      <span style={{width:8, height:8, borderRadius:2, background:c}}/>{l}
    </span>
  );
}

function TrafficChart() {
  // Generate deterministic-ish multi-series data
  const data = useMemoD(() => {
    const points = 64;
    const series = ['giga','yandex','openai','claude','blocked'].map((k,si) => {
      const arr = [];
      for (let i=0;i<points;i++) {
        const base = (50 + 30*Math.sin(i*0.18 + si)) * (1.2 - si*0.15);
        const wob = Math.sin(i*0.7 + si*1.7) * 8;
        const v = Math.max(2, base + wob + (k==='blocked' ? -base*0.85 : 0));
        arr.push(v);
      }
      return {k, arr};
    });
    return series;
  }, []);
  const max = 120;
  const colors = {giga:'var(--accent)', yandex:'var(--info)', openai:'oklch(0.72 0.16 290)', claude:'var(--warn)', blocked:'var(--danger)'};
  return (
    <svg viewBox="0 0 640 160" width="100%" height="160" preserveAspectRatio="none" style={{display:'block'}}>
      {/* Gridlines */}
      {[0,1,2,3].map(i => (
        <line key={i} x1="0" y1={i*40} x2="640" y2={i*40} stroke="var(--border-soft)" strokeDasharray="2 4"/>
      ))}
      {data.map(s => {
        const pts = s.arr.map((v,i) => `${(i/(s.arr.length-1))*640},${160 - (v/max)*160}`).join(' ');
        const area = `0,160 ${pts} 640,160`;
        return (
          <g key={s.k}>
            {s.k === 'giga' && <polygon points={area} fill="var(--accent)" opacity="0.08"/>}
            <polyline points={pts} fill="none" stroke={colors[s.k]} strokeWidth={s.k==='giga'?2:1.4} opacity={s.k==='blocked'?0.9:0.85}/>
          </g>
        );
      })}
      {/* Current marker */}
      <line x1="540" y1="0" x2="540" y2="160" stroke="var(--accent)" strokeDasharray="2 3" opacity="0.5"/>
    </svg>
  );
}

function RecentBlocks() {
  const items = [
    {t:'14:32:08', user:'analyst@bank.ru', cat:'AWS_SECRET', model:'Claude-3.5', act:'BLOCK'},
    {t:'14:31:47', user:'pm@retail.ru', cat:'ИНН', model:'GigaChat-Pro', act:'REDACT'},
    {t:'14:31:22', user:'unknown', cat:'PROMPT_INJECTION', model:'YandexGPT-5', act:'BLOCK'},
    {t:'14:30:51', user:'dev@telecom.ru', cat:'SSH_PRIVATE_KEY', model:'GPT-4o', act:'BLOCK'},
    {t:'14:30:11', user:'support@bank.ru', cat:'СНИЛС', model:'GigaChat-Pro', act:'REDACT'},
    {t:'14:29:44', user:'finance@bank.ru', cat:'CARD_PAN', model:'YandexGPT-5', act:'REDACT'},
    {t:'14:28:30', user:'dev@bank.ru', cat:'JWT_SECRET', model:'GPT-4o', act:'BLOCK'},
  ];
  return (
    <div className="card" style={{padding:0}}>
      <div style={{padding:'14px 18px', borderBottom:'1px solid var(--border-soft)', display:'flex', justifyContent:'space-between'}}>
        <div style={{fontSize:14, fontWeight:500}}>Последние срабатывания DLP</div>
        <a href="#" style={{fontSize:12, color:'var(--accent)', textDecoration:'none'}}>Открыть все →</a>
      </div>
      <table style={{width:'100%', borderCollapse:'collapse', fontSize:12.5}}>
        <thead>
          <tr style={{color:'var(--muted)', textAlign:'left'}}>
            {['Время','Identity','Категория','Модель','Действие'].map(h =>
              <th key={h} style={{padding:'10px 18px', fontWeight:400, fontSize:11, letterSpacing:'0.04em', textTransform:'uppercase'}}>{h}</th>
            )}
          </tr>
        </thead>
        <tbody>
          {items.map((it, i) => (
            <tr key={i} style={{borderTop:'1px solid var(--border-soft)'}}>
              <td className="mono" style={{padding:'9px 18px', color:'var(--muted)'}}>{it.t}</td>
              <td style={{padding:'9px 18px'}}>{it.user}</td>
              <td className="mono" style={{padding:'9px 18px'}}>{it.cat}</td>
              <td style={{padding:'9px 18px', color:'var(--fg-2)'}}>{it.model}</td>
              <td style={{padding:'9px 18px'}}>
                <span className="mono" style={{
                  fontSize:10.5, padding:'3px 7px', borderRadius:3,
                  background: it.act==='BLOCK' ? 'oklch(0.72 0.20 25 / 0.18)' : 'oklch(0.85 0.16 75 / 0.18)',
                  color: it.act==='BLOCK' ? 'var(--danger)' : 'var(--warn)'
                }}>{it.act}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RiskBoard() {
  const rows = [
    {n:'Управление РКО', score:8.4, mode:'block', users:124},
    {n:'Розничный бизнес', score:6.7, mode:'monitor', users:84},
    {n:'IT / DevOps', score:5.1, mode:'monitor', users:312},
    {n:'Юр. дирекция', score:4.4, mode:'block', users:38},
    {n:'Маркетинг', score:3.2, mode:'monitor', users:67},
  ];
  return (
    <div className="card" style={{padding:0}}>
      <div style={{padding:'14px 18px', borderBottom:'1px solid var(--border-soft)'}}>
        <div style={{fontSize:14, fontWeight:500}}>Risk Score · по подразделениям</div>
        <div className="mono" style={{fontSize:11, color:'var(--muted)', marginTop:4}}>0–10 · ml + rules + history</div>
      </div>
      <div style={{padding:'10px 18px'}}>
        {rows.map(r => (
          <div key={r.n} style={{display:'grid', gridTemplateColumns:'1.6fr 60px 1fr', gap:12, padding:'10px 0', borderBottom:'1px solid var(--border-soft)', alignItems:'center'}}>
            <div>
              <div style={{fontSize:13}}>{r.n}</div>
              <div className="mono" style={{fontSize:10.5, color:'var(--muted)', marginTop:2}}>{r.users} users · {r.mode}</div>
            </div>
            <div className="mono" style={{fontSize:14, color: r.score>7 ? 'var(--danger)' : r.score>5 ? 'var(--warn)' : 'var(--accent)', textAlign:'right'}}>{r.score}</div>
            <div style={{height:4, background:'var(--surface)', borderRadius:2, overflow:'hidden'}}>
              <div style={{width:`${r.score*10}%`, height:'100%',
                background: r.score>7 ? 'var(--danger)' : r.score>5 ? 'var(--warn)' : 'var(--accent)'}}/>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// === SHADOW AI ===
function ShadowAI() {
  const incidents = [
    {sev:'critical', t:'12 пользователей шлют код в ChatGPT через chat.openai.com',
      src:'corp DNS', cnt:'1,847 req / 24h', dept:'IT / DevOps'},
    {sev:'critical', t:'Account на api.anthropic.com оплачен через корп. карту',
      src:'cloud billing', cnt:'$4,120 / месяц', dept:'неизвестно'},
    {sev:'critical', t:'AWS_BEDROCK_KEY коммитнут в gitlab.bank.ru/risk-models',
      src:'git scanner', cnt:'commit 4a8c12f', dept:'Risk-modeling'},
    {sev:'high', t:'OpenAI SDK обнаружен в зависимостях backend-payments',
      src:'SBOM scan', cnt:'openai==1.84.0', dept:'Платежи'},
    {sev:'high', t:'mistral.ai · 47 уникальных DNS-резолвов за неделю',
      src:'corp DNS', cnt:'47 hosts', dept:'смешано'},
    {sev:'medium', t:'Lobehub self-hosted на 10.4.12.88 без auth',
      src:'k8s discovery', cnt:'1 pod', dept:'неизвестно'},
  ];
  return (
    <div>
      <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:18}}>
        <div>
          <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:4}}>Shadow AI Discovery</h3>
          <div className="mono" style={{fontSize:11, color:'var(--muted)'}}>5 sources · 7 active incidents · last scan 4m ago</div>
        </div>
        <div style={{display:'flex', gap:8}}>
          <button className="btn btn-ghost" style={{fontSize:12, padding:'8px 12px'}}>Все источники</button>
          <button className="btn btn-primary" style={{fontSize:12, padding:'8px 12px'}}>Экспорт в SIEM</button>
        </div>
      </div>

      {/* Source counters */}
      <div style={{display:'grid', gridTemplateColumns:'repeat(5, 1fr)', gap:12, marginBottom:20}}>
        {[
          {n:'DNS logs', v:'1.2M', sub:'today'},
          {n:'Proxy logs', v:'420K', sub:'Squid + BlueCoat'},
          {n:'SSO events', v:'8.4K', sub:'Authentik'},
          {n:'Cloud billing', v:'$12K', sub:'AWS · Yandex'},
          {n:'Git scanner', v:'14 repos', sub:'3 findings'},
        ].map(s => (
          <div key={s.n} className="card" style={{padding:'14px 16px'}}>
            <div className="mono" style={{fontSize:10, color:'var(--muted)', marginBottom:6, letterSpacing:'0.08em', textTransform:'uppercase'}}>{s.n}</div>
            <div style={{fontSize:20, fontWeight:500, letterSpacing:'-0.015em'}}>{s.v}</div>
            <div className="mono" style={{fontSize:11, color:'var(--accent)', marginTop:4}}>{s.sub}</div>
          </div>
        ))}
      </div>

      <div className="card" style={{padding:0}}>
        <div style={{padding:'14px 18px', borderBottom:'1px solid var(--border-soft)'}}>
          <div style={{fontSize:14, fontWeight:500}}>Открытые инциденты</div>
        </div>
        {incidents.map((it, i) => (
          <div key={i} style={{display:'grid', gridTemplateColumns:'auto 1fr auto auto auto', gap:18,
            padding:'14px 18px', borderTop: i>0 ? '1px solid var(--border-soft)' : 'none', alignItems:'center'}}>
            <span className="mono" style={{
              fontSize:10, padding:'3px 8px', borderRadius:3,
              background: it.sev==='critical' ? 'oklch(0.72 0.20 25 / 0.15)'
                : it.sev==='high' ? 'oklch(0.85 0.16 75 / 0.15)'
                : 'oklch(0.62 0.012 245 / 0.20)',
              color: it.sev==='critical' ? 'var(--danger)' : it.sev==='high' ? 'var(--warn)' : 'var(--muted)',
              textTransform:'uppercase', letterSpacing:'0.05em'
            }}>{it.sev}</span>
            <div>
              <div style={{fontSize:13.5}}>{it.t}</div>
              <div className="mono" style={{fontSize:11, color:'var(--muted)', marginTop:3}}>src: {it.src} · dept: {it.dept}</div>
            </div>
            <div className="mono" style={{fontSize:12, color:'var(--fg-2)'}}>{it.cnt}</div>
            <button className="btn btn-ghost" style={{fontSize:11, padding:'5px 10px'}}>Расследовать</button>
            <button className="btn btn-ghost" style={{fontSize:11, padding:'5px 10px', borderColor:'var(--accent-border)', color:'var(--accent)'}}>Заблокировать</button>
          </div>
        ))}
      </div>
    </div>
  );
}

// === POLICIES — visual builder preview ===
function Policies() {
  return (
    <div>
      <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:18}}>
        <div>
          <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:4}}>Visual Policy Builder</h3>
          <div className="mono" style={{fontSize:11, color:'var(--muted)'}}>policy: pii_redact_card_data · compiles to OPA Rego · dry-run on 30d</div>
        </div>
        <div style={{display:'flex', gap:8}}>
          <button className="btn btn-ghost" style={{fontSize:12, padding:'8px 12px'}}>Dry-run</button>
          <button className="btn btn-primary" style={{fontSize:12, padding:'8px 12px'}}>Запросить approval CISO</button>
        </div>
      </div>

      <div style={{display:'grid', gridTemplateColumns:'1.4fr 1fr', gap:14}}>
        <div className="card" style={{padding:0, position:'relative', overflow:'hidden'}}>
          <div className="dotgrid" style={{position:'absolute', inset:0, opacity:0.4}}/>
          <div style={{position:'relative', padding:'22px 24px'}}>
            <div style={{display:'flex', flexDirection:'column', gap:10}}>
              <PolicyBlock kind="when" title="WHEN" subtitle="request.kind == prompt">
                identity.dept ∈ [&quot;РКО&quot;, &quot;Розница&quot;, &quot;Юр&quot;]
              </PolicyBlock>
              <Connector/>
              <PolicyBlock kind="and" title="AND" subtitle="content match">
                detect.category ∈ [card_pan, ИНН, СНИЛС]
              </PolicyBlock>
              <Connector/>
              <PolicyBlock kind="and" title="AND NOT" subtitle="exclusion list">
                identity.user ∉ allowlist::compliance_team
              </PolicyBlock>
              <Connector/>
              <PolicyBlock kind="then" title="THEN" subtitle="action">
                <div style={{display:'flex', gap:8, flexWrap:'wrap'}}>
                  <span className="tag mono" style={{borderColor:'var(--accent-border)', color:'var(--accent)'}}>REDACT [REDACTED:PII]</span>
                  <span className="tag mono">LOG → audit_signed</span>
                  <span className="tag mono">METRIC dlp.card_redact++</span>
                </div>
              </PolicyBlock>
              <Connector/>
              <PolicyBlock kind="else" title="ELSE IF risk &gt; 8.0" subtitle="escalation">
                ACTION: hold + require_approval(CISO)
              </PolicyBlock>
            </div>
          </div>
        </div>

        <div style={{display:'flex', flexDirection:'column', gap:14}}>
          <div className="card" style={{padding:0}}>
            <div style={{padding:'12px 16px', borderBottom:'1px solid var(--border-soft)', display:'flex', justifyContent:'space-between'}}>
              <div style={{fontSize:13, fontWeight:500}}>Compiled OPA Rego</div>
              <span className="mono" style={{fontSize:10, color:'var(--accent)'}}>auto-generated</span>
            </div>
            <pre className="mono" style={{margin:0, padding:'14px 16px', fontSize:11.5, color:'var(--fg-2)', lineHeight:1.7, overflowX:'auto'}}>
{`package fabric.dlp.card

import rego.v1

redact contains finding if {
  input.request.kind == "prompt"
  input.identity.dept in {"РКО","Розница","Юр"}
  some finding in input.detect.findings
  finding.cat in {"card_pan","ИНН","СНИЛС"}
  not input.identity.user in
    data.allowlists.compliance_team
}

deny contains "risk_too_high" if {
  input.risk.score > 8.0
}`}
            </pre>
          </div>
          <div className="card" style={{padding:'16px 18px'}}>
            <div style={{fontSize:13, fontWeight:500, marginBottom:10}}>Dry-run · последние 30 дней</div>
            <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:14}}>
              <div>
                <div className="mono" style={{fontSize:10, color:'var(--muted)'}}>WOULD MATCH</div>
                <div style={{fontSize:22, fontWeight:500}}>1,284</div>
                <div className="mono" style={{fontSize:11, color:'var(--accent)'}}>+4.1% базы</div>
              </div>
              <div>
                <div className="mono" style={{fontSize:10, color:'var(--muted)'}}>FP RATE ESTIMATE</div>
                <div style={{fontSize:22, fontWeight:500}}>2.8%</div>
                <div className="mono" style={{fontSize:11, color:'var(--warn)'}}>36 случаев → ревью</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function PolicyBlock({kind, title, subtitle, children}) {
  const colors = {
    when:'var(--info)', and:'var(--fg-2)', then:'var(--accent)', else:'var(--warn)'
  };
  const c = colors[kind] || 'var(--fg-2)';
  return (
    <div style={{background:'var(--bg-elev)', border:`1px solid var(--border)`, borderLeft:`3px solid ${c}`, borderRadius:6, padding:'14px 16px'}}>
      <div style={{display:'flex', justifyContent:'space-between', alignItems:'baseline', marginBottom:8}}>
        <span className="mono" style={{fontSize:11, color:c, letterSpacing:'0.08em'}}>{title}</span>
        <span className="mono" style={{fontSize:10, color:'var(--muted)'}}>{subtitle}</span>
      </div>
      <div style={{fontFamily:'var(--font-mono)', fontSize:12, color:'var(--fg)'}}>{children}</div>
    </div>
  );
}
function Connector() {
  return <div style={{display:'flex', justifyContent:'center'}}><div style={{width:1, height:14, background:'var(--border)'}}/></div>;
}

// === ASSET GRAPH ===
function AssetGraph() {
  const ref = useRefD();
  const [hovered, setHovered] = useStateD(null);
  const nodes = [
    {id:'u1', x:120, y:80, label:'analyst@bank.ru', type:'user', kind:'Human'},
    {id:'u2', x:120, y:160, label:'dev@bank.ru', type:'user', kind:'Human'},
    {id:'u3', x:120, y:240, label:'risk-bot', type:'agent', kind:'Service · NHI'},
    {id:'u4', x:120, y:320, label:'pm@bank.ru', type:'user', kind:'Human'},

    {id:'g1', x:350, y:120, label:'AISec Gateway', type:'gateway', kind:'Edge · Envoy'},
    {id:'g2', x:350, y:280, label:'Policy: pii_strict', type:'policy', kind:'OPA · v 12'},

    {id:'m1', x:580, y:60, label:'GigaChat-Pro', type:'model', kind:'Сбер · prod'},
    {id:'m2', x:580, y:140, label:'YandexGPT-5', type:'model', kind:'Я.Облако · prod'},
    {id:'m3', x:580, y:220, label:'Claude-3.5', type:'model', kind:'external · risky'},
    {id:'m4', x:580, y:300, label:'risk-model-v3', type:'model', kind:'self-hosted · vLLM'},
    {id:'m5', x:580, y:380, label:'mcp-jira', type:'mcp', kind:'MCP · unverified'},
  ];
  const edges = [
    {a:'u1', b:'g1'}, {a:'u2', b:'g1'}, {a:'u3', b:'g1'}, {a:'u4', b:'g1'},
    {a:'g1', b:'g2'},
    {a:'g2', b:'m1'}, {a:'g2', b:'m2'}, {a:'g2', b:'m3', risk:true}, {a:'g2', b:'m4'}, {a:'g2', b:'m5', risk:true},
  ];
  const typeColor = {user:'var(--info)', agent:'var(--warn)', gateway:'var(--accent)', policy:'var(--accent)', model:'var(--fg-2)', mcp:'var(--danger)'};

  return (
    <div>
      <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:18}}>
        <div>
          <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:4}}>AI Asset Graph</h3>
          <div className="mono" style={{fontSize:11, color:'var(--muted)'}}>Apache AGE on PG 18.4 · 4 identities · 5 models · 1 MCP server · 2 risks</div>
        </div>
        <div style={{display:'flex', gap:14, fontSize:11.5}}>
          {Object.entries(typeColor).map(([k,c]) =>
            <LegendDot key={k} c={c} l={k}/>
          )}
        </div>
      </div>
      <div className="card" style={{padding:0, position:'relative', overflow:'hidden'}}>
        <div className="dotgrid" style={{position:'absolute', inset:0, opacity:0.5}}/>
        <svg viewBox="0 0 720 440" width="100%" height="440" style={{position:'relative'}}>
          {edges.map((e,i) => {
            const a = nodes.find(n=>n.id===e.a), b = nodes.find(n=>n.id===e.b);
            return <line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y}
              stroke={e.risk ? 'var(--danger)' : 'var(--border)'}
              strokeWidth={hovered && (hovered===e.a || hovered===e.b) ? 1.5 : 1}
              strokeDasharray={e.risk ? '4 3' : 'none'}
              opacity={hovered && hovered !== e.a && hovered !== e.b ? 0.2 : 0.8}/>;
          })}
          {nodes.map(n => (
            <g key={n.id} onMouseEnter={() => setHovered(n.id)} onMouseLeave={() => setHovered(null)} style={{cursor:'pointer'}}>
              <circle cx={n.x} cy={n.y} r={hovered===n.id ? 9 : 6}
                fill="var(--bg-elev)" stroke={typeColor[n.type]} strokeWidth={hovered===n.id ? 2 : 1.4}/>
              <text x={n.x + 14} y={n.y - 4} fill="var(--fg)" fontSize="12" fontFamily="var(--font-sans)">{n.label}</text>
              <text x={n.x + 14} y={n.y + 10} fill="var(--muted)" fontSize="10" fontFamily="var(--font-mono)">{n.kind}</text>
            </g>
          ))}
        </svg>
      </div>
    </div>
  );
}

// === AUDIT LOG ===
function AuditLog() {
  const rows = [
    {ts:'2026-05-12 14:32:08.412', sig:'sha256:9f3a…b81c', act:'block', kind:'dlp.secret', user:'analyst@bank.ru', sigOk:true},
    {ts:'2026-05-12 14:31:47.901', sig:'sha256:c1d4…11ee', act:'redact', kind:'dlp.pii.inn', user:'pm@retail.ru', sigOk:true},
    {ts:'2026-05-12 14:31:22.107', sig:'sha256:7be8…aa20', act:'block', kind:'prompt_injection', user:'unknown', sigOk:true},
    {ts:'2026-05-12 14:30:51.555', sig:'sha256:2acf…8801', act:'block', kind:'dlp.secret.ssh', user:'dev@telecom.ru', sigOk:true},
    {ts:'2026-05-12 14:30:11.044', sig:'sha256:88e2…f0c9', act:'redact', kind:'dlp.pii.snils', user:'support@bank.ru', sigOk:true},
    {ts:'2026-05-12 14:29:44.221', sig:'sha256:55b7…01fa', act:'allow', kind:'cache.hit', user:'finance@bank.ru', sigOk:true},
    {ts:'2026-05-12 14:28:30.180', sig:'sha256:fa10…cc44', act:'block', kind:'dlp.secret.jwt', user:'dev@bank.ru', sigOk:true},
    {ts:'2026-05-12 14:27:01.402', sig:'sha256:0921…ab14', act:'policy_change', kind:'rego.compile', user:'ciso@bank.ru', sigOk:true},
  ];
  return (
    <div>
      <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:18}}>
        <div>
          <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:4}}>Audit Log · подписанный WORM</h3>
          <div className="mono" style={{fontSize:11, color:'var(--muted)'}}>PG immutable signed log · HSM КриптоПро · параллельная запись в ClickHouse</div>
        </div>
        <div style={{display:'flex', gap:8}}>
          <button className="btn btn-ghost" style={{fontSize:12, padding:'8px 12px'}}>CEF → MaxPatrol</button>
          <button className="btn btn-ghost" style={{fontSize:12, padding:'8px 12px'}}>Syslog</button>
          <button className="btn btn-primary" style={{fontSize:12, padding:'8px 12px'}}>Evidence pack PDF</button>
        </div>
      </div>

      <div className="card" style={{padding:0}}>
        <table style={{width:'100%', borderCollapse:'collapse', fontSize:12.5}}>
          <thead>
            <tr style={{color:'var(--muted)', textAlign:'left'}}>
              {['Timestamp (UTC)','Signature','Action','Kind','Identity','Verify'].map(h =>
                <th key={h} style={{padding:'11px 16px', fontWeight:400, fontSize:11, letterSpacing:'0.04em', textTransform:'uppercase'}}>{h}</th>
              )}
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} style={{borderTop:'1px solid var(--border-soft)'}}>
                <td className="mono" style={{padding:'10px 16px', color:'var(--muted)'}}>{r.ts}</td>
                <td className="mono" style={{padding:'10px 16px', color:'var(--fg-2)'}}>{r.sig}</td>
                <td style={{padding:'10px 16px'}}>
                  <span className="mono" style={{
                    fontSize:10.5, padding:'3px 7px', borderRadius:3,
                    background: r.act==='block' ? 'oklch(0.72 0.20 25 / 0.15)'
                      : r.act==='redact' ? 'oklch(0.85 0.16 75 / 0.15)'
                      : r.act==='policy_change' ? 'oklch(0.78 0.13 230 / 0.15)'
                      : 'oklch(0.91 0.20 130 / 0.15)',
                    color: r.act==='block' ? 'var(--danger)'
                      : r.act==='redact' ? 'var(--warn)'
                      : r.act==='policy_change' ? 'var(--info)'
                      : 'var(--accent)',
                  }}>{r.act}</span>
                </td>
                <td className="mono" style={{padding:'10px 16px'}}>{r.kind}</td>
                <td style={{padding:'10px 16px'}}>{r.user}</td>
                <td style={{padding:'10px 16px'}}>
                  {r.sigOk && <span style={{color:'var(--accent)', display:'inline-flex', gap:6, alignItems:'center'}}>
                    <span style={{width:6, height:6, borderRadius:'50%', background:'var(--accent)'}}/>
                    <span className="mono" style={{fontSize:11}}>verified</span>
                  </span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

window.LiveDashboard = LiveDashboard;
