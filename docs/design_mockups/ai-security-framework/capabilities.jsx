// Capabilities + Discovery section
function Capabilities() {
  const caps = [
    {
      n:'01', t:'AI Gateway',
      d:'Envoy + Go ext_proc. TLS/mTLS, rate limiting, circuit breaker. Нативный SSE с sliding-window inspection. Reservation-based token budget.',
      stat:'<5ms', statLabel:'fast-path latency',
      tags:['Envoy 1.39+','Go ext_proc','SSE-aware','Reservation budget']
    },
    {
      n:'02', t:'DLP-движок',
      d:'Aho-Corasick multi-pattern scan. 30–50 готовых правил под РФ: ИНН, СНИЛС, паспорт, ОГРН, карты по Луна, мед-данные. Customer-extensible словари.',
      stat:'30–50', statLabel:'правил RU из коробки',
      tags:['Aho-Corasick','Presidio + ru-NER','Smart redaction','JSON-safe']
    },
    {
      n:'03', t:'Shadow AI Discovery',
      d:'DNS-логи, корпоративный прокси, SSO, cloud billing, Git-сканер на hardcoded LLM keys. Non-invasive в Phase 1. eBPF Tetragon в Phase 2.',
      stat:'5+', statLabel:'источников телеметрии',
      tags:['DNS / Proxy','SSO logs','Cloud billing','Git scanner']
    },
    {
      n:'04', t:'Visual Policy Builder',
      d:'CISO и compliance-офицеры не пишут Rego. Блочный конструктор поверх OPA Rego генерирует код, который dev`ы могут править руками.',
      stat:'20–30', statLabel:'базовых блоков',
      tags:['OPA Rego','YAML DSL','Dry-run mode','Policy simulator']
    },
    {
      n:'05', t:'AI Asset Graph',
      d:'Apache AGE на PostgreSQL 18.4. Люди, модели, агенты, MCP-серверы, RAG-индексы, API-ключи — всё first-class сущности с связями.',
      stat:'AGE / PG18', statLabel:'graph поверх реляционки',
      tags:['Apache AGE','Cypher-like','Non-human identity','SCIM 2.0']
    },
    {
      n:'06', t:'Compliance Evidence',
      d:'Immutable signed audit log в PG параллельно с ClickHouse. WORM-like хранилище для критичных evidence. HSM-подпись (Thales, КриптоПро).',
      stat:'WORM', statLabel:'tamper-evident storage',
      tags:['152-ФЗ','187-ФЗ КИИ','ФСТЭК-prep','CEF / Syslog']
    },
  ];

  return (
    <section className="section" style={{position:'relative', background:'oklch(0.155 0.012 245)'}}>
      <div className="container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr', gap:48, alignItems:'end', marginBottom:64}}>
          <div>
            <Eyebrow num="03 /">Возможности</Eyebrow>
            <h2 className="h2" style={{marginTop:18}}>
              Шесть модулей, которые<br/>работают в день&nbsp;один.
            </h2>
          </div>
          <p style={{fontSize:15, color:'var(--fg-2)', lineHeight:1.6, maxWidth:480, marginLeft:'auto'}}>
            MVP заморожен. Ничего лишнего. Все десять пунктов из чек-листа «что нужно
            к первому платному пилоту» — внутри. Phase 2 (MCP Registry, AIBOM, RAG security,
            Tetragon) добавляется поверх не ломая интеграции.
          </p>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:0,
          border:'1px solid var(--border-soft)', borderRadius:10, overflow:'hidden'}}>
          {caps.map((c, i) => {
            const isRight = i % 3 === 2;
            const isBottom = i >= 3;
            return (
              <div key={c.n} className="hover-lift" style={{
                padding:'32px 28px',
                borderRight: !isRight ? '1px solid var(--border-soft)' : 'none',
                borderTop: isBottom ? '1px solid var(--border-soft)' : 'none',
                background:'var(--bg-elev)',
                position:'relative', minHeight:280,
                display:'flex', flexDirection:'column',
              }}>
                <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start', marginBottom:24}}>
                  <span className="mono" style={{fontSize:11, color:'var(--accent)', letterSpacing:'0.1em'}}>
                    {c.n} / 06
                  </span>
                  <div style={{textAlign:'right'}}>
                    <div className="mono" style={{fontSize:22, color:'var(--fg)', letterSpacing:'-0.02em'}}>{c.stat}</div>
                    <div style={{fontSize:11, color:'var(--muted)', marginTop:2}}>{c.statLabel}</div>
                  </div>
                </div>
                <h3 style={{fontSize:22, fontWeight:500, letterSpacing:'-0.02em', marginBottom:12}}>{c.t}</h3>
                <p style={{fontSize:13.5, color:'var(--fg-2)', lineHeight:1.6, marginBottom:24, flex:1}}>{c.d}</p>
                <div style={{display:'flex', flexWrap:'wrap', gap:6}}>
                  {c.tags.map(t => <span key={t} className="tag mono" style={{fontSize:10}}>{t}</span>)}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

window.Capabilities = Capabilities;
