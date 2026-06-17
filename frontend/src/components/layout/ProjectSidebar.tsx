import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { KeyRound, Plus, Folder } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useProjects } from '../../features/projects/hooks/useProjects';
import { CreateProjectDialog } from '../../features/projects/components/CreateProjectDialog';
import { cn } from '@/lib/utils';

interface ProjectSidebarProps {
  className?: string;
}

export const ProjectSidebar: React.FC<ProjectSidebarProps> = ({ className }) => {
  const { projectId } = useParams<{ projectId: string }>();
  const { data: projects = [] } = useProjects();

  return (
    <aside className={cn("flex w-[260px] flex-shrink-0 flex-col border-r border-border bg-zinc-50/50 dark:bg-zinc-900/30", className)}>
      <Link to="/projects" className="flex h-14 items-center gap-2 border-b border-border px-4 hover:bg-zinc-100/50 dark:hover:bg-zinc-800/50 transition-colors">
        <KeyRound className="h-4 w-4" />
        <span className="text-sm font-semibold">Vaultify</span>
      </Link>

      <div className="px-3 pt-3">
        <CreateProjectDialog trigger={
          <Button variant="outline" className="w-full justify-start gap-2 text-sm">
            <Plus className="h-3.5 w-3.5" /> New Project
          </Button>
        } />
      </div>

      <div className="flex-1 overflow-y-auto px-3 py-3">
        <p className="px-2 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Projects
        </p>
        <nav className="flex flex-col gap-0.5">
          {projects.map((project: any) => {
            const isActive = project.id === projectId;
            return (
              <Link
                key={project.id}
                to={`/projects/${project.id}`}
                className={cn(
                  "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
                  isActive
                    ? "bg-zinc-200/70 font-medium text-foreground dark:bg-zinc-800"
                    : "text-muted-foreground hover:bg-zinc-100 hover:text-foreground dark:hover:bg-zinc-800/50"
                )}
              >
                <Folder className="h-3.5 w-3.5 flex-shrink-0" />
                <span className="truncate">{project.name}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </aside>
  );
};
