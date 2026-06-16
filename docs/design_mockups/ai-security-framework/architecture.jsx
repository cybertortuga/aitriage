// Architecture — 3 planes + Edge
function Architecture() {
  const [active, setActive] = React.useState(1);

  const planes = [
    {
      id: 0,
      name: 'Edge Layer',
      tag: 'PLANE 00',
      desc: 'Точки входа в Fabric. Лёгкие интеграции, никакой бизнес-логики.',
      items: [
        {n:'Python SDK', d:'httpx + pydantic, drop-in для OpenAI/Anthropic клиентов'},
        {n:'Go SDK', d:'нативный, нулевые аллокации на критическом пути'},
        {n:'Secure AI Portal', d:'минимальный chat-UI на Next.js для не-разработчиков'},
        {n:'Direct API', d:'OpenAI- и Anthropic-совместимые endpoint`ы'},
      ],
      stack: ['httpx','pydantic','Next.js 15','OpenAPI 3.1'],
    },
    {
      id: 1,
      name: 'Data Plane',
      tag: 'PLANE 01',
      desc: 'Envoy + Go ext_proc. Здесь проходит каждый байт каждого запроса. Stateless, под 5ms на fast-path.',
      items: [
        {n:'Envoy Proxy 1.39+', d:'L7 ingress, TLS/mTLS, RLS, нативный SSE-handling через sse_to_metadata'},
        {n:'Go ext_proc', d:'парсинг SSE, sliding-window inspection 200–300ms'},
        {n:'Aho-Corasick scan', d:'multi-pattern detection PII и секретов без аллокаций (sync.Pool)'},
        {n:'Reservation budget', d:'резерв токенов до вызова, settle после стрима'},
      ],
      stack: ['Envoy 1.39','Go 1.24','Proxy-Wasm','Valkey 8'],
    },
    {
      id: 2,
      name: 'Intelligence Plane',
      tag: 'PLANE 02',
      desc: 'ML-классификация и enrichment вне critical path. Aggressive timeout 50ms, fallback на rules-only.',
      items: [
        {n:'Prompt injection clf', d:'DeBERTa / RoBERTa через ONNX Runtime 1.20+'},
        {n:'PII classifier', d:'ru-NER модели + Microsoft Presidio + кастомные правила РФ'},
        {n:'Secret scanner', d:'Aho-Corasick + ML для контекстной классификации'},
        {n:'Embedding cache', d:'Qdrant 1.x, cosine > 0.95 → пропуск ML, снимает 60–80%'},
      ],
      stack: ['FastAPI','ONNX 1.20+','Qdrant 1.x','NATS JetStream'],
    },
    {
      id: 3,
      name: 'Control Plane',
      tag: 'PLANE 03',
      desc: 'Go modular monolith. Один бинарь, разделённый на модули. Дробим на сервисы только когда команда вырастет до 15+.',
      items: [
        {n:'Asset Graph', d:'Apache AGE поверх PostgreSQL 18 — связи людей, моделей, агентов, MCP-серверов'},
        {n:'Shadow AI Engine', d:'свой код, главное конкурентное IP'},
        {n:'Policy Service', d:'OPA Rego + YAML DSL + Visual Builder поверх'},
        {n:'Audit Service', d:'write to ClickHouse + PG immutable signed log'},
      ],
      stack: ['Go 1.24','chi + connect-go','PG 18.4 + AGE','OPA 0.69+'],
    },
  ];

  return (
    <section id="arch" className="section" style={{position:'relative'}}>
      <div className="container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:64, alignItems:'end', marginBottom:64}}>
          <div>
            <Eyebrow num="02 /">Архитектура</Eyebrow>
            <h2 className="h2" style={{marginTop:18}}>
              Три плоскости. Один бинарь Control&nbsp;Plane.<br/>
              <span style={{color:'var(--muted)'}}>Никакого микросервисного зоопарка на 8 человек.</span>
            </h2>
          </div>
          <p style={{fontSize:15, color:'var(--fg-2)', lineHeight:1.6, maxWidth:480, marginLeft:'auto'}}>
            Tiered detection как core principle: дешёвые правила в Go под 5ms,
            ML-enrichment async через NATS с aggressive timeout. Sync-вызовов Python
            из critical path нет нигде.
          </p>
        </div>

        {/* Plane navigator */}
        <div style={{display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:0, marginBottom:0,
          borderTop:'1px solid var(--border)', borderBottom:'1px solid var(--border)'}}>
          {planes.map(p => (
            <button key={p.id} onClick={() => setActive(p.id)} style={{
              background:'transparent', border:'none', textAlign:'left', cursor:'pointer',
              padding:'20px 24px',
              borderRight: p.id < planes.length-1 ? '1px solid var(--border)' : 'none',
              borderTop: active === p.id ? '2px solid var(--accent)' : '2px solid transparent',
              marginTop:-1,
              background: active === p.id ? 'var(--bg-elev)' : 'transparent',
              transition:'background .2s',
            }}>
              <div className="mono" style={{fontSize:11, color: active === p.id ? 'var(--accent)' : 'var(--muted)',
                letterSpacing:'0.1em', marginBottom:8}}>{p.tag}</div>
              <div style={{fontSize:18, fontWeight:500, letterSpacing:'-0.015em',
                color: active === p.id ? 'var(--fg)' : 'var(--fg-2)'}}>{p.name}</div>
            </button>
          ))}
        </div>

        {/* Active panel */}
        <div key={active} className="fadeup" style={{
          background:'var(--bg-elev)', borderLeft:'1px solid var(--border)', borderRight:'1px solid var(--border)',
          borderBottom:'1px solid var(--border)',
          display:'grid', gridTemplateColumns:'1fr 1.2fr', gap:0,
        }}>
          {/* Left description */}
          <div style={{padding:'36px 32px', borderRight:'1px solid var(--border)'}}>
            <div className="mono" style={{fontSize:11, color:'var(--muted)', marginBottom:12}}>
              › {planes[active].tag}
            </div>
            <h3 className="h3" style={{marginBottom:14, fontSize:28}}>{planes[active].name}</h3>
            <p style={{fontSize:14, color:'var(--fg-2)', lineHeight:1.65, marginBottom:24}}>
              {planes[active].desc}
            </p>
            <div className="mono" style={{fontSize:11, color:'var(--muted)', marginBottom:10, letterSpacing:'0.08em'}}>
              STACK
            </div>
            <div style={{display:'flex', flexWrap:'wrap', gap:6}}>
              {planes[active].stack.map(s => <span key={s} className="tag mono">{s}</span>)}
            </div>
          </div>

          {/* Right items grid */}
          <div style={{padding:'36px 32px', display:'grid', gridTemplateColumns:'1fr 1fr', gap:24}}>
            {planes[active].items.map((it, i) => (
              <div key={it.n} style={{paddingTop:18, borderTop:'1px solid var(--border-soft)'}}>
                <div className="mono" style={{fontSize:10, color:'var(--accent)', marginBottom:8}}>
                  0{i+1}
                </div>
                <div style={{fontSize:15, fontWeight:500, marginBottom:6}}>{it.n}</div>
                <div style={{fontSize:13, color:'var(--muted)', lineHeight:1.55}}>{it.d}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Flow diagram */}
        <FlowDiagram active={active} setActive={setActive}/>
      </div>
    </section>
  );
}

function FlowDiagram({ active, setActive }) {
  return (
    <div style={{marginTop:48, padding:'40px 24px', border:'1px solid var(--border-soft)', borderRadius:10}}>
      <div className="mono" style={{fontSize:11, color:'var(--muted)', letterSpacing:'0.1em', marginBottom:24}}>
        DATA FLOW · ONE REQUEST
      </div>
      <div style={{display:'grid', gridTemplateColumns:'auto 1fr auto 1fr auto 1fr auto 1fr auto', alignItems:'center', gap:0}}>
        {[
          {label:'USER / APP', sub:'SDK / Portal', id:0},
          null,
          {label:'ENVOY', sub:'TLS / mTLS / RLS', id:1, mark:'01'},
          null,
          {label:'EXT_PROC', sub:'AC + reserve', id:1, mark:'01'},
          null,
          {label:'POLICY', sub:'OPA Rego', id:3, mark:'03'},
          null,
          {label:'MODEL', sub:'GigaChat / GPT', id:null},
        ].map((node, i) => {
          if (node === null) {
            return <Arrow key={i} idx={i}/>;
          }
          const isActive = active === node.id;
          return (
            <div key={i} onClick={() => node.id !== null && setActive(node.id)}
              style={{
                cursor: node.id !== null ? 'pointer' : 'default',
                padding:'14px 18px',
                border:`1px solid ${isActive ? 'var(--accent)' : 'var(--border)'}`,
                borderRadius:6,
                background: isActive ? 'oklch(0.91 0.20 130 / 0.06)' : 'var(--bg)',
                textAlign:'center',
                transition:'all .2s',
              }}>
              {node.mark && <div className="mono" style={{fontSize:10, color: isActive ? 'var(--accent)' : 'var(--muted)', marginBottom:4}}>{node.mark}</div>}
              <div style={{fontSize:12, fontWeight:600, letterSpacing:'-0.005em'}}>{node.label}</div>
              <div style={{fontSize:11, color:'var(--muted)', marginTop:2, fontFamily:'var(--font-mono)'}}>{node.sub}</div>
            </div>
          );
        })}
      </div>

      {/* Side branches */}
      <div style={{marginTop:24, display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:16}}>
        <SideBranch from="ext_proc" to="Intelligence Plane" detail="async via NATS · t/o 50ms" plane={2} active={active} setActive={setActive}/>
        <SideBranch from="ext_proc" to="Embedding cache" detail="Qdrant · cos > 0.95" plane={2} active={active} setActive={setActive}/>
        <SideBranch from="all" to="Audit log" detail="CH + PG signed WORM" plane={3} active={active} setActive={setActive}/>
      </div>
    </div>
  );
}

function Arrow({ idx }) {
  return (
    <div style={{display:'flex', alignItems:'center', gap:4, padding:'0 8px'}}>
      <div style={{flex:1, height:1, background:'var(--border)', borderTop:'1px dashed var(--accent-border)', borderBottom:'none', border:0, borderTop:'1px dashed var(--border)'}}>
        <div style={{height:1, background:'var(--border)'}}/>
      </div>
      <span style={{color:'var(--accent)', fontFamily:'var(--font-mono)', fontSize:14}}>›</span>
    </div>
  );
}

function SideBranch({from, to, detail, plane, active, setActive}) {
  const isActive = active === plane;
  return (
    <div onClick={() => setActive(plane)} style={{
      cursor:'pointer',
      padding:'14px 16px',
      background: isActive ? 'oklch(0.91 0.20 130 / 0.06)' : 'transparent',
      border:`1px dashed ${isActive ? 'var(--accent)' : 'var(--border-soft)'}`,
      borderRadius:6,
      display:'flex', justifyContent:'space-between', alignItems:'center', gap:12
    }}>
      <div>
        <div className="mono" style={{fontSize:11, color:'var(--muted)'}}>{from} ↗ {to}</div>
        <div style={{fontSize:12, marginTop:4, color:'var(--fg-2)'}}>{detail}</div>
      </div>
    </div>
  );
}

window.Architecture = Architecture;
