import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { toast } from 'sonner';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

const formSchema = z.object({
  keyName: z.string().min(1, 'Key name is required').regex(/^[A-Z0-9_]+$/, 'Must be uppercase, alphanumeric, and underscores only'),
  value: z.string().min(1, 'Value is required'),
  environment: z.string(),
});

type FormValues = z.infer<typeof formSchema>;

interface CreateSecretDialogProps {
  projectId: string;
  environment: string;
  children: React.ReactNode;
}

export const CreateSecretDialog: React.FC<CreateSecretDialogProps> = ({ projectId, environment, children }) => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      keyName: '',
      value: '',
      environment: environment,
    },
  });

  // Update form environment if prop changes
  React.useEffect(() => {
    form.setValue('environment', environment);
  }, [environment, form]);

  // Reset form when dialog opens/closes to clear state
  React.useEffect(() => {
    if (!open) {
      form.reset({
        keyName: '',
        value: '',
        environment: environment,
      });
    }
  }, [open, environment, form]);

  const mutation = useMutation({
    mutationFn: async (values: FormValues) => {
      const response = await apiClient.post(`/projects/${projectId}/secrets`, {
        key_name: values.keyName,
        value: values.value,
        environment: values.environment,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'secrets'] });
      toast.success('Secret created successfully');
      setOpen(false);
      form.reset();
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create secret');
    },
  });

  const onSubmit = (values: FormValues) => {
    mutation.mutate(values);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Add Secret</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="keyName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Key Name</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="DATABASE_URL" 
                      {...field} 
                      onChange={(e) => field.onChange(e.target.value.toUpperCase())} 
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Value</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="s3cr3t_v4lu3" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="flex justify-end pt-4">
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? 'Saving...' : 'Save Secret'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};
