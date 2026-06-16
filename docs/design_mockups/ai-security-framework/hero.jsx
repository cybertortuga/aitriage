// Hero section
const { useState: useStateH, useEffect: useEffectH, useRef: useRefH } = React;

function Hero() {
  return (
    <section style={{position:'relative', paddingTop:60, paddingBottom:120, overflow:'hidden'}}>
      {/* Background grid */}
      <div className="grid-bg" style={{
        position:'absolute', inset:0,
        maskImage:'radial-gradient(ellipse at 50% 30%, #000 30%, transparent 80%)',
        WebkitMaskImage:'radial-gradient(ellipse at 50% 30%, #000 30%, transparent 80%)',
        pointerEvents:'none', opacity:0.8,
      }}/>
      {/* Accent glow */}
      <div style={{
        position:'absolute', top:-200, left:'50%', transform:'translateX(-50%)',
        width:900, height:600,
        background:'radial-gradient(ellipse, oklch(0.91 0.20 130 / 0.10), transparent 60%)',
        pointerEvents:'none', filter:'blur(20px)',
      }}/>

      <div className="container" style={{position:'relative'}}>
        {/* Top status strip */}
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:64,
          fontFamily:'var(--font-mono)', fontSize:11, color:'var(--muted)', letterSpacing:'0.05em'}}>
          <div style={{display:'flex', gap:16, alignItems:'center'}}>
            <span className="tag live">CONTROL LAYER · ACTIVE</span>
            <span>v1.0.4 · build 20260512</span>
          </div>
          <div style={{display:'flex', gap:16}}>
            <span>LAT 55.7558</span><span>LON 37.6173</span>
            <span style={{color:'var(--accent)'}}>● 99.98% SLA</span>
          </div>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'1.3fr 1fr', gap:80, alignItems:'end'}}>
          <div>
            <Eyebrow>AI Security Control Layer · Enterprise · СНГ-first</Eyebrow>
            <h1 className="display" style={{marginTop:24, marginBottom:32}}>
              <span className="text-gradient">Один слой контроля</span><br/>
              <span>между вашими людьми</span><br/>
              <span style={{color:'var(--muted)'}}>и любыми </span>
              <span style={{color:'var(--accent)', fontStyle:'italic', fontFamily:'serif', fontWeight:400}}>AI-системами.</span>
            </h1>
            <p style={{fontSize:18, lineHeight:1.55, color:'var(--fg-2)', maxWidth:560, margin:'0 0 40px'}}>
              AISec Fabric перехватывает каждый запрос к GPT, Claude, GigaChat и YandexGPT.
              Классифицирует данные, применяет политики, редактирует PII в потоке —
              и оставляет подписанный аудит-след для ФСТЭК, 152-ФЗ и КИИ.
            </p>
            <div style={{display:'flex', gap:12}}>
              <a href="#contact" className="btn btn-primary">Запросить пилот →</a>
              <a href="#arch" className="btn btn-ghost">Архитектура · 3 плоскости</a>
            </div>

            {/* Provider list */}
            <div style={{marginTop:64}}>
              <div className="mono" style={{fontSize:11, color:'var(--muted)', letterSpacing:'0.18em', textTransform:'uppercase', marginBottom:16}}>
                Поддерживаем из коробки
              </div>
              <div style={{display:'flex', flexWrap:'wrap', gap:8}}>
                {['GigaChat','YandexGPT','MTS AI','T-Lite','OpenAI','Anthropic','Gemini','vLLM','Ollama','TGI'].map(p =>
                  <span key={p} className="tag" style={{padding:'6px 10px', fontSize:12}}>{p}</span>
                )}
              </div>
            </div>
          </div>

          {/* Right side: live request console */}
          <RequestConsole/>
        </div>

        {/* Bottom strip — key metrics */}
        <div style={{marginTop:96, paddingTop:32, borderTop:'1px solid var(--border-soft)',
          display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:32}}>
          {[
            {n:'< 5ms', l:'fast-path detection latency'},
            {n:'30–50', l:'готовых правил под РФ из коробки'},
            {n:'60–80%', l:'трафика снимает embedding-кэш'},
            {n:'1–2 дня', l:'установка через Helm или single binary'},
          ].map((s,i) => (
            <div key={i}>
              <div className="stat-num">{s.n}</div>
              <div style={{fontSize:13, color:'var(--muted)', marginTop:8, maxWidth:200}}>{s.l}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// Live-feeling console mock
function RequestConsole() {
  const [tick, setTick] = useStateH(0);
  useEffectH(() => {
    const id = setInterval(() => setTick(t => t + 1), 2400);
    return () => clearInterval(id);
  }, []);

  const scenarios = [
    {
      user: 'analyst@bank.ru', model: 'GigaChat-Pro',
      prompt: 'Сравни выписки клиента 7708123456 за май и июнь, выдели аномалии',
      detect: [
        {k:'ИНН', v:'7708123456', a:'redact'},
        {k:'PII:financial_records', v:'2 matches', a:'allow + log'},
      ],
      verdict:'ALLOW · REDACTED', color:'var(--accent)'
    },
    {
      user: 'dev@telecom.ru', model: 'Claude-3.5',
      prompt: 'Закинь сюда наш AWS prod key чтобы я проверил формат AKIA...',
      detect: [
        {k:'AWS_ACCESS_KEY', v:'AKIA****PROD', a:'block'},
        {k:'risk_score', v:'9.4 / 10', a:'CISO alert'},
      ],
      verdict:'BLOCKED · SECRETS', color:'var(--danger)'
    },
    {
      user: 'pm@fintech.ru', model: 'YandexGPT-5',
      prompt: 'Дай мне сводку по продукту по этому брифу [...]',
      detect: [
        {k:'cache_hit', v:'cos=0.97', a:'fast path'},
        {k:'classification', v:'public', a:'allow'},
      ],
      verdict:'ALLOW · CACHED', color:'var(--info)'
    },
  ];
  const s = scenarios[tick % scenarios.length];

  return (
    <CornerFrame style={{padding:0, background:'var(--bg-elev)', border:'1px solid var(--border)', borderRadius:10}}>
      {/* Header */}
      <div style={{display:'flex', alignItems:'center', justifyContent:'space-between',
        padding:'14px 18px', borderBottom:'1px solid var(--border-soft)'}}>
        <div style={{display:'flex', alignItems:'center', gap:8, fontFamily:'var(--font-mono)', fontSize:11}}>
          <span style={{color:'var(--muted)'}}>FABRIC ›</span>
          <span style={{color:'var(--fg)'}}>data_plane</span>
          <span style={{color:'var(--muted)'}}>›</span>
          <span style={{color:'var(--accent)'}}>ext_proc.go</span>
        </div>
        <div style={{display:'flex', gap:6}}>
          {['var(--danger)','var(--warn)','var(--accent)'].map((c,i) =>
            <span key={i} style={{width:8, height:8, borderRadius:'50%', background:c}}/>)}
        </div>
      </div>

      {/* Body */}
      <div key={tick} className="fadeup" style={{padding:'18px 20px', fontFamily:'var(--font-mono)', fontSize:12, lineHeight:1.75}}>
        <div style={{color:'var(--muted)'}}>
          <span style={{color:'var(--accent)'}}>$</span> tail -f /var/log/fabric/proxy.jsonl | jq
        </div>
        <div style={{marginTop:10, color:'var(--muted)'}}>{'{'}</div>
        <Row k="ts" v={`"2026-05-${String(12+(tick%6)).padStart(2,'0')}T14:${22+(tick%30)}:00Z"`}/>
        <Row k="identity" v={`"${s.user}"`}/>
        <Row k="route" v={`"openai/${s.model}"`}/>
        <Row k="prompt" v={`"${s.prompt}"`} wrap/>
        <Row k="detect" raw>
          <div style={{paddingLeft:12, marginTop:2}}>
            {s.detect.map((d,i) => (
              <div key={i} style={{display:'flex', gap:10}}>
                <span style={{color:'var(--muted)'}}>{d.k.padEnd(20,'.')}</span>
                <span style={{color:'var(--fg)'}}>{d.v}</span>
                <span style={{color:'var(--accent)', marginLeft:'auto'}}>→ {d.a}</span>
              </div>
            ))}
          </div>
        </Row>
        <Row k="latency" v="3.4ms (rules) + 41ms (ml async)"/>
        <div style={{color:'var(--muted)'}}>{'}'}</div>
        <div style={{marginTop:14, padding:'10px 12px', borderRadius:6,
          background:'oklch(0.91 0.20 130 / 0.08)',
          border:`1px solid ${s.color === 'var(--danger)' ? 'oklch(0.72 0.20 25 / 0.4)' : 'var(--accent-border)'}`,
          color: s.color, fontWeight:500}}>
          ► verdict: {s.verdict}
        </div>
      </div>
    </CornerFrame>
  );
}

function Row({k, v, wrap, raw, children}) {
  return (
    <div style={{display:'flex', gap:8, alignItems:'flex-start'}}>
      <span style={{color:'var(--muted)', minWidth:96}}>  "{k}":</span>
      {raw ? children : (
        <span style={{color: k==='prompt' ? 'var(--fg-2)' : 'var(--fg)', wordBreak: wrap ? 'break-word' : 'normal'}}>{v},</span>
      )}
    </div>
  );
}

Object.assign(window, { Hero });
