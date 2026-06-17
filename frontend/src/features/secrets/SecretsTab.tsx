import React from 'react';
import { SecretsTable } from './components/SecretsTable';

export const SecretsTab: React.FC = () => {
  return (
    <div className="animate-in fade-in duration-300">
      <SecretsTable />
    </div>
  );
};
