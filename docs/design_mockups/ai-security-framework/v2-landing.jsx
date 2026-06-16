// v2 Landing — single-page master
const { useState: lS, useEffect: lE, useRef: lR } = React;

const SLIDES = [
  {
    eyebrow: 'AISec Fabric',
    title: 'Защищать ИИ —\nкак инфраструктуру',
    desc: 'Один слой контроля между вашими сотрудниками и любыми AI-системами. Перехватываем каждый запрос к GPT, Claude, GigaChat и YandexGPT — классифицируем, редактируем PII, оставляем подписанный аудит-след.',
    primary: 'Запросить КП',
    secondary: 'Платформа',
    primaryHref: '#contact',
    secondaryHref: '#products',
  },
  {
    eyebrow: 'Shadow AI Discovery',
    title: 'Видеть, что уже\nиспользуется',
    desc: 'DNS, корпоративный прокси, SSO, cloud-биллинг, сканер Git на hardcoded LLM-ключи. Несколько источников телеметрии без агентов на машинах. Карта теневого ИИ за 48 часов после развёртывания.',
    primary: 'Запросить КП',
    secondary: 'Как работает',
    primaryHref: '#contact',
    secondaryHref: '#how',
  },
  {
    eyebrow: 'Compliance Evidence',
    title: 'Регулятор как\nзащитный ров',
    desc: 'Подписанный иммутабельный аудит-лог, evidence-паки в один клик, CEF-экспорт в SIEM. 152-ФЗ, 187-ФЗ КИИ, ФСТЭК, ГОСТ — карта соответствия и evidence в формате, удобном проверяющему.',
    primary: 'Запросить КП',
    secondary: 'Compliance',
    primaryHref: '#contact',
    secondaryHref: '#compliance',
  },
  {
    eyebrow: 'AI Red Team · сервис',
    title: 'Атаковать раньше,\nчем атакуют вас',
    desc: 'Регулярный аудит ваших LLM-приложений: prompt injection, data exfiltration, jailbreaks, RAG-poisoning. Команда сертифицированных пентестеров. Отчёт под формат проверок ЦБ РФ и ФСТЭК.',
    primary: 'Заказать аудит',
    secondary: 'Сервисы',
    primaryHref: '#contact',
    secondaryHref: '#services',
  },
];

