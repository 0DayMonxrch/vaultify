import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useProjects } from '../hooks/useProjects';
import { cn } from '../../../lib/utils';
import { CreateProjectDialog } from './CreateProjectDialog';

export const ProjectSidebar: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const { data: projects, isLoading } = useProjects();

  return (
    <aside className="w-[280px] shrink-0 border-r border-border bg-muted/10 h-[calc(100vh-65px)] flex flex-col">
      <div className="p-4 border-b border-border/50">
        <h2 className="text-[11px] font-bold text-muted-foreground uppercase tracking-widest mb-4">Your Projects</h2>
        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3].map(i => <div key={i} className="h-9 bg-muted/60 rounded-md animate-pulse" />)}
          </div>
        ) : (
          <div className="space-y-1">
            {projects?.map(project => {
              const isActive = project.id === projectId;
              return (
                <Link
                  key={project.id}
                  to={`/projects/${project.id}`}
                  className={cn(
                    "flex items-center justify-between px-3 py-2 rounded-md transition-all text-sm font-medium",
                    isActive ? "bg-primary/10 text-primary shadow-sm ring-1 ring-primary/20" : "text-foreground hover:bg-muted/50"
                  )}
                >
                  <span className="truncate">{project.name}</span>
                  {isActive && (
                    <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-primary/20 text-primary ml-2 shrink-0">
                      {(project as any).role || 'MEMBER'}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>
        )}
      </div>
      <div className="mt-auto p-4 border-t border-border/50 bg-background/50">
        <CreateProjectDialog />
      </div>
    </aside>
  );
};
