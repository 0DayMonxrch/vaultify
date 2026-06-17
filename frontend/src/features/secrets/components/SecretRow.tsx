import React, { useState, useEffect, useRef } from 'react';
import { Key, Eye, Loader2, Check, Copy, EyeOff, RotateCcw, MoreHorizontal, Pencil, Trash2 } from 'lucide-react';
import type { Secret } from '../../../api/endpoints/secrets.api';
import { useRevealSecret } from '../hooks/useRevealSecret';
import { useClipboard } from '../../../hooks/useClipboard';
import { useProject } from '../../projects/hooks/useProject';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { TableRow, TableCell } from '@/components/ui/table';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

const envBadgeClass = (env: string) => {
  if (env === 'production') {
    return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-400';
  }
  if (env === 'staging') {
    return 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-400';
  }
  return 'border-border bg-muted text-muted-foreground';
};

interface SecretRowProps {
  project_id: string;
  secret: Secret;
}

export const SecretRow: React.FC<SecretRowProps> = ({ project_id, secret }) => {
  const [rowState, setRowState] = useState<'masked' | 'revealing' | 'revealed' | 'error'>('masked');
  const [plaintext, setPlaintext] = useState<string | null>(null);
  
  const [secondsLeft, setSecondsLeft] = useState<number>(30);
  const [justCopied, setJustCopied] = useState(false);
  
  const { mutateAsync: fetchSecret } = useRevealSecret();
  const { copy } = useClipboard();
  const { data: project } = useProject(project_id);
  const isOwner = project?.isOwner || false;

  const timerRef = useRef<number | null>(null);

  const clearCountdown = () => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = null;
  };

  const handleHide = () => {
    clearCountdown();
    setPlaintext(null);
    setRowState('masked');
  };

  const startCountdown = () => {
    clearCountdown();
    setSecondsLeft(30);
    timerRef.current = window.setInterval(() => {
      setSecondsLeft((prev) => {
        if (prev <= 1) {
          handleHide();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const handleReveal = async () => {
    setRowState('revealing');
    try {
      const value = await fetchSecret({ projectId: project_id, secretId: secret.id });
      setPlaintext(value);
      setRowState('revealed');
      startCountdown();
    } catch (e) {
      setRowState('error');
      setTimeout(() => {
        setRowState((curr) => curr === 'error' ? 'masked' : curr);
      }, 3000);
    }
  };

  const handleCopy = () => {
    if (plaintext) {
      copy(plaintext);
      setJustCopied(true);
      setTimeout(() => setJustCopied(false), 1500);
    }
  };

  useEffect(() => {
    return () => {
      clearCountdown();
      setPlaintext(null);
    };
  }, []);

  return (
    <TableRow className="group border-b border-border last:border-0 hover:bg-zinc-50 dark:hover:bg-zinc-900/40">
      <TableCell className="py-2.5">
        <div className="flex items-center gap-1.5">
          <Key className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="font-mono text-sm font-medium">{secret.key_name}</span>
        </div>
      </TableCell>

      <TableCell className="py-2.5">
        <Badge variant="outline" className={cn("text-xs font-medium uppercase", envBadgeClass(secret.environment))}>
          {secret.environment}
        </Badge>
      </TableCell>

      <TableCell className="py-2.5">
        {rowState === 'masked' && (
          <div className="flex items-center gap-2">
            <span className="select-none font-mono text-sm text-muted-foreground">••••••••••••</span>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleReveal}>
              <Eye className="h-3.5 w-3.5" />
            </Button>
          </div>
        )}

        {rowState === 'revealing' && (
          <div className="flex items-center gap-2">
            <span className="select-none font-mono text-sm text-muted-foreground">••••••••••••</span>
            <Button variant="ghost" size="icon" className="h-7 w-7" disabled>
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            </Button>
          </div>
        )}

        {rowState === 'revealed' && (
          <div className="flex items-center gap-3">
            <span className="max-w-[220px] truncate font-mono text-sm">{plaintext}</span>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleCopy}>
              {justCopied ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
            </Button>
            <div className="flex items-center gap-1.5 text-xs tabular-nums text-muted-foreground">
              <svg viewBox="0 0 16 16" className="h-4 w-4 -rotate-90">
                <circle cx="8" cy="8" r="6.5" fill="none" strokeWidth="2" className="stroke-muted" />
                <circle
                  cx="8" cy="8" r="6.5" fill="none" strokeWidth="2"
                  strokeDasharray={40.8}
                  strokeDashoffset={40.8 * (1 - secondsLeft / 30)}
                  className="stroke-foreground transition-all duration-1000 ease-linear"
                />
              </svg>
              <span>Hiding in {secondsLeft}s</span>
            </div>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleHide}>
              <EyeOff className="h-3.5 w-3.5" />
            </Button>
          </div>
        )}

        {rowState === 'error' && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-destructive">Couldn't decrypt</span>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleReveal}>
              <RotateCcw className="h-3.5 w-3.5" />
            </Button>
          </div>
        )}
      </TableCell>

      <TableCell className="py-2.5 text-right">
        <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-7 w-7">
                <MoreHorizontal className="h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem>
                <Pencil className="mr-2 h-3.5 w-3.5" /> Edit
              </DropdownMenuItem>

              {isOwner ? (
                <DropdownMenuItem className="text-destructive focus:text-destructive">
                  <Trash2 className="mr-2 h-3.5 w-3.5" /> Delete
                </DropdownMenuItem>
              ) : (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div>
                        <DropdownMenuItem disabled>
                          <Trash2 className="mr-2 h-3.5 w-3.5" /> Delete
                        </DropdownMenuItem>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent side="left" className="text-xs">
                      Only the project owner can delete secrets
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </TableCell>
    </TableRow>
  );
};
