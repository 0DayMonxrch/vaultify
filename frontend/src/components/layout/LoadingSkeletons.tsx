import React from 'react';
import { cn } from '../../lib/utils';

export const Skeleton: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({ className, ...props }) => {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-muted/60", className)}
      {...props}
    />
  );
}

export const ProjectCardSkeleton: React.FC = () => {
  return (
    <div className="border border-border rounded-xl p-6 flex flex-col space-y-4 shadow-sm bg-card">
      <Skeleton className="h-6 w-1/2" />
      <Skeleton className="h-4 w-1/3" />
      <div className="mt-8 pt-4 border-t border-border/50">
        <Skeleton className="h-3 w-1/4" />
      </div>
    </div>
  );
};
