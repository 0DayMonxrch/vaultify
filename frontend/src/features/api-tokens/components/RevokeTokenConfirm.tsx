import React, { useState } from 'react';
import { useRevokeToken } from '../hooks/useTokens';
import { Button } from '../../../components/ui/button';

interface Props {
  tokenId: string;
  tokenName: string;
}

export const RevokeTokenConfirm: React.FC<Props> = ({ tokenId, tokenName }) => {
  const [isOpen, setIsOpen] = useState(false);
  const { mutateAsync: revokeToken, isPending } = useRevokeToken();

  const handleRevoke = async () => {
    try {
      await revokeToken(tokenId);
      setIsOpen(false);
    } catch(e) {
      console.error(e);
    }
  };

  return (
    <>
      <Button variant="ghost" size="xs" onClick={() => setIsOpen(true)} className="text-destructive hover:bg-destructive/10 hover:text-destructive h-7 text-xs px-3">
        Revoke
      </Button>

      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div className="bg-card border border-border rounded-xl p-8 w-full max-w-md shadow-2xl animate-in zoom-in-95 duration-200">
            <h2 className="text-xl font-bold mb-4 text-destructive tracking-tight">Revoke Token?</h2>
            <p className="text-sm text-muted-foreground mb-8 leading-relaxed">
              Are you sure you want to revoke the token <strong className="text-foreground">{tokenName}</strong>? Any systems using this token will instantly lose access to the API. This action cannot be undone.
            </p>
            <div className="flex justify-end space-x-3 pt-4 border-t border-border">
              <Button variant="ghost" onClick={() => setIsOpen(false)}>Cancel</Button>
              <Button variant="destructive" onClick={handleRevoke} disabled={isPending}>
                {isPending ? 'Revoking...' : 'Revoke Immediately'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
