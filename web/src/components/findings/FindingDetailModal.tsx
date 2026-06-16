import React, { useState, useEffect } from 'react';
import {
  Code,
  Cpu,
  XCircle,
  Clock,
  User,
  ExternalLink,
  MessageSquare,
  ShieldAlert,
} from 'lucide-react';
import { motion } from 'framer-motion';
import ReactMarkdown from 'react-markdown';
import Modal from '../common/Modal';
import type { Finding } from '../../types';
import api from '../../services/api';
import { useTranslation } from 'react-i18next';

interface FindingDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  finding: Finding | null;
  onUpdate?: (updatedFinding: Finding) => void;
}

const FindingDetailModal: React.FC<FindingDetailModalProps> = ({
  isOpen,
  onClose,
  finding,
  onUpdate,
}) => {
  const { t } = useTranslation('components');
  const [activeTab, setActiveTab] = useState<'details' | 'code' | 'ai'>('details');
  const [fileContent, setFileContent] = useState<string>('');
  const [aiAnalysis, setAiAnalysis] = useState<string>('');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);

  useEffect(() => {
    if (isOpen && finding?.file_path) {
      fetchFileContent();
      // Reset AI analysis when switching findings
      setAiAnalysis(finding.ai_analysis || '');
      setActiveTab('details');
    }
  }, [isOpen, finding]);

  const fetchFileContent = async () => {
    if (!finding?.file_path) return;
    try {
      const { data } = await api.get(`/file?path=${finding.file_path}`);
      if (data.ok) {
        setFileContent(data.content);
      }
    } catch (err) {
      console.error('Failed to fetch file content:', err);
      setFileContent(t('components.findingDetailModal.errorLoadingFile'));
    }
  };

  const handleRunAI = async () => {
    if (!finding) return;
    setIsAnalyzing(true);
    try {
      const { data } = await api.post('/analyze', {
        id: finding.id.toString(),
        type: 'finding',
      });
      if (data.ok) {
        setAiAnalysis(data.analysis);
        // In a real app, we'd also update the finding in the parent state
      }
    } catch (err) {
      console.error('AI Analysis failed:', err);
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleUpdateStatus = async (newStatus: string) => {
    if (!finding) return;
    setIsUpdating(true);
    try {
      const { data } = await api.put(`/findings/${finding.id}`, {
        action: 'status',
        status: newStatus,
      });
      if (data.ok && onUpdate) {
        onUpdate({ ...finding, status: newStatus });
      }
    } catch (err) {
      console.error('Failed to update status:', err);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleTriageClick = async () => {
    await handleUpdateStatus('triage');
    setActiveTab('ai');
    handleRunAI();
  };

  if (!finding) return null;

  const severityColor =
    {
      critical: 'text-error border-error bg-error/10',
      high: 'text-error border-error bg-error/5',
      medium: 'text-tertiary-container border-tertiary-container bg-tertiary-container/5',
      low: 'text-outline border-outline bg-outline/5',
    }[finding.severity.toLowerCase()] || 'text-outline border-outline';

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`[${finding.rule_id}] ${finding.title}`}
      maxWidth="max-w-6xl"
    >
      <div className="flex flex-col h-full space-y-6">
        {/* Top Info Bar */}
        <div className="flex items-center space-x-4">
          <div
            className={`px-3 py-1 border text-[10px] font-black uppercase tracking-widest ${severityColor}`}
          >
            {finding.severity}
          </div>
          <div className="px-3 py-1 border border-outline-variant bg-surface-container-low text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
            {finding.status}
          </div>
          <div className="flex-1" />
          <div className="flex items-center space-x-2 text-[11px] text-on-surface-variant font-mono">
            <Clock size={14} />
            <span>{new Date(finding.created_at).toLocaleString()}</span>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-outline-variant space-x-8">
          {[
            { id: 'details', label: t('components.findingDetailModal.tabs.details'), icon: ShieldAlert },
            { id: 'code', label: t('components.findingDetailModal.tabs.code'), icon: Code },
            { id: 'ai', label: t('components.findingDetailModal.tabs.ai'), icon: Cpu },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`pb-4 flex items-center space-x-2 text-[11px] font-black uppercase tracking-widest transition-none cursor-pointer relative ${
                activeTab === tab.id
                  ? 'text-primary-fixed-dim'
                  : 'text-on-surface-variant hover:text-on-surface'
              }`}
            >
              <tab.icon size={16} />
              <span>{tab.label}</span>
              {activeTab === tab.id && (
                <motion.div
                  layoutId="activeTab"
                  className="absolute bottom-0 left-0 right-0 h-[2px] bg-primary-fixed-dim"
                />
              )}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-hidden min-h-[400px]">
          {activeTab === 'details' && (
            <div className="grid grid-cols-3 gap-8 h-full">
              <div className="col-span-2 space-y-6">
                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-2">{t('components.findingDetailModal.sections.description')}</h3>
                  <p className="text-body-base text-on-surface leading-relaxed">
                    {finding.description || t('components.findingDetailModal.noDescription')}
                  </p>
                </section>

                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-2">{t('components.findingDetailModal.sections.impact')}</h3>
                  <p className="text-body-base text-on-surface leading-relaxed italic">
                    {finding.impact || t('components.findingDetailModal.impactPending')}
                  </p>
                </section>

                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-2">{t('components.findingDetailModal.sections.fileLocation')}</h3>
                  <div className="p-3 border border-outline-variant bg-surface-container-lowest font-mono text-[12px] flex items-center justify-between">
                    <span className="text-on-surface">
                      {finding.file_path}:{finding.line_number}
                    </span>
                    <button className="text-primary-fixed-dim hover:underline flex items-center space-x-1 cursor-pointer">
                      <ExternalLink size={12} />
                      <span className="text-[10px]">{t('components.findingDetailModal.openInExplorer')}</span>
                    </button>
                  </div>
                </section>
              </div>

              <div className="space-y-6 border-l border-outline-variant pl-8">
                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-3">{t('components.findingDetailModal.sections.triageActions')}</h3>
                  <div className="space-y-2">
                    <button
                      onClick={() => handleUpdateStatus('in_progress')}
                      disabled={isUpdating || finding.status === 'in_progress'}
                      className="w-full btn-mechanical flex items-center justify-center space-x-2"
                    >
                      <Clock size={14} />
                      <span>{t('components.findingDetailModal.actions.markInProgress')}</span>
                    </button>
                    <button
                      onClick={handleTriageClick}
                      disabled={isUpdating || finding.status === 'triage'}
                      className="w-full btn-mechanical-primary flex items-center justify-center space-x-2"
                    >
                      <Cpu size={14} />
                      <span>{t('components.findingDetailModal.actions.markAsTriage')}</span>
                    </button>
                    <button
                      onClick={() => handleUpdateStatus('false_positive')}
                      disabled={isUpdating || finding.status === 'false_positive'}
                      className="w-full btn-mechanical-error flex items-center justify-center space-x-2"
                    >
                      <XCircle size={14} />
                      <span>{t('components.findingDetailModal.actions.falsePositive')}</span>
                    </button>
                  </div>
                </section>

                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-2">{t('components.findingDetailModal.sections.metadata')}</h3>
                  <div className="space-y-3 p-3 border border-outline-variant bg-surface-container-low">
                    <div className="flex justify-between items-center text-[10px]">
                      <span className="text-on-surface-variant opacity-60 uppercase">{t('components.findingDetailModal.metadata.stack')}:</span>
                      <span className="px-1.5 py-0.5 bg-primary/10 border border-primary/30 text-primary font-black uppercase">
                        {finding.stack || t('components.findingDetailModal.metadata.agnostic')}
                      </span>
                    </div>
                    <div className="flex justify-between items-center text-[10px]">
                      <span className="text-on-surface-variant opacity-60 uppercase">{t('components.findingDetailModal.metadata.cwe')}:</span>
                      <span className="text-on-surface font-mono">{finding.cwe_id || 'N/A'}</span>
                    </div>
                    <div className="flex justify-between items-center text-[10px]">
                      <span className="text-on-surface-variant opacity-60 uppercase">{t('components.findingDetailModal.metadata.owasp')}:</span>
                      <span className="text-on-surface font-mono">{finding.owasp || 'N/A'}</span>
                    </div>
                  </div>
                </section>

                <section>
                  <h3 className="text-label-xs text-on-surface-variant mb-2">{t('components.findingDetailModal.sections.assignment')}</h3>
                  <div className="flex items-center space-x-3 p-3 border border-outline-variant bg-surface-container-low">
                    <div className="w-8 h-8 bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
                      <User size={16} />
                    </div>
                    <div className="flex-1">
                      <div className="text-[11px] font-bold text-on-surface uppercase tracking-tight">
                        {finding.assigned_to
                          ? `User_ID_${finding.assigned_to}`
                          : t('components.findingDetailModal.assignment.autoAssigned')}
                      </div>
                      <div className="text-[9px] text-on-surface-variant opacity-60 italic">
                        {t('components.findingDetailModal.assignment.secOps')}
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            </div>
          )}

          {activeTab === 'code' && (
            <div className="h-full flex flex-col space-y-4">
              <div className="flex-1 bg-surface-container-lowest border border-outline-variant skeuo-inset overflow-auto p-4 font-mono text-[13px] relative group">
                <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-none">
                  <div className="px-2 py-1 bg-surface-container-high border border-outline-variant text-[10px] text-on-surface-variant">
                    {finding.file_path}
                  </div>
                </div>
                <pre className="text-on-surface">
                  {fileContent.split('\n').map((line, i) => {
                    const isTargetLine = i + 1 === finding.line_number;
                    return (
                      <div
                        key={i}
                        className={`flex ${isTargetLine ? 'bg-error/20 border-l-2 border-error -ml-4 pl-4' : ''}`}
                      >
                        <span className="w-12 text-on-surface-variant select-none opacity-30 inline-block text-right pr-4">
                          {i + 1}
                        </span>
                        <span>{line}</span>
                      </div>
                    );
                  })}
                </pre>
              </div>
            </div>
          )}

          {activeTab === 'ai' && (
            <div className="h-full flex flex-col space-y-6">
              {!aiAnalysis && !isAnalyzing ? (
                <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-outline-variant space-y-4 text-center p-12">
                  <div className="w-16 h-16 bg-primary-fixed-dim/10 flex items-center justify-center text-primary-fixed-dim">
                    <Cpu size={32} />
                  </div>
                  <div>
                    <h4 className="text-header-md text-on-surface mb-2 tracking-widest uppercase">
                      {t('components.findingDetailModal.ai.requestRemediation')}
                    </h4>
                    <p className="text-body-base text-on-surface-variant max-w-md mx-auto">
                      {t('components.findingDetailModal.ai.description')}
                    </p>
                  </div>
                  <button
                    onClick={handleRunAI}
                    className="btn-mechanical-primary flex items-center space-x-2 px-8"
                  >
                    <MessageSquare size={16} />
                    <span>{t('components.findingDetailModal.ai.generateFixPlan')}</span>
                  </button>
                </div>
              ) : (
                <div className="flex-1 overflow-auto bg-surface-container-lowest border border-outline-variant p-6">
                  {isAnalyzing ? (
                    <div className="flex flex-col items-center justify-center h-full gap-6">
                      <div className="flex gap-1.5">
                        {[0, 1, 2].map((i) => (
                          <div
                            key={i}
                            className="w-2 h-5 bg-ai-accent animate-pulse"
                            style={{ animationDelay: `${i * 0.15}s` }}
                          />
                        ))}
                      </div>
                      <span className="text-label-caps font-label-caps text-ai-accent tracking-[0.3em]">
                        {t('components.findingDetailModal.ai.analyzing')}
                      </span>
                    </div>
                  ) : (
                    <div className="prose prose-invert prose-sm max-w-none prose-pre:bg-surface-container prose-pre:border prose-pre:border-outline-variant">
                      <ReactMarkdown>{aiAnalysis}</ReactMarkdown>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer actions for overall modal */}
        <div className="pt-4 border-t border-outline-variant flex justify-between items-center">
          <div className="text-[10px] text-on-surface-variant font-mono uppercase tracking-widest">
            {t('components.findingDetailModal.footer.entityId')}: {finding.id}
          </div>
          <button onClick={onClose} className="btn-mechanical">
            {t('components.findingDetailModal.footer.close')}
          </button>
        </div>
      </div>
    </Modal>
  );
};

export default FindingDetailModal;
