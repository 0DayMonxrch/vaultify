import React from 'react';
import { Outlet, Link, useParams, useLocation, useSearchParams } from 'react-router-dom';
import { Menu, ChevronDown, KeySquare, LogOut } from 'lucide-react';
import { useAuth } from '../../auth/useAuth';
import { useProjects } from '../../features/projects/hooks/useProjects';

import { ProjectSidebar } from './ProjectSidebar';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Breadcrumb, BreadcrumbList, BreadcrumbItem, BreadcrumbLink, BreadcrumbSeparator, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';
import { ModeToggle } from '../ModeToggle';

export const AppShell: React.FC = () => {
  const { user, logout } = useAuth();
  const { projectId } = useParams<{ projectId: string }>();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const { data: projects = [] } = useProjects();

  const activeProject = projects.find((p: any) => p.id === projectId);
  const projectName = activeProject ? activeProject.name : (location.pathname === '/tokens' ? 'API Tokens' : 'Projects');
  
  const currentTab = searchParams.get('tab') || 'secrets';
  let currentSection = 'Overview';
  
  const isProjectRoute = !!projectId;
  
  const tabs = isProjectRoute ? [
    { id: 'secrets', label: 'Secrets', href: `/projects/${projectId}?tab=secrets`, active: currentTab === 'secrets' },
    { id: 'members', label: 'Members', href: `/projects/${projectId}?tab=members`, active: currentTab === 'members' },
    { id: 'audit', label: 'Audit Log', href: `/projects/${projectId}?tab=audit`, active: currentTab === 'audit' },
    { id: 'settings', label: 'Settings', href: `/projects/${projectId}?tab=settings`, active: currentTab === 'settings' },
  ] : [];

  if (isProjectRoute) {
    const activeTabObj = tabs.find(t => t.active);
    if (activeTabObj) currentSection = activeTabObj.label;
  } else if (location.pathname === '/tokens') {
    currentSection = 'Personal Access Tokens';
  }

  const initials = user?.email ? user.email.substring(0, 2).toUpperCase() : 'U';

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <ProjectSidebar className="hidden md:flex" />

      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 flex-shrink-0 items-center justify-between border-b border-border px-4 md:px-6">
          <div className="flex items-center gap-3">
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 md:hidden">
                  <Menu className="h-4 w-4" />
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="p-0 w-[260px]">
                <ProjectSidebar className="flex border-r-0" />
              </SheetContent>
            </Sheet>

            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href={isProjectRoute ? `/projects/${projectId}` : '#'} className="text-sm">
                    {projectName}
                  </BreadcrumbLink>
                </BreadcrumbItem>
                {isProjectRoute && (
                  <>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      <BreadcrumbPage className="text-sm font-medium">{currentSection}</BreadcrumbPage>
                    </BreadcrumbItem>
                  </>
                )}
              </BreadcrumbList>
            </Breadcrumb>
          </div>

          <div className="flex items-center gap-2">
            <ModeToggle />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 gap-2 px-2">
                  <Avatar className="h-6 w-6">
                    <AvatarFallback className="text-xs">{initials}</AvatarFallback>
                  </Avatar>
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                </Button>
              </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <div className="px-2 py-1.5 text-xs text-muted-foreground">{user?.email}</div>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link to="/tokens" className="flex items-center w-full cursor-pointer">
                  <KeySquare className="mr-2 h-3.5 w-3.5" /> API Tokens
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-destructive focus:text-destructive cursor-pointer" onClick={() => logout()}>
                <LogOut className="mr-2 h-3.5 w-3.5" /> Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          </div>
        </header>

        {isProjectRoute && (
          <nav className="flex flex-shrink-0 gap-6 border-b border-border px-4 md:px-6">
            {tabs.map(tab => (
              <Link
                key={tab.id}
                to={tab.href}
                className={cn(
                  "border-b-2 py-3 text-sm transition-colors",
                  tab.active
                    ? "border-foreground font-medium text-foreground"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                )}
              >
                {tab.label}
              </Link>
            ))}
          </nav>
        )}

        <main className="flex-1 overflow-y-auto p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
