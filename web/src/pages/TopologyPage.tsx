import React, { useState, useEffect, useRef } from 'react';
import { useTitle } from '../hooks/useTitle';
import { useCopilotStore } from '../store/CopilotStore';
import ForceGraph2D from 'react-force-graph-2d';
import api from '../services/api';
import { useTranslation } from 'react-i18next';

export const TopologyPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('topology.title'));

  const getTypeDescription = (type: string) => {
    switch (type?.toUpperCase()) {
      case 'APPLICATION':
        return t('topology.typeDescriptions.application');
      case 'DATABASE':
        return t('topology.typeDescriptions.database');
      case 'CACHE':
        return t('topology.typeDescriptions.cache');
      case 'PROXY':
        return t('topology.typeDescriptions.proxy');
      case 'STORAGE':
        return t('topology.typeDescriptions.storage');
      case 'MESSAGE_BROKER':
        return t('topology.typeDescriptions.messageBroker');
      default:
        return t('topology.typeDescriptions.default');
    }
  };

  const [graphData, setGraphData] = useState<{ nodes: any[]; links: any[] }>({
    nodes: [],
    links: [],
  });
  const [selectedNode, setSelectedNode] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const { setIsOpen, setContext } = useCopilotStore();
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.clientWidth,
          height: containerRef.current.clientHeight,
        });
      }
    };

    window.addEventListener('resize', updateDimensions);
    updateDimensions();
    return () => window.removeEventListener('resize', updateDimensions);
  }, []);

  useEffect(() => {
    api
      .get('/topology')
      .then((res) => {
        if (res.data.ok) {
          setGraphData({
            nodes: res.data.nodes || [],
            links: res.data.links || [],
          });

          if (res.data.nodes?.length > 0) {
            const crit = res.data.nodes.find(
              (n: any) => n.risk === 'CRITICAL' || n.risk === 'CRIT',
            );
            setSelectedNode(crit || res.data.nodes[0]);
          }
        }
      })
      .catch((err) => console.error('Failed to fetch topology:', err))
      .finally(() => setLoading(false));
  }, []);

  const handleAIConsult = () => {
    if (!selectedNode) return;
    setContext(
      t('topology.copilotContext', {
        id: selectedNode.id,
        type: selectedNode.type || t('topology.systemNodeDefault'),
        risk: selectedNode.risk,
        status: selectedNode.status || t('topology.statusUnknown'),
      })
    );
    setIsOpen(true);
  };

  const getRiskColor = (risk: string) => {
    switch (risk?.toUpperCase()) {
      case 'CRITICAL':
      case 'CRIT':
        return 'var(--accent-color)';
      case 'HIGH':
        return 'var(--accent-color-hover)';
      case 'MEDIUM':
      case 'MED':
        return '#f59e0b';
      default:
        return 'rgba(244, 244, 245, 0.4)';
    }
  };

  return (
    <div className="flex flex-col h-full overflow-hidden bg-v2-bg">
      {/* Page Header */}
      <div className="px-6 py-4 border-b border-v2-border-soft flex justify-between items-center shrink-0 bg-v2-surface">
        <div>
          <p className="text-[9px] font-bold tracking-widest text-v2-muted mb-1 uppercase">
            {t('topology.breadcrumb')}
          </p>
          <h1 className="text-xl font-bold tracking-tight text-white uppercase">{t('topology.header')}</h1>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 px-3 py-1.5 bg-v2-bg border border-v2-border-soft rounded-lg">
            <div className="w-1.5 h-1.5 rounded-full bg-success animate-pulse" />
            <span className="text-[10px] font-bold tracking-widest text-success uppercase">
              {t('topology.scannerOnline')}
            </span>
          </div>
          <button onClick={() => setLoading(true)} className="v2-btn v2-btn-ghost px-4 h-8">
            <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
              refresh
            </span>
            <span>{t('topology.rescanInfra')}</span>
          </button>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Graph Area */}
        <div ref={containerRef} className="flex-1 border-r border-v2-border-soft relative bg-v2-bg">
          <div
            className="absolute inset-0 opacity-10 pointer-events-none"
            style={{
              backgroundImage:
                'radial-gradient(circle at 1px 1px, var(--v2-border-soft) 1px, transparent 0)',
              backgroundSize: '32px 32px',
            }}
          />

          {loading ? (
            <div className="absolute inset-0 flex items-center justify-center z-20">
              <div className="animate-pulse text-[11px] font-bold tracking-widest text-v2-red uppercase">
                {t('topology.initializing')}
              </div>
            </div>
          ) : graphData.nodes.length === 0 ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center z-20 p-8 text-center bg-v2-bg">
              <span className="material-symbols-outlined text-4xl mb-4 text-v2-muted">
                account_tree
              </span>
              <h3 className="text-[13px] font-bold text-white mb-2 uppercase tracking-widest">
                {t('topology.noData')}
              </h3>
              <p className="text-[12px] text-v2-fg-2 max-w-sm mb-6 leading-relaxed">
                {t('topology.noDataDesc')}
              </p>
              <button
                onClick={async () => {
                  setLoading(true);
                  try {
                    await api.post('/scan', { path: '.' });
                    window.location.reload();
                  } catch (err) {
                    console.error('Scan failed', err);
                  } finally {
                    setLoading(false);
                  }
                }}
                className="v2-btn v2-btn-red"
              >
                <span className="material-symbols-outlined text-sm">bolt</span>
                <span>{t('topology.runScan')}</span>
              </button>
            </div>
          ) : (
            <>
              <div className="absolute top-4 left-4 bg-v2-surface-2 border border-v2-border-soft p-4 rounded-xl z-10 flex flex-col gap-3 pointer-events-none text-mono select-none">
                <div className="text-[9px] font-bold tracking-widest uppercase text-v2-muted">
                  {t('topology.legendTitle')}
                </div>
                <div className="flex flex-col gap-2.5 text-[11px] font-bold text-v2-fg-2">
                  <div className="flex items-center gap-3">
                    <span className="w-3.5 h-3.5 border-2 border-white/50 rounded-full inline-block"></span>
                    <span>{t('topology.nodeTypes.application')}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className="w-3.5 h-3.5 bg-v2-muted inline-block"
                      style={{
                        clipPath: 'polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)',
                      }}
                    ></span>
                    <span>{t('topology.nodeTypes.database')}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className="w-3.5 h-3.5 bg-v2-muted inline-block"
                      style={{ clipPath: 'polygon(50% 0%, 100% 100%, 0% 100%)' }}
                    ></span>
                    <span>{t('topology.nodeTypes.cache')}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className="w-3.5 h-3.5 bg-v2-muted inline-block"
                      style={{ clipPath: 'polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)' }}
                    ></span>
                    <span>{t('topology.nodeTypes.proxy')}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="w-3.5 h-3.5 bg-v2-muted inline-block"></span>
                    <span>{t('topology.nodeTypes.storage')}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="w-3.5 h-2 bg-v2-muted inline-block rounded-full"></span>
                    <span>{t('topology.nodeTypes.messageBroker')}</span>
                  </div>
                </div>
              </div>

              <ForceGraph2D
                graphData={graphData}
                width={dimensions.width}
                height={dimensions.height}
                nodeLabel={(node: any) => node.name || node.id}
                nodeColor={(node: any) => getRiskColor(node.risk)}
                nodeRelSize={6}
                linkColor={() => 'rgba(255, 255, 255, 0.15)'}
                linkDirectionalParticles={1}
                linkDirectionalParticleSpeed={0.003}
                linkDirectionalParticleWidth={1.5}
                onNodeClick={(node) => setSelectedNode(node)}
                cooldownTicks={100}
                d3AlphaDecay={0.02}
                d3VelocityDecay={0.3}
                nodeCanvasObject={(node: any, ctx, globalScale) => {
                  const label = node.name || node.id;
                  const fontSize = 10 / globalScale;
                  const riskColor = getRiskColor(node.risk);
                  const typeUpper = node.type?.toUpperCase();

                  // Pulsing outer ring
                  const t = Date.now() / 1000;
                  const pulse = Math.sin(t * 3) * 1.8 + 1.8;

                  ctx.fillStyle = riskColor;

                  // Draw distinct visual shape based on node type
                  if (typeUpper === 'APPLICATION') {
                    ctx.beginPath();
                    ctx.arc(node.x, node.y, 6.5, 0, 2 * Math.PI, false);
                    ctx.fill();
                    ctx.lineWidth = 1.5 / globalScale;
                    ctx.strokeStyle = '#ffffff';
                    ctx.stroke();
                  } else if (typeUpper === 'DATABASE') {
                    ctx.beginPath();
                    const size = 6;
                    for (let i = 0; i < 6; i++) {
                      const angle = (i * Math.PI) / 3;
                      ctx.lineTo(node.x + size * Math.cos(angle), node.y + size * Math.sin(angle));
                    }
                    ctx.closePath();
                    ctx.fill();
                  } else if (typeUpper === 'CACHE') {
                    ctx.beginPath();
                    const size = 6;
                    ctx.moveTo(node.x, node.y - size);
                    ctx.lineTo(node.x + size, node.y + size * 0.85);
                    ctx.lineTo(node.x - size, node.y + size * 0.85);
                    ctx.closePath();
                    ctx.fill();
                  } else if (typeUpper === 'PROXY') {
                    ctx.beginPath();
                    const size = 6;
                    ctx.moveTo(node.x, node.y - size);
                    ctx.lineTo(node.x + size, node.y);
                    ctx.lineTo(node.x, node.y + size);
                    ctx.lineTo(node.x - size, node.y);
                    ctx.closePath();
                    ctx.fill();
                  } else if (typeUpper === 'STORAGE') {
                    ctx.beginPath();
                    const size = 5;
                    ctx.rect(node.x - size, node.y - size, size * 2, size * 2);
                    ctx.fill();
                  } else if (typeUpper === 'MESSAGE_BROKER') {
                    ctx.beginPath();
                    ctx.ellipse(node.x, node.y, 7, 4.5, 0, 0, 2 * Math.PI);
                    ctx.fill();
                  } else {
                    ctx.beginPath();
                    ctx.arc(node.x, node.y, 4.5, 0, 2 * Math.PI, false);
                    ctx.fill();
                  }

                  // Draw selection highlight or pulse
                  const baseRadius = typeUpper === 'APPLICATION' ? 6.5 : 6;
                  if (selectedNode?.id === node.id) {
                    ctx.beginPath();
                    ctx.arc(node.x, node.y, baseRadius + 3.5 + pulse, 0, 2 * Math.PI, false);
                    ctx.strokeStyle = riskColor;
                    ctx.lineWidth = 1.2 / globalScale;
                    ctx.globalAlpha = 0.5;
                    ctx.stroke();
                    ctx.globalAlpha = 1.0;
                  } else if (
                    node.risk === 'CRITICAL' ||
                    node.risk === 'CRIT' ||
                    node.risk === 'HIGH'
                  ) {
                    ctx.beginPath();
                    ctx.arc(node.x, node.y, baseRadius + 2.5 + pulse, 0, 2 * Math.PI, false);
                    ctx.strokeStyle = riskColor;
                    ctx.lineWidth = 0.8 / globalScale;
                    ctx.globalAlpha = 0.3;
                    ctx.stroke();
                    ctx.globalAlpha = 1.0;
                  }

                  // Draw label
                  ctx.font = `${fontSize}px "Geist Mono", monospace`;
                  ctx.textAlign = 'center';
                  ctx.textBaseline = 'middle';
                  ctx.fillStyle =
                    selectedNode?.id === node.id ? '#ffffff' : 'rgba(255, 255, 255, 0.6)';
                  ctx.fillText(label, node.x, node.y + 13.5);
                }}
              />
            </>
          )}
        </div>

        {/* Detail Pane */}
        <div className="w-80 shrink-0 flex flex-col overflow-y-auto cyber-scrollbar border-l border-v2-border-soft bg-v2-surface-2">
          {selectedNode ? (
            <div className="p-6 flex flex-col gap-6">
              <div>
                <div className="text-[9px] font-bold tracking-widest text-v2-muted mb-2 uppercase">
                  {t('topology.nodeIdentity')}
                </div>
                <h2 className="text-xl text-white font-bold tracking-tight uppercase break-words">
                  {selectedNode.id}
                </h2>
              </div>

              <div
                className={`p-5 rounded-xl border ${['CRITICAL', 'CRIT'].includes(selectedNode.risk?.toUpperCase()) ? 'border-v2-red-line bg-v2-red-soft' : 'border-v2-border-soft bg-v2-bg'} flex flex-col gap-3`}
              >
                <div className="flex items-center gap-2">
                  <div
                    className={`w-2 h-2 rounded-full ${['CRITICAL', 'CRIT'].includes(selectedNode.risk?.toUpperCase()) ? 'bg-v2-red animate-pulse' : 'bg-white'}`}
                  />
                  <span
                    className={`text-[11px] font-bold tracking-widest uppercase ${['CRITICAL', 'CRIT'].includes(selectedNode.risk?.toUpperCase()) ? 'text-v2-red' : 'text-white'}`}
                  >
                    {selectedNode.risk?.toUpperCase()}{t('topology.riskSuffix')}
                  </span>
                </div>
                <div className="font-mono text-[11px] text-v2-fg-2 uppercase flex flex-col gap-1">
                  <span>{t('topology.statusLabel')} {selectedNode.status || t('topology.statusUnknown')}</span>
                  <span>{t('topology.typeLabel')} {selectedNode.type || t('topology.systemNodeDefault')}</span>
                </div>
                <div className="text-[10px] text-v2-muted italic mt-1 font-mono uppercase tracking-tight opacity-70">
                  {getTypeDescription(selectedNode.type)}
                </div>
              </div>

              <div>
                <div className="text-[9px] font-bold tracking-widest text-v2-muted mb-3 border-b border-v2-border-soft pb-2 uppercase">
                  {t('topology.systemIntelligence')}
                </div>
                <div className="flex flex-col gap-3">
                  <div className="flex justify-between items-center">
                    <span className="text-[11px] font-bold text-v2-muted uppercase tracking-wider">
                      {t('topology.lastAudit')}
                    </span>
                    <span className="font-mono text-[11px] text-white">
                      {selectedNode.created_at
                        ? new Date(selectedNode.created_at).toLocaleDateString()
                        : t('topology.notAvailable')}
                    </span>
                  </div>
                </div>
              </div>

              <div className="mt-auto flex flex-col gap-3 pt-6">
                <button onClick={handleAIConsult} className="v2-btn v2-btn-red w-full">
                  <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                    smart_toy
                  </span>
                  <span>{t('topology.aiConsult')}</span>
                </button>

                <button className="v2-btn v2-btn-ghost w-full">
                  <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>
                    analytics
                  </span>
                  <span>{t('topology.viewFullTrace')}</span>
                </button>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center p-8 text-center opacity-50">
              <span className="material-symbols-outlined text-4xl mb-4 text-v2-muted">
                account_tree
              </span>
              <span className="text-[10px] font-bold tracking-widest text-v2-muted uppercase">
                {t('topology.selectNode')}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
