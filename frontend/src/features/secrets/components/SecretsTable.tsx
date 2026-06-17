import React from 'react';
import { useSearchParams, useParams } from 'react-router-dom';
import { useSecrets } from '../hooks/useSecrets';
import { SecretRow } from './SecretRow';
import { KeyRound, Plus } from 'lucide-react';
import { EnvironmentTabs } from './EnvironmentTabs';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from 'sonner';

export const SecretsTable: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const [searchParams] = useSearchParams();
  const env = searchParams.get('env') || 'production';

  const { data: secrets = [], isLoading } = useSecrets(projectId!, env);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center mb-4">
         <EnvironmentTabs />
         <Button size="sm" onClick={() => toast.info('Secret creation backend integration is scheduled for the next phase!')}>
           <Plus className="mr-1.5 h-3.5 w-3.5" /> Add secret
         </Button>
      </div>
      <div className="overflow-hidden rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="text-xs uppercase tracking-wide text-muted-foreground">Key</TableHead>
              <TableHead className="text-xs uppercase tracking-wide text-muted-foreground">Environment</TableHead>
              <TableHead className="text-xs uppercase tracking-wide text-muted-foreground">Value</TableHead>
              <TableHead className="text-right text-xs uppercase tracking-wide text-muted-foreground">Actions</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {isLoading ? (
              <TableRow>
                 <TableCell colSpan={4} className="py-12 text-center text-muted-foreground">Loading secrets...</TableCell>
              </TableRow>
            ) : secrets.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="py-12 text-center">
                  <KeyRound className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
                  <p className="text-sm font-medium">No secrets yet</p>
                  <p className="mb-4 text-sm text-muted-foreground">
                    Add your first secret to this environment.
                  </p>
                  <Button size="sm" onClick={() => toast.info('Secret creation backend integration is scheduled for the next phase!')}>
                    <Plus className="mr-1.5 h-3.5 w-3.5" /> Add secret
                  </Button>
                </TableCell>
              </TableRow>
            ) : (
              secrets.map((secret: any) => <SecretRow key={secret.id} project_id={projectId!} secret={secret} />)
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
};
