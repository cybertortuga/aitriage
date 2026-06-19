import { useState, useCallback, useRef } from 'react';

export interface PipelineStats {
  tp: number;
  fp: number;
  nr: number;
  poc: number;
  total: number;
}

export interface PipelineUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

export interface PipelineStep {
  step: number;
  total: number;
  label: string;
  progress: number;
  warning?: string;
  stats?: PipelineStats;
}

export type PipelineStatus = 'idle' | 'running' | 'done' | 'error';

export interface PipelineResult {
  report: string;
  fix_spec: string;
  stats: PipelineStats;
  usage: PipelineUsage;
}

export const usePipeline = () => {
  const [status, setStatus] = useState<PipelineStatus>('idle');
  const [progress, setProgress] = useState(0);
  const [currentStep, setCurrentStep] = useState<PipelineStep | null>(null);
  const [steps, setSteps] = useState<PipelineStep[]>([]);
  const [result, setResult] = useState<PipelineResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const runPipeline = useCallback((productId: number) => {
    // Reset state
    setStatus('running');
    setProgress(0);
    setCurrentStep(null);
    setSteps([]);
    setResult(null);
    setError(null);

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const es = new EventSource(`/api/pipeline?product_id=${productId}`);
    eventSourceRef.current = es;

    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.error) {
          setStatus('error');
          setError(data.error);
          es.close();
          return;
        }

        if (data.done) {
          setStatus('done');
          setProgress(100);
          setResult({
            report: data.report || '',
            fix_spec: data.fix_spec || '',
            stats: data.stats || { tp: 0, fp: 0, nr: 0, poc: 0, total: 0 },
            usage: data.usage || { prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 },
          });
          es.close();
          return;
        }

        // Progress update
        if (data.progress !== undefined) {
          setProgress(data.progress);
        }

        const step: PipelineStep = {
          step: data.step ?? 0,
          total: data.total ?? 4,
          label: data.label ?? '',
          progress: data.progress ?? 0,
          warning: data.warning,
          stats: data.stats,
        };

        setCurrentStep(step);
        setSteps((prev) => {
          // Replace or append step
          const existing = prev.findIndex((s) => s.step === step.step && s.label === step.label);
          if (existing >= 0) {
            const updated = [...prev];
            updated[existing] = step;
            return updated;
          }
          return [...prev, step];
        });
      } catch {
        // Ignore parse errors
      }
    };

    es.onerror = () => {
      if (es.readyState === EventSource.CLOSED) {
        return; // Normal close
      }
      setStatus('error');
      setError('Connection to pipeline lost');
      es.close();
    };
  }, []);

  const cancel = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setStatus('idle');
  }, []);

  return {
    status,
    progress,
    currentStep,
    steps,
    result,
    error,
    runPipeline,
    cancel,
  };
};
