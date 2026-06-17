import React from 'react';
import { TokensTable } from './components/TokensTable';

export const ApiTokensPage: React.FC = () => {
  return (
    <div className="max-w-6xl mx-auto p-6 md:p-10 space-y-8 animate-in fade-in duration-500">
      <div>
        <h1 className="text-3xl font-extrabold tracking-tight text-foreground">Account Tokens</h1>
        <p className="text-muted-foreground font-medium mt-1">Configure service accounts for CLI and CI/CD pipelines.</p>
      </div>
      <TokensTable />
    </div>
  );
};
