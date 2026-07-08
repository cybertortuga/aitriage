import React, { useEffect, useRef, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useTitle } from '../hooks/useTitle';
import { securityService } from '../services/securityService';
import { useCopilotStore } from '../store/CopilotStore';
import { useAuthStore } from '../store/AuthStore';

/* ── V2 TUI Color Palette ────────────────────────────────── */
const C = {
  bg: 'var(--v2-bg)',
  surfaceBright: 'var(--v2-surface-2)',
  surfaceLowest: 'var(--v2-bg)',
  surface: 'var(--v2-surface)',
  surfaceHigh: 'var(--v2-elev)',
  surfaceHighest: 'var(--v2-border)',
  text: 'var(--v2-fg)',
  textVariant: 'var(--v2-fg-2)',
  outline: 'var(--v2-border-soft)',
  gray: 'var(--v2-muted)',
  primary: '#ffffff',
  primaryCont: 'var(--v2-red)',
  primaryDim: 'var(--v2-red-dim)',
  secondary: '#b8c3ff',
  tertiaryCont: '#ffdb3f',
  error: 'var(--v2-red)',
  success: '#39ff14',
};

/* ── Types ────────────────────────────────────────────────────────────── */
interface TermLine {
  spans: { text: string; color: string; bold?: boolean; bg?: string }[];
}

const span = (text: string, color: string = C.text, bold = false, bg?: string) => ({
  text,
  color,
  bold,
  bg,
});
const line = (...spans: ReturnType<typeof span>[]): TermLine => ({ spans });
const emptyLine = (): TermLine => ({ spans: [{ text: '', color: C.text }] });

/* ── Helpers ──────────────────────────────────────────────────────────── */
const pad = (s: string, n: number) => s.padEnd(n);
const rpad = (s: string, n: number) => s.padStart(n);
const sep = (w: number) => '─'.repeat(w);
const sevColor = (sev: string): string => {
  switch (sev.toUpperCase()) {
    case 'CRITICAL':
      return C.error;
    case 'HIGH':
      return C.tertiaryCont;
    case 'MEDIUM':
      return C.secondary;
    case 'LOW':
      return C.gray;
    default:
      return C.text;
  }
};

const sevBar = (
  label: string,
  count: number,
  total: number,
  color: string,
  width = 40,
): TermLine => {
  const barW = Math.max(width - 16, 4);
  const filled = total > 0 ? Math.max(count > 0 ? 1 : 0, Math.round((count / total) * barW)) : 0;
  const empty = barW - filled;
  return line(
    span(pad(label, 6), color, true),
    span(rpad(String(count), 4), color),
    span(' '),
    span('█'.repeat(filled), color),
    span('░'.repeat(empty), C.surfaceHigh),
  );
};

const chipLine = (
  key: string,
  value: string,
  keyColor = C.gray,
  valColor = C.primaryDim,
): TermLine => {
  return line(
    span(key, keyColor, true),
    span(':', C.outline),
    span(' '),
    span(value, valColor, true),
  );
};

const sectionHeader = (title: string): TermLine => {
  return line(span(`[ ${title} ]`, C.gray, true));
};

const bulletLine = (label: string, value: string, labelW = 16): TermLine => {
  return line(span(' • ', C.outline), span(pad(label, labelW), C.gray), span(value, C.text));
};

const tableRow = (
  cols: { text: string; width: number; color: string; bold?: boolean }[],
 ): TermLine => {
  return line(...cols.map((c) => span(pad(c.text, c.width), c.color, c.bold)));
};

/* ── Command Definitions ─────────────────────────────────────────────── */
const COMMANDS: Record<string, string> = {
  help: 'Display available commands',
  scan: 'Run a full security scan on the specified path (default: .)',
  rules: 'List all loaded security rules',
  findings: 'Show current findings summary',
  health: 'Check security engine status',
  metrics: 'Display system metrics and telemetry',
  whoami: 'Show current session info',
  analyze: 'AI-powered analysis of a finding by ID',
  deps: 'Show dependency audit summary',
  topology: 'Display attack surface topology',
  clear: 'Clear terminal output',
  version: 'Show version info',
};

