import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useProjectMembers } from '../hooks/useProject';
import { useAddMember, useRemoveMember } from '../hooks/useProjectMutations';
import { usePermission } from '../../../hooks/usePermission';
import { Button } from '../../../components/ui/button';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';

const inviteSchema = z.object({
  email: z.string().email(),
  role: z.enum(['owner', 'member'])
});

export const MembersPanel: React.FC = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const { data: members, isLoading } = useProjectMembers(projectId!);
  const { canManageMembers } = usePermission();
  const { mutateAsync: removeMember } = useRemoveMember(projectId!);
  const { mutateAsync: addMember, isPending: isAdding } = useAddMember(projectId!);
  
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const { register, handleSubmit, reset, formState: { errors } } = useForm<z.infer<typeof inviteSchema>>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { role: 'member' }
  });

  const onInvite = async (data: z.infer<typeof inviteSchema>) => {
    try {
      await addMember(data);
      setIsInviteOpen(false);
      reset();
    } catch(e: any) {
      if (e.response?.status === 404) {
        toast.error('User not found. They must create an account first.');
      } else {
        toast.error('Failed to invite member. Please try again.');
      }
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-card p-4 rounded-xl border border-border/50 shadow-sm">
        <div>
          <h2 className="text-lg font-bold text-foreground">Project Members</h2>
          <p className="text-sm text-muted-foreground">Manage who has access to this project's secrets.</p>
        </div>
        {canManageMembers && (
          <Button onClick={() => setIsInviteOpen(true)}>Invite Member</Button>
        )}
      </div>

      <div className="border border-border rounded-xl overflow-hidden bg-card shadow-sm">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-muted/30 border-b border-border/50">
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Email</th>
              <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest">Role</th>
              {canManageMembers && (
                <th className="p-4 text-[11px] font-bold text-muted-foreground uppercase tracking-widest text-right">Actions</th>
              )}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr><td colSpan={3} className="p-8 text-center text-muted-foreground">Loading members...</td></tr>
            ) : members?.map(member => (
              <tr key={member.user_id} className="border-b border-border/50">
                <td className="p-4 font-medium text-sm">{member.email || member.user_id}</td>
                <td className="p-4">
                  <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded bg-secondary text-secondary-foreground">
                    {member.role}
                  </span>
                </td>
                {canManageMembers && (
                  <td className="p-4 text-right">
                    <Button variant="ghost" size="xs" onClick={() => removeMember(member.user_id)} className="text-destructive hover:bg-destructive/10 hover:text-destructive">
                      Remove
                    </Button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isInviteOpen && canManageMembers && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div className="bg-card border border-border rounded-xl p-8 w-full max-w-md shadow-2xl animate-in zoom-in-95 duration-200">
            <h2 className="text-xl font-bold mb-6">Invite Member</h2>
            <form onSubmit={handleSubmit(onInvite)} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1.5">Email Address</label>
                <input {...register('email')} className="w-full px-3 py-2 border rounded-lg bg-background focus:ring-2 focus:ring-primary/50" />
                {errors.email && <p className="text-destructive text-xs mt-1">{errors.email.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5">Role</label>
                <select {...register('role')} className="w-full px-3 py-2 border rounded-lg bg-background focus:ring-2 focus:ring-primary/50">
                  <option value="member">Member</option>
                  <option value="owner">Owner</option>
                </select>
              </div>
              <div className="flex justify-end space-x-3 pt-4 border-t border-border mt-4">
                <Button variant="ghost" type="button" onClick={() => setIsInviteOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={isAdding}>{isAdding ? 'Inviting...' : 'Send Invite'}</Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
