// Main app composition + Tweaks
const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accent": "#d4ff3a",
  "density": "comfortable",
  "showMarquee": true,
  "showPrinciples": true
}/*EDITMODE-END*/;

function App() {
  const [t, setTweak] = useTweaks(TWEAK_DEFAULTS);

  React.useEffect(() => {
    const root = document.documentElement;
    if (t.accent) {
      root.style.setProperty('--accent', t.accent);
      root.style.setProperty('--accent-soft', t.accent + '24');
      root.style.setProperty('--accent-border', t.accent + '66');
      root.style.setProperty('--accent-ink', '#0e1604');
    }
    root.style.setProperty('--container',
      t.density === 'compact' ? '1180px' :
      t.density === 'wide' ? '1380px' : '1280px');
  }, [t]);

  return (
    <>
      <NavBar/>
      <Hero/>
      {t.showMarquee && <IntegrationStrip/>}
      <Architecture/>
      <Capabilities/>
      <LiveDashboard/>
      <Compliance/>
      {t.showPrinciples && <Principles/>}
      <Roadmap/>
      <CTA/>
      <Footer/>

      <TweaksPanel>
        <TweakSection label="Брендинг"/>
        <TweakColor label="Сигнальный цвет" value={t.accent}
          options={['#d4ff3a','#00ffa3','#74f5ff','#ff7a3d','#c4a4ff','#ffd84a']}
          onChange={(v) => setTweak('accent', v)}/>

        <TweakSection label="Layout"/>
        <TweakRadio label="Плотность" value={t.density}
          options={['compact','comfortable','wide']}
          onChange={(v) => setTweak('density', v)}/>

        <TweakSection label="Секции"/>
        <TweakToggle label="Marquee «Built on»" value={t.showMarquee}
          onChange={(v) => setTweak('showMarquee', v)}/>
        <TweakToggle label="Секция «Принципы»" value={t.showPrinciples}
          onChange={(v) => setTweak('showPrinciples', v)}/>
      </TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App/>);