function V2HeroSlider() {
  const [idx, setIdx] = lS(0);
  const [paused, setPaused] = lS(false);
  const [progress, setProgress] = lS(0);

  lE(() => {
    if (paused) return;
    const dur = 7000;
    const start = Date.now();
    const tick = () => {
      const elapsed = Date.now() - start;
      const p = Math.min(elapsed / dur, 1);
      setProgress(p * 100);
      if (p >= 1) {
        setIdx(i => (i + 1) % SLIDES.length);
      }
    };
    const id = setInterval(tick, 200);  // 5 fps progress is plenty
    return () => clearInterval(id);
  }, [idx, paused]);

  lE(() => { setProgress(0); }, [idx]);

  const next = () => setIdx(i => (i + 1) % SLIDES.length);
  const prev = () => setIdx(i => (i - 1 + SLIDES.length) % SLIDES.length);

  const slide = SLIDES[idx];

  return (
    <section
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      style={{position:'relative', minHeight:'calc(100vh - 76px)', overflow:'hidden', background:'#000'}}>
      <div style={{position:'absolute', inset:0, opacity:0.95}}>
        <V2RibbonBg variant={idx}/>
      </div>
      <div style={{position:'absolute', inset:0,
        background:'linear-gradient(90deg, #000 0%, #000 28%, rgba(0,0,0,0.7) 50%, rgba(0,0,0,0.2) 80%, transparent 100%)',
        pointerEvents:'none'}}/>
      <div style={{position:'absolute', bottom:0, left:0, right:0, height:200,
        background:'linear-gradient(to top, #000, transparent)', pointerEvents:'none'}}/>

      <V2AlertCorner/>

      <div className="v2-container" style={{position:'relative', height:'calc(100vh - 76px)',
        minHeight:680, display:'flex', flexDirection:'column'}}>
        <div style={{paddingTop:64, display:'flex', alignItems:'center', gap:20}}>
          <button className="v2-slider-arrow" onClick={prev} aria-label="Назад">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M15 18l-6-6 6-6"/></svg>
          </button>
          <button className="v2-slider-arrow" onClick={next} aria-label="Вперёд">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 18l6-6-6-6"/></svg>
          </button>
          <div style={{display:'flex', gap:10, flex:1, maxWidth:560, marginLeft:20}}>
            {SLIDES.map((_, i) => (
              <button key={i} onClick={() => setIdx(i)} style={{
                flex:1, height:24, background:'transparent', border:'none', cursor:'pointer', padding:'10px 0'
              }}>
                <div className="v2-slider-bar">
                  <div className="v2-slider-bar-fill" style={{
                    width: i < idx ? '100%' : i === idx ? `${progress}%` : '0%'
                  }}/>
                </div>
              </button>
            ))}
          </div>
          <div className="v2-mono" style={{fontSize:13, color:'var(--v2-muted)', letterSpacing:'0.05em'}}>
            <span style={{color:'var(--v2-fg)'}}>{String(idx+1).padStart(2,'0')}</span>
            <span> / {String(SLIDES.length).padStart(2,'0')}</span>
          </div>
        </div>

        <div style={{flex:1, display:'flex', alignItems:'center', paddingBottom:80}}>
          <div key={idx} className="v2-fade-in" style={{maxWidth:780}}>
            <div className="v2-eyebrow-red" style={{marginBottom:32}}>{slide.eyebrow}</div>
            <h1 className="v2-display" style={{marginBottom:32, whiteSpace:'pre-line'}}>{slide.title}</h1>
            <p className="v2-lead" style={{maxWidth:620, marginBottom:48}}>{slide.desc}</p>
            <div style={{display:'flex', gap:12, flexWrap:'wrap'}}>
              <a href={slide.primaryHref} className="v2-btn v2-btn-red">{slide.primary}</a>
              <a href={slide.secondaryHref} className="v2-btn v2-btn-ghost">{slide.secondary}</a>
            </div>
          </div>
        </div>

        <div style={{position:'absolute', bottom:32, left:40, right:40,
          display:'flex', justifyContent:'space-between', alignItems:'center', flexWrap:'wrap', gap:16}}>
          <div className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)', letterSpacing:'0.05em'}}>
            AI SECURITY · ENTERPRISE · СНГ-FIRST
          </div>
          <a href="#stats" style={{color:'var(--v2-fg-2)', textDecoration:'none', fontSize:13,
            display:'inline-flex', alignItems:'center', gap:10}}>
            <span>Прокрутить вниз</span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14M5 12l7 7 7-7"/>
            </svg>
          </a>
        </div>
      </div>
    </section>
  );
}

