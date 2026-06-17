import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useCreateProject } from '../hooks/useProjectMutations';
import { Button } from '../../../components/ui/button';
import { useNavigate } from 'react-router-dom';

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().min(1, 'Slug is required').regex(/^[a-z0-9-]+$/, 'Lowercase alphanumeric and dashes only'),
});

type FormData = z.infer<typeof schema>;

export const CreateProjectDialog: React.FC<{ trigger?: React.ReactNode }> = ({ trigger }) => {
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();
  const { mutateAsync: createProject, isPending } = useCreateProject();
  
  const { register, handleSubmit, formState: { errors }, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: FormData) => {
    try {
      const res = await createProject(data);
      setIsOpen(false);
      reset();
      if (res && res.id) {
        navigate(`/projects/${res.id}`);
      }
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <>
      {trigger ? (
        <div onClick={() => setIsOpen(true)} className="cursor-pointer">{trigger}</div>
      ) : (
        <Button onClick={() => setIsOpen(true)}>New Project</Button>
      )}

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div className="bg-card border border-border rounded-xl p-8 w-full max-w-md shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <h2 className="text-2xl font-bold mb-6 text-foreground tracking-tight">Create New Project</h2>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground">Project Name</label>
                <input
                  {...register('name')}
                  className="w-full px-3 py-2.5 border border-input rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                  placeholder="e.g. Production Vault"
                />
                {errors.name && <p className="text-destructive text-xs font-medium mt-1.5">{errors.name.message}</p>}
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground">URL Slug</label>
                <input
                  {...register('slug')}
                  className="w-full px-3 py-2.5 border border-input rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                  placeholder="e.g. prod-vault"
                />
                {errors.slug && <p className="text-destructive text-xs font-medium mt-1.5">{errors.slug.message}</p>}
              </div>

              <div className="flex justify-end space-x-3 pt-4 border-t border-border">
                <Button variant="ghost" type="button" onClick={() => setIsOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={isPending}>
                  {isPending ? 'Creating...' : 'Create Project'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
};
