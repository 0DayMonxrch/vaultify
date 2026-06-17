import React from 'react';
import { cn } from '../../../lib/utils';

export const ActionBadge: React.FC<{ action: string }> = ({ action }) => {
  let colorClass = 'bg-muted text-muted-foreground';

  if (action.includes('WRITE') || action.includes('CREATE')) {
    colorClass = 'bg-emerald-500/15 text-emerald-600 border border-emerald-500/20';
  } else if (action.includes('READ') || action.includes('VIEW')) {
    colorClass = 'bg-blue-500/15 text-blue-600 border border-blue-500/20';
  } else if (action.includes('DELETE') || action.includes('FAILED')) {
    colorClass = 'bg-destructive/15 text-destructive border border-destructive/20';
  }

  return (
    <span className={cn('text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded-md shadow-sm', colorClass)}>
      {action}
    </span>
  );
};