function V2Stats() {
  const stats = [
    {n:'< 5 мс', l:'детекция в потоке без потери UX'},
    {n:'48 ч', l:'до полной карты теневого ИИ'},
    {n:'10+', l:'провайдеров LLM из коробки'},
    {n:'1–2 дня', l:'установка через Helm или single binary'},
  ];
  return (
    <section id="stats" className="v2-section" style={{paddingTop:160, paddingBottom:120}}>
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 1fr 1fr 1fr', gap:0,
          borderTop:'1px solid var(--v2-border-soft)', borderBottom:'1px solid var(--v2-border-soft)'}}>
          {stats.map((s, i) => (
            <div key={i} style={{
              padding:'48px 40px',
              borderRight: i < stats.length-1 ? '1px solid var(--v2-border-soft)' : 'none',
            }}>
              <div style={{fontSize:'clamp(48px, 5vw, 72px)', fontWeight:600,
                letterSpacing:'-0.04em', lineHeight:1, marginBottom:20}}>{s.n}</div>
              <div style={{fontSize:15, color:'var(--v2-muted)', maxWidth:200, lineHeight:1.45}}>{s.l}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function V2Services() {
  const items = [
    {t:'Пилот за 2 месяца', d:'Установка, интеграция, обучение CISO. Отчёт с практическими находками и оценкой эффекта.'},
    {t:'AI Red Team', d:'Пентест ваших LLM-приложений. Prompt injection, RAG-poisoning, exfiltration. Регулярные кампании.'},
    {t:'Threat Hunting', d:'24/7 мониторинг AI-инцидентов. Подключение к вашему SIEM. Реагирование на критичные срабатывания.'},
    {t:'Обучение CISO', d:'Курс по AI Security. Сертификация специалистов. AI-risk framework для совета директоров.'},
  ];
  return (
    <section id="services" className="v2-section" style={{background:'#08080a', borderTop:'1px solid var(--v2-border-soft)', borderBottom:'1px solid var(--v2-border-soft)'}}>
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 2fr', gap:64, alignItems:'start', marginBottom:48}}>
          <div>
            <div className="v2-eyebrow-red" style={{marginBottom:24}}>Сервисы</div>
            <h2 className="v2-h2">
              Продукт + люди.<br/>
              <span style={{color:'var(--v2-muted)'}}>Не только лицензия.</span>
            </h2>
          </div>
          <p className="v2-lead" style={{maxWidth:600, marginTop:24}}>
            Платформа без команды защитников — игрушка. Мы поставляем продукт вместе
            с сервисами: внедрение, регулярный red team, threat hunting,
            обучение CISO. Один поставщик, один контракт, один ответственный.
          </p>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:0,
          borderTop:'1px solid var(--v2-border-soft)'}}>
          {items.map((s, i) => (
            <a key={s.t} href="#contact" style={{
              padding:'40px 32px',
              borderRight: i < items.length-1 ? '1px solid var(--v2-border-soft)' : 'none',
              borderBottom:'1px solid var(--v2-border-soft)',
              textDecoration:'none', color:'inherit',
              display:'flex', flexDirection:'column', justifyContent:'space-between',
              minHeight:240, transition:'background .15s',
            }}
            onMouseOver={e=>e.currentTarget.style.background='#0c0c0e'}
            onMouseOut={e=>e.currentTarget.style.background='transparent'}>
              <div>
                <h3 style={{fontSize:22, fontWeight:600, letterSpacing:'-0.015em', marginBottom:16}}>{s.t}</h3>
                <p style={{fontSize:14, color:'var(--v2-muted)', lineHeight:1.55, margin:0}}>{s.d}</p>
              </div>
              <span className="v2-link v2-link-arrow" style={{fontSize:14, marginTop:24}}>Запросить КП</span>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}

function V2Research() {
  const items = [
    {tag:'whitepaper', title:'Десять пунктов чек-листа: что должно быть в первом AI Security пилоте',
      meta:'24 страницы · PDF · апрель 2026'},
    {tag:'отчёт', title:'Shadow AI в крупных банках РФ. Полевые данные за 6 месяцев',
      meta:'18 страниц · PDF · март 2026'},
    {tag:'кейс', title:'Как T2 закрыл утечку клиентских данных через ChatGPT за 11 дней',
      meta:'8 страниц · PDF · февраль 2026'},
  ];
  return (
    <section className="v2-section">
      <div className="v2-container">
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-end',
          marginBottom:48, flexWrap:'wrap', gap:32}}>
          <div>
            <div className="v2-eyebrow-red" style={{marginBottom:24}}>Исследования</div>
            <h2 className="v2-h2">Что мы знаем о&nbsp;вашем рынке.</h2>
          </div>
          <a href="#" className="v2-btn v2-btn-ghost">Все материалы →</a>
        </div>

        <div style={{display:'grid', gridTemplateColumns:'1.4fr 1fr 1fr', gap:16}}>
          {items.map((it, i) => (
            <a key={i} href="#" className="v2-card" style={{
              textDecoration:'none', color:'inherit',
              minHeight: i === 0 ? 420 : 320,
              display:'flex', flexDirection:'column', justifyContent:'space-between',
              position:'relative', overflow:'hidden',
              padding: i === 0 ? '36px' : '28px',
            }}>
              {i === 0 && (
                <div style={{position:'absolute', top:0, right:0, width:'60%', height:'100%', pointerEvents:'none', opacity:0.5}}>
                  <div style={{position:'absolute', inset:0, background:'radial-gradient(circle at 70% 30%, rgba(255,13,44,0.35), transparent 60%)'}}/>
                </div>
              )}
              <div style={{position:'relative'}}>
                <span className="v2-tag v2-tag-red" style={{marginBottom:24}}>{it.tag}</span>
                <h3 style={{
                  fontSize: i === 0 ? 30 : 21,
                  fontWeight:600, letterSpacing:'-0.018em', lineHeight:1.2,
                  marginTop:24, maxWidth: i === 0 ? '70%' : '100%'
                }}>{it.title}</h3>
              </div>
              <div style={{position:'relative', display:'flex', justifyContent:'space-between',
                alignItems:'center', marginTop:48}}>
                <span className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)'}}>{it.meta}</span>
                <span className="v2-link v2-link-arrow">Скачать</span>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}

function V2Customers() {
  const sectors = [
    'банки топ-30','телеком-операторы','страхование','ритейл','финтех',
    'промышленность','госкорпорации','медицина','маркетплейсы','логистика'
  ];
  return (
    <section className="v2-section-sm" style={{borderTop:'1px solid var(--v2-border-soft)', paddingTop:64, paddingBottom:64}}>
      <div className="v2-container">
        <div style={{display:'grid', gridTemplateColumns:'1fr 2.4fr', gap:64, alignItems:'center'}}>
          <div>
            <div className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)', letterSpacing:'0.1em', marginBottom:14}}>
              КЛИЕНТСКАЯ БАЗА
            </div>
            <h3 style={{fontSize:24, fontWeight:600, letterSpacing:'-0.018em', margin:0}}>
              Работаем с крупным enterprise в РФ
            </h3>
          </div>
          <div style={{display:'flex', flexWrap:'wrap', gap:8}}>
            {sectors.map(s => (
              <span key={s} className="v2-tag" style={{padding:'10px 18px', fontSize:14, fontWeight:500}}>{s}</span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function V2CTA() {
  return (
    <section id="contact" className="v2-section" style={{paddingTop:120, paddingBottom:120, position:'relative', overflow:'hidden'}}>
      <div style={{position:'absolute', inset:0,
        background:'radial-gradient(ellipse at 70% 50%, rgba(255,13,44,0.18), transparent 60%)',
        pointerEvents:'none'}}/>
      <div style={{position:'absolute', inset:0, background:'linear-gradient(to bottom, #000 0%, rgba(0,0,0,0.85) 50%, #000 100%)', pointerEvents:'none'}}/>
      <div className="v2-container" style={{position:'relative'}}>
        <div style={{maxWidth:880}}>
          <div className="v2-eyebrow-red" style={{marginBottom:32}}>Связаться</div>
          <h2 className="v2-display" style={{fontSize:'clamp(48px, 6vw, 92px)', marginBottom:32}}>
            Поговорим о вашем<br/>
            <span style={{color:'var(--v2-red)'}}>периметре ИИ.</span>
          </h2>
          <p className="v2-lead" style={{maxWidth:620, marginBottom:48}}>
            Стоимость лицензии и сервисов — по запросу КП. Срок развёртывания
            пилота — 1–2 дня. SSO/AD с первого дня. 10+ провайдеров LLM.
            Готовые правила DLP под российский периметр.
            Отчёт за пилот с практическими находками и оценкой эффекта.
          </p>
          <div style={{display:'flex', gap:12, flexWrap:'wrap'}}>
            <a href="mailto:sales@aisecfabric.ru" className="v2-btn v2-btn-red">Запросить КП</a>
            <a href="#" className="v2-btn v2-btn-ghost">Связаться с CISO-командой</a>
          </div>

          <div style={{marginTop:80, display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:32,
            borderTop:'1px solid var(--v2-border-soft)', paddingTop:48}}>
            <div>
              <div className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)', marginBottom:8, letterSpacing:'0.08em'}}>SALES</div>
              <a href="mailto:sales@aisecfabric.ru" style={{color:'var(--v2-fg)', textDecoration:'none', fontSize:18}}>sales@aisecfabric.ru</a>
            </div>
            <div>
              <div className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)', marginBottom:8, letterSpacing:'0.08em'}}>CISO HOTLINE 24/7</div>
              <a href="tel:+74951234567" style={{color:'var(--v2-fg)', textDecoration:'none', fontSize:18}}>+7 495 123 45 67</a>
            </div>
            <div>
              <div className="v2-mono" style={{fontSize:12, color:'var(--v2-muted)', marginBottom:8, letterSpacing:'0.08em'}}>ОФИС МОСКВА</div>
              <div style={{color:'var(--v2-fg)', fontSize:18}}>Пресненская наб., 12</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function V2Landing() {
  return (
    <>
      <V2Nav active="home"/>
      <V2HeroSlider/>
      <V2Stats/>
      <V2OrbitalSection/>
      <V2ConsolePreview/>
      <V2HowItWorks/>
      <V2ComplianceStrip/>
      <V2Services/>
      <V2Research/>
      <V2Customers/>
      <V2CTA/>
      <V2Footer/>
    </>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<V2Landing/>);
