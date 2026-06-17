import React from 'react';
import { AuditTable } from './components/AuditTable';

export const AuditLogPage: React.FC = () => {
  return (
    <div className="animate-in fade-in duration-300">
      <AuditTable />
    </div>
  );
};
