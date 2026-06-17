import React from 'react';
import { useProjects } from './hooks/useProjects';
import { ProjectCardSkeleton } from '../../components/layout/LoadingSkeletons';
import { ProjectCard } from './components/ProjectCard';
import { CreateProjectDialog } from './components/CreateProjectDialog';

export const ProjectsListPage: React.FC = () => {
  const { data: projects, isLoading, isError } = useProjects();

  return (
    <div className="max-w-7xl mx-auto p-6 md:p-10 space-y-8 animate-in fade-in duration-500">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-6 pb-6 border-b border-border/50">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground">Projects</h1>
          <p className="text-muted-foreground font-medium mt-1">Manage your secret vaults and teams securely.</p>
        </div>
        <CreateProjectDialog />
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <ProjectCardSkeleton />
          <ProjectCardSkeleton />
          <ProjectCardSkeleton />
          <ProjectCardSkeleton />
        </div>
      ) : isError ? (
        <div className="text-destructive border border-destructive/20 bg-destructive/10 p-6 rounded-xl font-medium">
          Failed to load projects. Please try refreshing the page.
        </div>
      ) : !projects || projects.length === 0 ? (
        <div className="border-2 border-dashed border-border rounded-2xl p-16 text-center bg-muted/10 flex flex-col items-center">
          <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-foreground">No projects yet</h3>
          <p className="text-muted-foreground font-medium mt-2 max-w-sm mb-6">Create your first project to start securely managing your environment variables and API keys.</p>
          <CreateProjectDialog />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {projects.map(project => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  );
};
