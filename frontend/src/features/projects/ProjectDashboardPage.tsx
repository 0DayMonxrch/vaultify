import React from 'react';
import { useSearchParams, useParams } from 'react-router-dom';
import { SecretsTab } from '../secrets/SecretsTab';
import { MembersTab } from './tabs/MembersTab';
import { AuditTab } from '../audit/AuditTab';
import { SettingsTab } from './tabs/SettingsTab';

export const ProjectDashboardPage: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const [searchParams] = useSearchParams();
  const tab = searchParams.get('tab') || 'secrets';

  if (!projectId) return null;

  return (
    <div className="animate-in fade-in duration-300">
      {tab === 'secrets' && <SecretsTab />}
      {tab === 'members' && <MembersTab />}
      {tab === 'audit' && <AuditTab />}
      {tab === 'settings' && <SettingsTab />}
    </div>
  );
};
