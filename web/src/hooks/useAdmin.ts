import { useState, useEffect } from 'react';
import api from '../services/api';
import i18n from '../i18n';

export interface User {
  id: number;
  username: string;
  email: string | null;
  full_name: string | null;
  global_role: string;
  is_active: boolean;
  is_admin: boolean;
}

export interface AuditLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  entity_type: string;
  entity_id: string;
  details: string;
  created_at: string;
}

export interface SystemConfig {
  [key: string]: string;
}

export const useAdmin = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [config, setConfig] = useState<SystemConfig>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = async () => {
    try {
      const { data } = await api.get('/admin/users');
      setUsers(data.users || []);
    } catch (err) {
      setError(i18n.t('errors.fetchUsers'));
    }
  };

  const fetchAuditLogs = async () => {
    try {
      const { data } = await api.get('/audit');
      setAuditLogs(data.audit_logs || []);
    } catch (err) {
      setError(i18n.t('errors.fetchAuditLogs'));
    }
  };

  const fetchConfig = async () => {
    try {
      const { data } = await api.get('/admin/config');
      setConfig(data || {});
    } catch (err) {
      setError(i18n.t('errors.fetchConfig'));
    }
  };

  const createUser = async (userData: any) => {
    try {
      await api.post('/admin/users', userData);
      await fetchUsers();
      return { ok: true };
    } catch (err: any) {
      return { ok: false, error: err.response?.data?.error || i18n.t('errors.createUser') };
    }
  };

  const deleteUser = async (id: number) => {
    try {
      await api.delete(`/admin/users?id=${id}`);
      await fetchUsers();
      return { ok: true };
    } catch (err: any) {
      return { ok: false, error: err.response?.data?.error || i18n.t('errors.deleteUser') };
    }
  };

  const updateConfig = async (newConfig: SystemConfig) => {
    try {
      await api.post('/admin/config', newConfig);
      await fetchConfig();
      return { ok: true };
    } catch (err: any) {
      return { ok: false, error: err.response?.data?.error || i18n.t('errors.updateConfig') };
    }
  };

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await Promise.all([fetchUsers(), fetchAuditLogs(), fetchConfig()]);
      setLoading(false);
    };
    init();
  }, []);

  return {
    users,
    auditLogs,
    config,
    loading,
    error,
    fetchUsers,
    fetchAuditLogs,
    fetchConfig,
    createUser,
    deleteUser,
    updateConfig,
  };
};
