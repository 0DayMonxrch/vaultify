import React from 'react';
import { useTokens } from '../hooks/useTokens';
import { RevokeTokenConfirm } from './RevokeTokenConfirm';
import { CreateTokenDialog } from './CreateTokenDialog';

export const TokensTable: React.FC = () => {
  const { data: tokens, isLoading, isError } = useTokens();

  const activeTokens = tokens?.filter(t => !t.revoked);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-card p-4 rounded-xl border border-border/50 shadow-sm">
        <div>
          <h2 className="text-lg font-bold text-foreground">Active Tokens</h2>
          <p className="text-sm text-muted-foreground">Manage machine identities mapped to your account.</p>
        </div>
        <CreateTokenDialog />
      </div>

      <div className="border border-border rounded-xl overflow-hidden bg-card shadow-sm">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-muted/30 border-b border-border/50">
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Name</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Prefix</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Scope</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Created</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Last Used</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={6} className="p-10 text-center">
                  <div className="animate-pulse space-y-4 max-w-lg mx-auto">
                    <div className="h-10 bg-muted/60 rounded-md w-full"></div>
                    <div className="h-10 bg-muted/60 rounded-md w-full"></div>
                  </div>
                </td>
              </tr>
            ) : isError ? (
              <tr><td colSpan={6} className="p-10 text-center text-destructive font-medium bg-destructive/5">Failed to load API tokens.</td></tr>
            ) : !activeTokens || activeTokens.length === 0 ? (
              <tr>
                <td colSpan={6} className="p-16 text-center bg-muted/5">
                  <div className="w-12 h-12 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg className="w-6 h-6 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                  </div>
                  <h3 className="text-lg font-bold text-foreground">No tokens generated</h3>
                  <div className="text-muted-foreground mt-1 max-w-sm mx-auto text-sm">Create an API token to programmatically manage your secrets via CLI or CI/CD pipelines.</div>
                </td>
              </tr>
            ) : (
              activeTokens.map(token => (
                <tr key={token.id} className="border-b border-border/50 hover:bg-muted/10 transition-colors">
                  <td className="p-4 font-semibold text-sm text-foreground">{token.name}</td>
                  <td className="p-4 font-mono text-xs text-muted-foreground tracking-wider">{token.token_prefix}••••••••</td>
                  <td className="p-4">
                    <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded bg-secondary text-secondary-foreground">
                      {token.role}
                    </span>
                  </td>
                  <td className="p-4 text-xs text-muted-foreground font-medium">
                    {new Date(token.created_at).toLocaleDateString()}
                  </td>
                  <td className="p-4 text-xs text-muted-foreground font-medium">
                    {token.last_used_at ? new Date(token.last_used_at).toLocaleDateString() : 'Never'}
                  </td>
                  <td className="p-4 text-right">
                    <RevokeTokenConfirm tokenId={token.id} tokenName={token.name} />
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
