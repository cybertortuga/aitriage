import { useState, useEffect } from 'react';
import api from '../services/api';
import type { Finding } from '../types';

export interface PromptTemplate {
  id: string;
  label: string;
  icon: string;
  description: string;
  template: string;
}

/**
 * Interpolate {{placeholder}} values in a template string with finding data.
 */
export function interpolatePrompt(template: string, f: Finding): string {
  return template
    .replace(/\{\{title\}\}/g, f.title || 'N/A')
    .replace(/\{\{severity\}\}/g, f.severity || 'N/A')
    .replace(/\{\{rule_id\}\}/g, f.rule_id || 'N/A')
    .replace(/\{\{file\}\}/g, f.file_path || f.file || 'N/A')
    .replace(/\{\{line\}\}/g, String(f.line_number || 0))
    .replace(/\{\{stack\}\}/g, f.stack || 'unknown')
    .replace(/\{\{cwe_id\}\}/g, f.cwe_id || 'Not mapped')
    .replace(/\{\{description\}\}/g, f.description || f.fix_suggestion || f.suggestion || 'No description available.');
}

/**
 * Hook to fetch unified prompt templates from the server.
 * These templates are the single source of truth shared with CI/CD pipeline.
 */
export const usePrompts = () => {
  const [templates, setTemplates] = useState<PromptTemplate[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get('/prompts')
      .then(({ data }) => {
        if (data.ok && Array.isArray(data.templates)) {
          setTemplates(data.templates);
        }
      })
      .catch((err) => {
        console.error('Failed to fetch prompt templates', err);
      })
      .finally(() => setLoading(false));
  }, []);

  return { templates, loading };
};
