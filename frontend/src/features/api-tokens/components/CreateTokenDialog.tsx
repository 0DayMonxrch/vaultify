import React, { useState } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useCreateToken } from '../hooks/useTokens';
import { useProjects } from '../../projects/hooks/useProjects';
import { useClipboard } from '../../../hooks/useClipboard';
import { Button } from '../../../components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  role: z.enum(['read', 'write']),
  project_id: z.string().min(1, 'Project is required')
});

export const CreateTokenDialog: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  // CRITICAL: Raw token strictly in component state.
  const [rawToken, setRawToken] = useState<string | null>(null);
  const [timeLeft, setTimeLeft] = useState(30);
  const { mutateAsync: createToken, isPending } = useCreateToken();
  const { data: projects } = useProjects();
  const { copy, copied } = useClipboard();

  const { register, handleSubmit, formState: { errors }, reset, control } = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { role: 'read', project_id: '' }
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      const response = await createToken(data);
      setRawToken(response.token);
      setTimeLeft(30);
      reset();
    } catch (e) {
      console.error(e);
    }
  };

  React.useEffect(() => {
    if (!rawToken) return;

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          setRawToken(null);
          setIsOpen(false);
          clearInterval(timer);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [rawToken]);

  const closeDialog = () => {
    setIsOpen(false);
    // CRITICAL: Annihilate raw token state
    setRawToken(null);
    reset();
  };

  return (
    <>
      <Button onClick={() => setIsOpen(true)}>Generate Token</Button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div className="bg-card border border-border rounded-xl p-8 w-full max-w-md shadow-2xl animate-in zoom-in-95 duration-200">
            {rawToken ? (
              <div className="space-y-6">
                <div className="flex items-center space-x-3 text-amber-500 border-b border-border pb-4">
                  <svg className="w-6 h-6 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                  <h2 className="text-xl font-bold tracking-tight">Save your token now</h2>
                  <span className="ml-auto bg-amber-500/10 text-amber-600 px-2.5 py-0.5 rounded-full text-sm font-bold animate-pulse">
                    {timeLeft}s
                  </span>
                </div>
                <div className="bg-muted/50 p-5 rounded-xl border border-border shadow-inner">
                  <p className="text-sm font-medium mb-4 text-muted-foreground leading-relaxed">
                    This token will <strong className="text-foreground">never be shown again</strong>. Copy it and store it in a secure location immediately.
                  </p>
                  <div className="flex items-center space-x-3 bg-background p-2 rounded-lg border border-border">
                    <code className="flex-1 block px-2 text-sm text-foreground break-all font-mono">
                      {rawToken}
                    </code>
                    <Button variant="secondary" onClick={() => copy(rawToken)} className="shrink-0">
                      {copied ? 'Copied!' : 'Copy'}
                    </Button>
                  </div>
                </div>
                <div className="pt-2 flex justify-end">
                  <Button onClick={closeDialog}>I have saved it</Button>
                </div>
              </div>
            ) : (
              <>
                <h2 className="text-xl font-bold mb-6 text-foreground tracking-tight">Generate API Token</h2>
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                  <div>
                    <label className="block text-sm font-medium mb-1.5 text-foreground">Token Name</label>
                    <input
                      {...register('name')}
                      className="w-full px-3 py-2.5 border rounded-lg bg-background focus:ring-2 focus:ring-primary/50 transition-all"
                      placeholder="e.g. GitHub Actions CI"
                    />
                    {errors.name && <p className="text-destructive text-xs mt-1.5 font-medium">{errors.name.message}</p>}
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1.5 text-foreground">Project</label>
                    <Controller
                      name="project_id"
                      control={control}
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                          <SelectTrigger className="w-full">
                            <SelectValue placeholder="Select project" />
                          </SelectTrigger>
                          <SelectContent>
                            {projects?.map(p => (
                              <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    />
                    {errors.project_id && <p className="text-destructive text-xs mt-1.5 font-medium">{errors.project_id.message}</p>}
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1.5 text-foreground">Scope</label>
                    <Controller
                      name="role"
                      control={control}
                      render={({ field }) => (
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                          <SelectTrigger className="w-full">
                            <SelectValue placeholder="Select scope" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="read">Read Only (Secrets Decryption)</SelectItem>
                            <SelectItem value="write">Write (Secrets Management)</SelectItem>
                          </SelectContent>
                        </Select>
                      )}
                    />
                  </div>
                  <div className="flex justify-end space-x-3 pt-6 border-t border-border mt-6">
                    <Button variant="ghost" type="button" onClick={closeDialog}>Cancel</Button>
                    <Button type="submit" disabled={isPending}>{isPending ? 'Generating...' : 'Generate Token'}</Button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
};
