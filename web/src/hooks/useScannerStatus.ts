import { useState, useEffect } from 'react';
import api from '../services/api';

export interface ScannerStatus {
  ok: boolean;
  tools: Record<string, boolean>;
  loading: boolean;
}

export const useScannerStatus = () => {
  const [status, setStatus] = useState<ScannerStatus>({
    ok: false,
    tools: {},
    loading: true,
  });

  useEffect(() => {
    const checkStatus = async () => {
      try {
        const { data } = await api.get('/health');
        setStatus({
          ok: data.ok,
          tools: data.tools || {},
          loading: false,
        });
      } catch (err) {
        setStatus((prev) => ({ ...prev, ok: false, loading: false }));
      }
    };

    checkStatus();
    const interval = setInterval(checkStatus, 15000); // Pulse every 15s
    return () => clearInterval(interval);
  }, []);

  return status;
};