/* ── Terminal Component ──────────────────────────────────────────────── */
export const TerminalPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('terminal.title'));
  const { setIsOpen, setContext } = useCopilotStore();
  const [lines, setLines] = useState<TermLine[]>([]);
  const [input, setInput] = useState('');
  const [history, setHistory] = useState<string[]>([]);
  const [historyIdx, setHistoryIdx] = useState(-1);
  const [startTime] = useState(() => Date.now());
  const [uptime, setUptime] = useState('00:00:00');
  const terminalEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  /* ── Boot sequence ──────────────────────────────────────────────────── */
  useEffect(() => {
    const bootLines: TermLine[] = [
      emptyLine(),
      line(
        span(' A I T R I A G E', C.primaryCont, true),
        span(' // ', C.outline),
        span(t('terminal.boot.subtitle'), C.gray),
      ),
      line(span(' ' + sep(58), C.outline)),
      emptyLine(),
      chipLine(t('terminal.boot.version'), 'v2.1.0', C.gray, C.primaryDim),
      chipLine(t('terminal.boot.engine'), t('terminal.boot.engineVal'), C.gray, C.text),
      chipLine(t('terminal.boot.status'), t('terminal.boot.statusVal'), C.gray, C.primaryDim),
      emptyLine(),
      line(
        span(t('terminal.boot.instructions.type'), C.textVariant),
        span('help', C.primaryCont, true),
        span(t('terminal.boot.instructions.forCommands'), C.textVariant),
        span('Tab', C.primaryCont, true),
        span(t('terminal.boot.instructions.forAutocomplete'), C.textVariant),
      ),
      line(span(' ' + sep(58), C.outline)),
      emptyLine(),
    ];
    const handle = requestAnimationFrame(() => {
      setLines(bootLines);
    });
    return () => cancelAnimationFrame(handle);
  }, [t]);

  /* ── Uptime clock ───────────────────────────────────────────────────── */
  useEffect(() => {
    const iv = setInterval(() => {
      const elapsed = Math.floor((Date.now() - startTime) / 1000);
      const h = String(Math.floor(elapsed / 3600)).padStart(2, '0');
      const m = String(Math.floor((elapsed % 3600) / 60)).padStart(2, '0');
      const s = String(elapsed % 60).padStart(2, '0');
      setUptime(`${h}:${m}:${s}`);
    }, 1000);
    return () => clearInterval(iv);
  }, [startTime]);

  /* ── Auto-scroll ────────────────────────────────────────────────────── */
  useEffect(() => {
    terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [lines]);

  /* ── Output helpers ─────────────────────────────────────────────────── */
  const addLines = useCallback((...newLines: TermLine[]) => {
    setLines((prev) => [...prev, ...newLines]);
  }, []);

  const addText = useCallback(
    (msg: string, color = C.text, bold = false) => {
      addLines(line(span(msg, color, bold)));
    },
    [addLines],
  );

  /* ── Command execution ──────────────────────────────────────────────── */
  const handleCommand = useCallback(
    async (rawInput: string) => {
      const trimmed = rawInput.trim();
      if (!trimmed) return;

      // Show the command in prompt style
      addLines(line(span(' ➜ ', C.primaryCont, true), span(trimmed, C.text, true)));

      const parts = trimmed.split(/\s+/);
      const cmd = parts[0].toLowerCase();
      const args = parts.slice(1);

      switch (cmd) {
        /* ── help ──────────────────────────────────────────────────────── */
        case 'help': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.availableCommands')), emptyLine());
          Object.entries(COMMANDS).forEach(([k, desc]) => {
            addLines(
              line(span(' ' + pad(k, 12), C.primaryCont, true), span('— ' + t(`terminal.commands.${k}`, desc), C.textVariant)),
            );
          });
          addLines(
            emptyLine(),
            line(span(t('terminal.shortcuts.title'), C.gray, true)),
            line(span(' ↑/↓', C.primaryCont), span(t('terminal.shortcuts.history'), C.textVariant)),
            line(span(' Tab', C.primaryCont), span(t('terminal.shortcuts.complete'), C.textVariant)),
            line(span(' Ctrl+L', C.primaryCont), span(t('terminal.shortcuts.clear'), C.textVariant)),
            emptyLine(),
          );
          break;
        }

        /* ── scan ──────────────────────────────────────────────────────── */
        case 'scan': {
          const path = args[0] || '.';
          addLines(emptyLine(), sectionHeader(t('terminal.headers.scanInit')), bulletLine(t('terminal.scan.target'), path));

          const phases = [
            t('terminal.scan.phases.init'),
            t('terminal.scan.phases.deps'),
            t('terminal.scan.phases.rules'),
            t('terminal.scan.phases.tech'),
            t('terminal.scan.phases.run'),
            t('terminal.scan.phases.entropy'),
            t('terminal.scan.phases.secrets'),
            t('terminal.scan.phases.aggregate'),
          ];

          for (const phase of phases) {
            addLines(line(span(' ⠋ ', C.primaryCont), span(phase, C.primaryDim)));
            await new Promise((r) => setTimeout(r, 80));
          }

          try {
            const res = await securityService.startScan(path);

            addLines(
              emptyLine(),
              line(span(' ✓ ', C.primaryCont), span(t('terminal.scan.completed'), C.text)),
              emptyLine(),
              sectionHeader(t('terminal.headers.scanResult')),
              emptyLine(),
            );

            const scoreColor =
              res.security_score < 50
                ? C.error
                : res.security_score < 70
                  ? C.tertiaryCont
                  : C.primaryDim;
            addLines(
              line(
                span(t('terminal.scan.scoreLabel'), C.gray, true),
                span(String(res.security_score), scoreColor, true),
                span('/100 ', C.gray),
                span(' ' + res.security_grade + ' ', C.surface, true),
              ),
            );
            if (res.health_check?.verdict) {
              const gateColor = res.health_check.verdict.passed ? C.primaryDim : C.error;
              addLines(
                line(
                  span('Security Gate: ', C.gray, true),
                  span(res.health_check.verdict.status.toUpperCase(), gateColor, true),
                  span(' · ', C.gray),
                  span(res.health_check.policy?.profile || 'baseline', C.textVariant),
                ),
              );
            }
            addLines(emptyLine());

            const sev: Record<string, number> = { CRITICAL: 0, HIGH: 0, MEDIUM: 0, LOW: 0 };
            res.findings.forEach((f) => {
              sev[f.severity.toUpperCase()] = (sev[f.severity.toUpperCase()] || 0) + 1;
            });
            const total = res.findings.length;

            addLines(
              sectionHeader(t('terminal.scan.severityFindings', { count: total })),
              emptyLine(),
              sevBar(t('terminal.severity.crit'), sev.CRITICAL, total, C.error),
              sevBar(t('terminal.severity.high'), sev.HIGH, total, C.tertiaryCont),
              sevBar(t('terminal.severity.med'), sev.MEDIUM, total, C.secondary),
              sevBar(t('terminal.severity.low'), sev.LOW, total, C.gray),
              emptyLine(),
              chipLine(t('terminal.scan.durationLabel'), res.duration, C.gray, C.text),
              chipLine(t('terminal.scan.scanIdLabel'), res.scan_id, C.gray, C.textVariant),
              emptyLine(),
            );
          } catch (err) {
            addLines(
              emptyLine(),
              line(
                span(t('terminal.scan.failed'), C.error),
                span(err instanceof Error ? err.message : String(err), C.error),
              ),
              emptyLine(),
            );
          }
          break;
        }

        /* ── rules ─────────────────────────────────────────────────────── */
        case 'rules': {
          addLines(
            emptyLine(),
            sectionHeader(t('terminal.headers.ruleEngine')),
            emptyLine(),
            line(span(t('terminal.rules.fetching'), C.primaryDim)),
          );
          try {
            const rules = await securityService.getRules();
            const stacks = new Set(rules.map((r) => r.stack).filter(Boolean));

            addLines(
              emptyLine(),
              chipLine(t('terminal.rules.loadedLabel'), t('terminal.rules.rulesCount', { count: rules.length }), C.gray, C.primaryDim),
              chipLine(
                t('terminal.rules.stacksLabel'),
                String(stacks.size) + ' (' + [...stacks].join(', ') + ')',
                C.gray,
                C.text,
              ),
              emptyLine(),
            );

            addLines(
              tableRow([
                { text: t('terminal.rules.table.sev'), width: 10, color: C.gray, bold: true },
                { text: t('terminal.rules.table.ruleId'), width: 30, color: C.gray, bold: true },
                { text: t('terminal.rules.table.name'), width: 40, color: C.gray, bold: true },
              ]),
            );
            addLines(line(span(' ' + sep(78), C.outline)));

            rules.slice(0, 15).forEach((r) => {
              addLines(
                tableRow([
                  { text: r.severity, width: 10, color: sevColor(r.severity), bold: true },
                  {
                    text: r.id.length > 28 ? r.id.slice(0, 25) + '...' : r.id,
                    width: 30,
                    color: C.textVariant,
                  },
                  {
                    text: r.name.length > 38 ? r.name.slice(0, 35) + '...' : r.name,
                    width: 40,
                    color: C.text,
                  },
                ]),
              );
            });

            if (rules.length > 15) {
              addLines(
                line(
                  span(
                    t('terminal.rules.more', { count: rules.length - 15 }),
                    C.gray,
                  ),
                ),
              );
            }
            addLines(emptyLine());
          } catch {
            addLines(line(span(t('terminal.rules.fetchFailed'), C.error)));
            addLines(emptyLine());
          }
          break;
        }

        /* ── findings ──────────────────────────────────────────────────── */
        case 'findings': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.findingsTriage')), emptyLine());
          try {
            const findings = await securityService.getFindings();
            const sev: Record<string, number> = { CRITICAL: 0, HIGH: 0, MEDIUM: 0, LOW: 0 };
            findings.forEach((f) => {
              sev[f.severity.toUpperCase()] = (sev[f.severity.toUpperCase()] || 0) + 1;
            });
            const total = findings.length;

            addLines(
              chipLine(t('terminal.findings.totalLabel'), String(total), C.gray, total > 0 ? C.error : C.primaryDim),
              emptyLine(),
              sevBar(t('terminal.severity.crit'), sev.CRITICAL, total, C.error),
              sevBar(t('terminal.severity.high'), sev.HIGH, total, C.tertiaryCont),
              sevBar(t('terminal.severity.med'), sev.MEDIUM, total, C.secondary),
              sevBar(t('terminal.severity.low'), sev.LOW, total, C.gray),
              emptyLine(),
            );

            const critical = findings
              .filter((f) => f.severity === 'CRITICAL' || f.severity === 'HIGH')
              .slice(0, 10);
            if (critical.length > 0) {
              addLines(sectionHeader(t('terminal.headers.criticalFindings')), emptyLine());
              addLines(
                tableRow([
                  { text: t('terminal.findings.table.id'), width: 16, color: C.gray, bold: true },
                  { text: t('terminal.findings.table.sev'), width: 10, color: C.gray, bold: true },
                  { text: t('terminal.findings.table.issue'), width: 32, color: C.gray, bold: true },
                  { text: t('terminal.findings.table.file'), width: 24, color: C.gray, bold: true },
                ]),
              );
              addLines(line(span(' ' + sep(80), C.outline)));
              critical.forEach((f) => {
                addLines(
                  tableRow([
                    {
                      text: f.id.length > 14 ? f.id.slice(0, 11) + '...' : f.id,
                      width: 16,
                      color: C.textVariant,
                    },
                    { text: f.severity, width: 10, color: sevColor(f.severity), bold: true },
                    {
                      text: f.name.length > 30 ? f.name.slice(0, 27) + '...' : f.name,
                      width: 32,
                      color: C.text,
                    },
                    {
                      text: f.file.length > 22 ? '...' + f.file.slice(-19) : f.file,
                      width: 24,
                      color: C.gray,
                    },
                  ]),
                );
              });
            }
            addLines(emptyLine());
          } catch {
            addLines(line(span(t('terminal.findings.fetchFailed'), C.error)), emptyLine());
          }
          break;
        }

        /* ── health ────────────────────────────────────────────────────── */
        case 'health': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.sysHealth')), emptyLine());
          try {
            const status = await securityService.getHealth();
            const tools = status.tools || {};
            const activeCount = Object.values(tools).filter(Boolean).length;
            const totalTools = Object.keys(tools).length;

            addLines(
              line(
                span(' ● ', status.ok ? C.primaryDim : C.error),
                span(status.ok ? t('terminal.health.nominal') : t('terminal.health.degraded'), status.ok ? C.primaryDim : C.error, true),
                span(' '),
                span(t('terminal.health.enginesActive', { activeCount, totalTools }), C.textVariant),
              ),
              emptyLine(),
              sectionHeader(t('terminal.headers.engineStatus')),
              emptyLine(),
            );

            Object.entries(tools).forEach(([tool, active]) => {
              const statusStr = active ? t('terminal.health.activeStatus') : t('terminal.health.offlineStatus');
              const color = active ? C.success : C.error;
              addLines(
                line(span(' ' + pad(tool.toUpperCase(), 16), C.text), span(statusStr, color, true)),
              );
            });
            addLines(emptyLine());
          } catch {
            addLines(
              line(span(' ● ', C.error), span(t('terminal.health.unreachable'), C.error, true)),
              line(span(t('terminal.health.checkFailed'), C.gray)),
              emptyLine(),
            );
          }
          break;
        }

        /* ── metrics ───────────────────────────────────────────────────── */
        case 'metrics': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.sysMetrics')), emptyLine());
          try {
            const metrics = await securityService.getMetrics();
            if (metrics) {
              Object.entries(metrics).forEach(([key, value]) => {
                addLines(bulletLine(key.toUpperCase(), String(value)));
              });
            } else {
              addLines(line(span(t('terminal.metrics.none'), C.gray)));
            }
            addLines(emptyLine());
          } catch {
            addLines(line(span(t('terminal.metrics.fetchFailed'), C.error)), emptyLine());
          }
          break;
        }

        /* ── whoami ────────────────────────────────────────────────────── */
        case 'whoami': {
          const user = useAuthStore.getState().user;
          const token = useAuthStore.getState().token;
          addLines(
            emptyLine(),
            sectionHeader(t('terminal.headers.sessionInfo')),
            emptyLine(),
            bulletLine(t('terminal.session.operatorLabel'), user?.email || user?.username || t('terminal.session.anonymous')),
            bulletLine(t('terminal.session.roleLabel'), user?.global_role?.toUpperCase() || 'GUEST'),
            bulletLine(t('terminal.session.sessionLabel'), token?.slice(0, 16) + '...' || 'N/A'),
            bulletLine(t('terminal.session.uptimeLabel'), uptime),
            emptyLine(),
          );
          break;
        }

        /* ── analyze ───────────────────────────────────────────────────── */
        case 'analyze': {
          const id = args[0];
          if (!id) {
            addLines(
              emptyLine(),
              line(span(t('terminal.analyze.usage'), C.gray), span(t('terminal.analyze.usageCmd'), C.primaryCont)),
              line(
                span(t('terminal.analyze.runPrefix'), C.gray),
                span(t('terminal.analyze.findingsCmd'), C.primaryCont),
                span(t('terminal.analyze.runSuffix'), C.gray),
              ),
              emptyLine(),
            );
            break;
          }
          addLines(
            emptyLine(),
            sectionHeader(t('terminal.headers.aiAnalysis')),
            line(span(t('terminal.analyze.analyzingPrefix'), C.gray), span(id, C.primaryCont, true)),
            line(span(t('terminal.analyze.processing'), C.primaryDim)),
          );
          try {
            const res = await securityService.analyzeFinding(id);
            addLines(emptyLine());
            if (res.ok) {
              res.analysis.split('\n').forEach((l) => {
                if (l.startsWith('##')) {
                  addLines(line(span(' ' + l.replace(/^#+\s*/, ''), C.primaryCont, true)));
                } else if (l.startsWith('- ') || l.startsWith('* ')) {
                  addLines(line(span(' • ', C.outline), span(l.slice(2), C.text)));
                } else {
                  addLines(line(span(' ' + l, C.textVariant)));
                }
              });
            } else {
              addLines(line(span(' ' + (res.error || t('terminal.analyze.failed')), C.error)));
            }
            addLines(emptyLine());
          } catch {
            addLines(line(span(t('terminal.analyze.requestFailed'), C.error)), emptyLine());
          }
          break;
        }

        /* ── topology ──────────────────────────────────────────────────── */
        case 'topology': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.attackSurface')), emptyLine());
          try {
            const findings = await securityService.getFindings();
            const files = new Map<string, number>();
            findings.forEach((f) => {
              files.set(f.file, (files.get(f.file) || 0) + 1);
            });
            const sorted = [...files.entries()].sort((a, b) => b[1] - a[1]).slice(0, 15);
            if (sorted.length > 0) {
              addLines(sectionHeader(t('terminal.headers.topFiles')), emptyLine());
              sorted.forEach(([file, count]) => {
                const barLen = Math.min(count * 3, 30);
                addLines(
                  line(
                    span(
                      ' ' + pad(file.length > 35 ? '...' + file.slice(-32) : file, 38),
                      C.textVariant,
                    ),
                    span(rpad(String(count), 4), C.text, true),
                    span(
                      '█'.repeat(barLen),
                      count > 5 ? C.error : count > 2 ? C.tertiaryCont : C.primaryDim,
                    ),
                  ),
                );
              });
            } else {
              addLines(
                line(
                  span(t('terminal.topology.noDataPrefix'), C.gray),
                  span('scan', C.primaryCont),
                  span(t('terminal.topology.noDataSuffix'), C.gray),
                ),
              );
            }
            addLines(emptyLine());
          } catch {
            addLines(line(span(t('terminal.topology.fetchFailed'), C.error)), emptyLine());
          }
          break;
        }

        /* ── deps ───────────────────────────────────────────────────────── */
        case 'deps': {
          addLines(emptyLine(), sectionHeader(t('terminal.headers.depsAudit')), emptyLine());
          addLines(line(span(t('terminal.deps.noData'), C.gray)));
          addLines(
            line(
              span(t('terminal.deps.usePrefix'), C.gray),
              span(t('terminal.deps.scanCmd'), C.primaryCont),
              span(t('terminal.deps.useSuffix'), C.gray),
            ),
          );
          addLines(emptyLine());
          break;
        }

        /* ── version ───────────────────────────────────────────────────── */
        case 'version': {
          addLines(
            emptyLine(),
            line(span(t('terminal.version.title'), C.primaryCont, true)),
            chipLine(t('terminal.version.versionLabel'), 'v2.1.0', C.gray, C.primaryDim),
            chipLine(t('terminal.version.engineLabel'), t('terminal.version.engineVal'), C.gray, C.text),
            chipLine(t('terminal.version.llmLabel'), t('terminal.version.llmVal'), C.gray, C.secondary),
            chipLine(t('terminal.version.webUiLabel'), t('terminal.version.webUiVal'), C.gray, C.text),
            emptyLine(),
          );
          break;
        }

        /* ── clear ──────────────────────────────────────────────────────── */
        case 'clear': {
          setLines([]);
          break;
        }

        /* ── unknown ────────────────────────────────────────────────────── */
        default: {
          addLines(
            emptyLine(),
            line(span(t('terminal.unknown.notFound'), C.error), span(cmd, C.error, true)),
            line(
              span(t('terminal.unknown.typePrefix'), C.gray),
              span('help', C.primaryCont, true),
              span(t('terminal.unknown.typeSuffix'), C.gray),
            ),
            emptyLine(),
          );
        }
      }
    },
    [addLines, addText, uptime, t],
  );

  /* ── Form submit ────────────────────────────────────────────────────── */
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    setHistory((prev) => [input.trim(), ...prev.slice(0, 49)]);
    setHistoryIdx(-1);
    handleCommand(input.trim());
    setInput('');
  };

  /* ── Keyboard handling ──────────────────────────────────────────────── */
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (history.length > 0) {
        const newIdx = Math.min(historyIdx + 1, history.length - 1);
        setHistoryIdx(newIdx);
        setInput(history[newIdx]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIdx > 0) {
        const newIdx = historyIdx - 1;
        setHistoryIdx(newIdx);
        setInput(history[newIdx]);
      } else {
        setHistoryIdx(-1);
        setInput('');
      }
    } else if (e.key === 'Tab') {
      e.preventDefault();
      const partial = input.trim().toLowerCase();
      if (partial) {
        const matches = Object.keys(COMMANDS).filter((c) => c.startsWith(partial));
        if (matches.length === 1) {
          setInput(matches[0] + ' ');
        } else if (matches.length > 1) {
          addLines(line(span(' ' + matches.join(' '), C.primaryDim)));
        }
      }
    } else if (e.key === 'l' && e.ctrlKey) {
      e.preventDefault();
      setLines([]);
    }
  };

  return (
    <div className="flex flex-col min-h-full v2-mono" style={{ background: C.bg }}>
      {/* ── Header Bar ──────────────────────────────────── */}
      <div
        className="shrink-0 flex items-center justify-between px-5 h-12 border-b"
        style={{ borderColor: C.outline, background: C.surfaceLowest }}
      >
        <div className="flex items-center gap-4">
          <span className="v2-tag v2-tag-red">AITRIAGE</span>
          <span className="text-[11px] font-bold text-white tracking-widest uppercase">
            {t('terminal.header.title')}
          </span>
          <span className="text-[11px]" style={{ color: C.outline }}>
            │
          </span>
          <span className="text-[11px] flex items-center gap-1.5 uppercase font-bold text-v2-muted">
            <span className="w-1.5 h-1.5 bg-v2-red animate-pulse rounded-full" />
            <span style={{ color: C.primaryCont }}>{t('terminal.header.statusNominal')}</span>
          </span>
        </div>

        <div className="flex items-center gap-4">
          <span className="text-[10px] text-v2-muted">
            {t('terminal.header.uptime')}<span style={{ color: C.outline }}>:</span>{' '}
            <span style={{ color: C.primaryCont }}>{uptime}</span>
          </span>
          <span className="text-[11px]" style={{ color: C.outline }}>
            │
          </span>
          <button
            onClick={() => {
              setContext(
                `Terminal Session History:\n${lines.map((l) => l.spans.map((s) => s.text).join('')).join('\n')}`,
              );
              setIsOpen(true);
            }}
            className="flex items-center gap-1.5 text-[10px] font-bold tracking-widest text-v2-red hover:text-white transition-colors uppercase"
          >
            <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
              smart_toy
            </span>
            {t('terminal.header.copilot')}
          </button>
        </div>
      </div>

      {/* ── Terminal Body ───────────────────────────────────────────── */}
      <div
        ref={containerRef}
        className="flex-1 overflow-y-auto cyber-scrollbar px-6 py-6 cursor-text"
        style={{ background: C.bg }}
        onClick={() => inputRef.current?.focus()}
      >
        {lines.map((termLine, i) => (
          <div
            key={i}
            className="leading-[1.7] whitespace-pre"
            style={{ fontSize: '13px', minHeight: '20px' }}
          >
            {termLine.spans.map((s, j) => (
              <span
                key={j}
                style={{
                  color: s.color,
                  fontWeight: s.bold ? 700 : 400,
                  background: s.bg,
                }}
              >
                {s.text}
              </span>
            ))}
          </div>
        ))}

        {/* ── Prompt ───────────────────────────────────────────────── */}
        <form
          onSubmit={handleSubmit}
          className="flex items-center leading-[1.7] mt-1"
          style={{ fontSize: '13px' }}
        >
          <span style={{ color: C.primaryCont, fontWeight: 700 }}> ➜ </span>
          <span style={{ color: C.gray }}>~ </span>
          <input
            ref={inputRef}
            type="text"
            className="bg-transparent border-none outline-none flex-1 ml-1 text-white"
            style={{
              fontFamily: 'inherit',
              fontSize: 'inherit',
              caretColor: C.primaryCont,
            }}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            autoFocus
            spellCheck={false}
            autoComplete="off"
          />
          <div
            className="w-[8px] h-[16px] animate-pulse rounded-sm"
            style={{ background: C.primaryCont }}
          />
        </form>
        <div ref={terminalEndRef} className="h-4" />
      </div>

      {/* ── Footer Bar ──────────────────────────────────── */}
      <div
        className="shrink-0 flex items-center justify-between px-6 h-8 border-t text-[10px]"
        style={{ borderColor: C.outline, background: C.surfaceLowest, color: C.gray }}
      >
        <div className="flex items-center gap-3">
          <span>
            <span style={{ color: C.primaryCont }}>↑↓</span> {t('terminal.footer.history')}
          </span>
          <span style={{ color: C.outline }}>│</span>
          <span>
            <span style={{ color: C.primaryCont }}>Tab</span> {t('terminal.footer.complete')}
          </span>
          <span style={{ color: C.outline }}>│</span>
          <span>
            <span style={{ color: C.primaryCont }}>Ctrl+L</span> {t('terminal.footer.clear')}
          </span>
        </div>
        <div>{history.length > 0 && <span>{t('terminal.footer.commandsCount', { count: history.length })}</span>}</div>
      </div>
    </div>
  );
};
