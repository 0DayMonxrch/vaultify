import React from 'react';
import { MembersPanel } from '../components/MembersPanel';

export const MembersTab: React.FC = () => {
  return (
    <div className="animate-in fade-in duration-300">
      <MembersPanel />
    </div>
  );
};
