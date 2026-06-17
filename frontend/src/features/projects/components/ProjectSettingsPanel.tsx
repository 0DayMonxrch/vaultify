import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useProject } from '../hooks/useProject';
import { useUpdateProject, useDeleteProject } from '../hooks/useProjectMutations';
import { usePermission } from '../../../hooks/usePermission';
import { Button } from '../../../components/ui/button';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';

const schema = z.object({
  name: z.string().min(1),
  slug: z.string().min(1).regex(/^[a-z0-9-]+$/),
});

export const ProjectSettingsPanel: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();
  const { canDeleteProject, canUpdateProject } = usePermission();
  const { data: project, isLoading } = useProject(projectId!);
  const { mutateAsync: updateProject, isPending: isUpdating } = useUpdateProject(projectId!);
  const { mutateAsync: deleteProject, isPending: isDeleting } = useDeleteProject();

  const { register, handleSubmit } = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    values: {
      name: project?.name || '',
      slug: project?.slug || '',
    }
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      await updateProject(data);
    } catch(e) {
      console.error(e);
    }
  };

  const handleDelete = () => {
    toast('Are you absolutely sure?', {
      description: 'This action cannot be undone. This will permanently delete the project and all associated secrets.',
      action: {
        label: 'Delete Project',
        onClick: async () => {
          try {
            await deleteProject(projectId!);
            navigate('/projects');
            toast.success('Project deleted successfully');
          } catch(e) {
            console.error(e);
            toast.error('Failed to delete project');
          }
        }
      },
      cancel: {
        label: 'Cancel',
        onClick: () => {}
      }
    });
  };

  if (!canUpdateProject) {
    return (
      <div className="p-8 text-center bg-destructive/10 border border-destructive/20 rounded-xl text-destructive font-medium shadow-sm">
        You do not have permission to view project settings. Only Project Owners can modify configurations.
      </div>
    );
  }

  if (isLoading) return <div className="animate-pulse h-64 bg-muted/20 rounded-xl"></div>;

  return (
    <div className="space-y-8 max-w-3xl">
      <div className="bg-card p-8 rounded-xl border border-border/50 shadow-sm">
        <h2 className="text-xl font-bold mb-6 text-foreground tracking-tight">General Settings</h2>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <div>
            <label className="block text-sm font-medium mb-1.5">Project Name</label>
            <input {...register('name')} className="w-full px-3 py-2 border rounded-lg bg-background focus:ring-2 focus:ring-primary/50" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1.5">URL Slug</label>
            <input {...register('slug')} className="w-full px-3 py-2 border rounded-lg bg-background focus:ring-2 focus:ring-primary/50" />
            <p className="text-xs text-muted-foreground mt-2">Used by the Vaultify CLI to reference this project natively.</p>
          </div>
          <div className="pt-2">
            <Button type="submit" disabled={isUpdating}>{isUpdating ? 'Saving...' : 'Save Changes'}</Button>
          </div>
        </form>
      </div>

      {canDeleteProject && (
        <div className="bg-destructive/5 p-8 rounded-xl border border-destructive/20 shadow-sm">
          <h2 className="text-lg font-bold text-destructive mb-2">Danger Zone</h2>
          <p className="text-sm text-destructive/80 mb-6 font-medium">Deleting this project will permanently remove all associated secrets, environments, and audit logs. This cannot be recovered.</p>
          <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
            {isDeleting ? 'Deleting...' : 'Delete Project'}
          </Button>
        </div>
      )}
    </div>
  );
};
