import React from 'react';
import { useSearchParams } from 'react-router-dom';
import { cn } from '../../../lib/utils';

const ENVIRONMENTS = ['production', 'staging', 'development', 'preview'];

export const EnvironmentTabs: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const currentEnv = searchParams.get('env') || 'production';

  const setEnv = (env: string) => {
    searchParams.set('env', env);
    setSearchParams(searchParams);
  };

  return (
    <div className="flex space-x-1 bg-muted/40 p-1 rounded-lg border border-border/50 inline-flex shadow-inner">
      {ENVIRONMENTS.map(env => (
        <button
          key={env}
          onClick={() => setEnv(env)}
          className={cn(
            "px-4 py-1.5 text-xs font-bold uppercase tracking-wider rounded-md transition-all",
            currentEnv === env
              ? "bg-background text-foreground shadow-sm ring-1 ring-border"
              : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
          )}
        >
          {env}
        </button>
      ))}
    </div>
  );
};
