export type Role = 'owner' | 'member' | string;

export const canManageMembers = (role?: Role): boolean => {
  return role?.toLowerCase() === 'owner';
};

export const canDeleteProject = (role?: Role): boolean => {
  return role?.toLowerCase() === 'owner';
};

export const canUpdateProject = (role?: Role): boolean => {
  return role?.toLowerCase() === 'owner';
};

export const canDeleteSecret = (role?: Role): boolean => {
  return role?.toLowerCase() === 'owner';
};

export const canCreateWriteTokens = (role?: Role): boolean => {
  return role?.toLowerCase() === 'owner';
};

export const canWriteSecrets = (role?: Role): boolean => {
  // Both Owners and Members can create/update secrets
  return role?.toLowerCase() === 'owner' || role?.toLowerCase() === 'member';
};
