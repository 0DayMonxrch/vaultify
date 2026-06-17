import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { UserPlus, AlertCircle, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';

import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Form, FormField, FormItem, FormLabel, FormControl, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { register as registerApi } from '../../../api/endpoints/auth.api';
import { ModeToggle } from '@/components/ModeToggle';
import { toast } from 'sonner';

const registerSchema = z.object({
  email: z.string().email({ message: 'Invalid email address' }),
  password: z.string().min(8, { message: 'Password must be at least 8 characters' }),
  confirmPassword: z.string().min(1, { message: 'Please confirm your password' }),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export const RegisterForm: React.FC = () => {
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);

  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: '',
    },
  });

  const { mutateAsync: registerMutation, isPending } = useMutation({
    mutationFn: registerApi,
    onSuccess: () => {
      toast.success("Registration successful! Please sign in.");
      navigate('/login');
    },
    onError: (err: any) => {
      setServerError(err?.response?.data?.message || 'Registration failed. Please try again or use a different email.');
    }
  });

  const onSubmit = async (data: RegisterFormValues) => {
    setServerError(null);
    await registerMutation({ email: data.email, password: data.password });
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 relative">
      <div className="absolute top-4 right-4 md:top-8 md:right-8">
        <ModeToggle />
      </div>
      <Card className="w-full max-w-sm border-border shadow-sm">
        <CardHeader className="space-y-2 text-center">
          <div className="mx-auto flex h-9 w-9 items-center justify-center rounded-md bg-zinc-900 dark:bg-zinc-50">
            <UserPlus className="h-4 w-4 text-zinc-50 dark:text-zinc-900" />
          </div>
          <CardTitle className="text-xl font-semibold">Create an account</CardTitle>
          <CardDescription className="text-sm text-muted-foreground">
            Sign up to securely manage your secrets.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {serverError && (
            <Alert variant="destructive" className="mb-4">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Registration failed</AlertTitle>
              <AlertDescription className="text-sm">{serverError}</AlertDescription>
            </Alert>
          )}

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField name="email" control={form.control} render={({ field }) => (
                <FormItem>
                  <FormLabel className="text-sm">Email</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      placeholder="you@company.com"
                      autoComplete="email"
                      className="font-sans"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className="text-xs" />
                </FormItem>
              )} />

              <FormField name="password" control={form.control} render={({ field }) => (
                <FormItem>
                  <FormLabel className="text-sm">Password</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="new-password" {...field} />
                  </FormControl>
                  <FormMessage className="text-xs" />
                </FormItem>
              )} />

              <FormField name="confirmPassword" control={form.control} render={({ field }) => (
                <FormItem>
                  <FormLabel className="text-sm">Confirm Password</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="new-password" {...field} />
                  </FormControl>
                  <FormMessage className="text-xs" />
                </FormItem>
              )} />

              <Button type="submit" className="w-full" disabled={isPending}>
                {isPending ? (
                  <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Creating account...</>
                ) : "Sign up"}
              </Button>
            </form>
          </Form>
        </CardContent>

        <CardFooter className="justify-center">
          <p className="text-sm text-muted-foreground">
            Already have an account?{" "}
            <a href="/login" className="font-medium text-foreground underline-offset-4 hover:underline" onClick={(e) => { e.preventDefault(); navigate('/login'); }}>
              Sign in
            </a>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
};
