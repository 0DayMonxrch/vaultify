import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { KeyRound, AlertCircle, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { ModeToggle } from '@/components/ModeToggle';

import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Form, FormField, FormItem, FormLabel, FormControl, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useAuth } from '../useAuth';

const loginSchema = z.object({
  email: z.string().email({ message: 'Invalid email address' }),
  password: z.string().min(1, { message: 'Password is required' }),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export const LoginForm: React.FC = () => {
  const { login, demoLogin } = useAuth();
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);
  const [isDemoLoading, setIsDemoLoading] = useState(false);

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const { isSubmitting } = form.formState;

  const onSubmit = async (data: LoginFormValues) => {
    try {
      setServerError(null);
      await login(data);
      navigate('/projects');
    } catch (err: any) {
      setServerError(err?.response?.data?.message || 'Invalid credentials. Please try again.');
    }
  };

  const onDemoClick = async () => {
    try {
      setIsDemoLoading(true);
      setServerError(null);
      await demoLogin();
      navigate('/projects');
    } catch (err: any) {
      const errorData = err?.response?.data;
      if (typeof errorData === 'string' && errorData.trim() !== '') {
        setServerError(errorData);
      } else {
        setServerError('Demo login failed or rate-limited. Please try again later.');
      }
    } finally {
      setIsDemoLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4 relative">
      <div className="absolute top-4 right-4 md:top-8 md:right-8">
        <ModeToggle />
      </div>
      <Card className="w-full max-w-sm border-border shadow-sm">
        <CardHeader className="space-y-2 text-center">
          <div className="mx-auto flex h-9 w-9 items-center justify-center rounded-md bg-zinc-900 dark:bg-zinc-50">
            <KeyRound className="h-4 w-4 text-zinc-50 dark:text-zinc-900" />
          </div>
          <CardTitle className="text-xl font-semibold">Sign in to Vaultify</CardTitle>
          <CardDescription className="text-sm text-muted-foreground">
            Access your team's encrypted secrets.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {serverError && (
            <Alert variant="destructive" className="mb-4">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Sign in failed</AlertTitle>
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
                  <div className="flex items-center justify-between">
                    <FormLabel className="text-sm">Password</FormLabel>
                  </div>
                  <FormControl>
                    <Input type="password" autoComplete="current-password" {...field} />
                  </FormControl>
                  <FormMessage className="text-xs" />
                </FormItem>
              )} />

              <Button type="submit" className="w-full" disabled={isSubmitting || isDemoLoading}>
                {isSubmitting ? (
                  <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Signing in...</>
                ) : "Sign in"}
              </Button>
            </form>
          </Form>

          <div className="relative mt-4 mb-4">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-border" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">Or</span>
            </div>
          </div>

          <Button 
            variant="outline" 
            className="w-full" 
            onClick={onDemoClick}
            disabled={isSubmitting || isDemoLoading}
          >
            {isDemoLoading ? (
              <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Provisioning Demo...</>
            ) : "Try the Demo"}
          </Button>
        </CardContent>

        <CardFooter className="justify-center">
          <p className="text-sm text-muted-foreground">
            Don't have an account?{" "}
            <a href="/register" className="font-medium text-foreground underline-offset-4 hover:underline" onClick={(e) => { e.preventDefault(); navigate('/register'); }}>
              Create one
            </a>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
};
