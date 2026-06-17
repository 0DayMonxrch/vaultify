import React from 'react';
import { useSearchParams, useParams } from 'react-router-dom';
import { useAuditLog } from '../hooks/useAuditLog';
import { ActionBadge } from './ActionBadge';
import { Button } from '../../../components/ui/button';

export const AuditTable: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const page = parseInt(searchParams.get('page') || '1', 10);

  const { data, isLoading, isError } = useAuditLog(projectId!, page);

  const handlePrevious = () => {
    if (page > 1) {
      searchParams.set('page', (page - 1).toString());
      setSearchParams(searchParams);
    }
  };

  const handleNext = () => {
    if (data && page < data.total_pages) {
      searchParams.set('page', (page + 1).toString());
      setSearchParams(searchParams);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-card p-4 rounded-xl border border-border/50 shadow-sm">
        <div>
          <h2 className="text-lg font-bold text-foreground">Audit Log</h2>
          <p className="text-sm text-muted-foreground">Immutable trail of all cryptographic and administrative actions.</p>
        </div>
      </div>

      <div className="border border-border rounded-xl overflow-hidden bg-card shadow-sm flex flex-col">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-muted/30 border-b border-border/50">
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Timestamp</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">User / Identity</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Action</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Target</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && !data ? (
              <tr>
                <td colSpan={4} className="p-10 text-center">
                  <div className="animate-pulse space-y-4 max-w-lg mx-auto">
                    <div className="h-10 bg-muted/60 rounded-md w-full"></div>
                    <div className="h-10 bg-muted/60 rounded-md w-full"></div>
                  </div>
                </td>
              </tr>
            ) : isError ? (
              <tr><td colSpan={4} className="p-10 text-center text-destructive font-medium bg-destructive/5">Failed to load audit logs.</td></tr>
            ) : !data || !data.data || data.data.length === 0 ? (
              <tr>
                <td colSpan={4} className="p-16 text-center bg-muted/5 text-muted-foreground font-medium">
                  No audit events recorded yet.
                </td>
              </tr>
            ) : (
              data.data.map(event => (
                <tr key={event.id} className="border-b border-border/50 hover:bg-muted/10 transition-colors">
                  <td className="p-4 text-xs text-muted-foreground font-medium tracking-tight">
                    {new Date(event.created_at).toLocaleString()}
                  </td>
                  <td className="p-4 font-semibold text-sm text-foreground">{event.user_email}</td>
                  <td className="p-4"><ActionBadge action={event.action} /></td>
                  <td className="p-4 font-mono text-xs text-muted-foreground">{event.target_key_name || '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        
        {/* URL-driven Pagination Bar */}
        <div className="p-4 border-t border-border/50 bg-muted/30 flex items-center justify-between">
          <span className="text-xs text-muted-foreground font-semibold">
            Page {page} {data ? `of ${Math.max(1, data.total_pages)}` : ''}
          </span>
          <div className="flex space-x-3">
            <Button variant="outline" size="sm" onClick={handlePrevious} disabled={page <= 1} className="h-8">
              Previous
            </Button>
            <Button variant="outline" size="sm" onClick={handleNext} disabled={!data || page >= data.total_pages} className="h-8">
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
