import { useParams } from 'react-router-dom';
import { useProjects } from '../features/projects/hooks/useProjects';
import { useProjectMembers } from '../features/projects/hooks/useProject';
import { useAuth } from '../auth/useAuth';
import * as permissions from '../auth/permissions';

export const usePermission = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const { data: projects } = useProjects();
  const { data: members } = useProjectMembers(projectId || '');
  const { user } = useAuth();
  
  const activeProject = projects?.find(p => p.id === projectId);
  
  let role = 'member'; // fallback
  if (members && user) {
     const myRecord = members.find((m: any) => m.user_id === user.id);
     if (myRecord) {
        role = myRecord.role;
     }
  } else if (activeProject && (activeProject as any).role) {
     role = (activeProject as any).role;
  }

  return {
    canManageMembers: permissions.canManageMembers(role),
    canDeleteProject: permissions.canDeleteProject(role),
    canUpdateProject: permissions.canUpdateProject(role),
    canDeleteSecret: permissions.canDeleteSecret(role),
    canCreateWriteTokens: permissions.canCreateWriteTokens(role),
    canWriteSecrets: permissions.canWriteSecrets(role),
    role,
  };
};
