import React from 'react';
import { ProjectSettingsPanel } from '../components/ProjectSettingsPanel';

export const SettingsTab: React.FC = () => {
  return (
    <div className="animate-in fade-in duration-300">
      <ProjectSettingsPanel />
    </div>
  );
};
