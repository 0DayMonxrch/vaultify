import React from 'react';
import { Link } from 'react-router-dom';
import type { Project } from '../../../api/endpoints/projects.api';

interface ProjectCardProps {
  project: Project;
}

export const ProjectCard: React.FC<ProjectCardProps> = ({ project }) => {
  return (
    <Link to={`/projects/${project.id}`} className="block group h-full">
      <div className="border border-border rounded-xl p-6 transition-all duration-200 shadow-sm hover:shadow-md hover:border-primary/40 bg-card h-full flex flex-col relative overflow-hidden">
        <div className="absolute top-0 left-0 w-1 h-full bg-primary/0 group-hover:bg-primary transition-colors"></div>
        <h3 className="font-semibold text-lg text-foreground group-hover:text-primary transition-colors">{project.name}</h3>
        <p className="text-sm font-medium text-muted-foreground mt-1">/{project.slug}</p>
        
        <div className="mt-auto pt-5 border-t border-border/40 mt-6 flex items-center justify-between text-xs text-muted-foreground">
          <span>Created {new Date(project.created_at).toLocaleDateString()}</span>
          <span className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center text-primary font-medium">
            Open Vault &rarr;
          </span>
        </div>
      </div>
    </Link>
  );
};
