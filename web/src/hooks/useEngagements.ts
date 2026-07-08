import { useState, useEffect } from 'react';
import type { Engagement } from '../types';
import api from '../services/api';
import i18n from '../i18n';

export const useEngagements = (productId?: number) => {
  const [engagements, setEngagements] = useState<Engagement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchEngagements = async () => {
      try {
        const url = productId ? `/engagements?product_id=${productId}` : '/engagements';
        const { data } = await api.get<Engagement[]>(url);
        setEngagements(data || []);
      } catch (err: any) {
        setError(err.message || i18n.t('errors.fetchEngagements'));
      } finally {
        setLoading(false);
      }
    };

    fetchEngagements();
  }, [productId]);

  return { engagements, loading, error };
};
