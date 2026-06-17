import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useUpdateSecret } from '../hooks/useUpdateSecret';
import type { Secret } from '@/api/endpoints/secrets.api';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
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
  value: z.string().min(1, 'Value is required'),
});

type FormValues = z.infer<typeof formSchema>;

interface EditSecretDialogProps {
  projectId: string;
  secret: Secret;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const EditSecretDialog: React.FC<EditSecretDialogProps> = ({ projectId, secret, open, onOpenChange }) => {
  const { mutate: updateSecret, isPending } = useUpdateSecret();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      value: '',
    },
  });

  // Reset form when dialog opens/closes to clear state
  React.useEffect(() => {
    if (!open) {
      form.reset({
        value: '',
      });
    }
  }, [open, form]);

  const onSubmit = (values: FormValues) => {
    updateSecret({ projectId, secretId: secret.id, value: values.value }, {
      onSuccess: () => {
        onOpenChange(false);
      }
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Update Secret ({secret.key_name})</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>New Value</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="s3cr3t_v4lu3" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="flex justify-end pt-4">
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Updating...' : 'Update Secret'}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};
