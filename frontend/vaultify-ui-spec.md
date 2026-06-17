# Vaultify — Frontend UI Design Spec

Target stack: React 19, Tailwind CSS v4, shadcn/ui, lucide-react.
Audience: implementation agent. This document specifies DOM structure, shadcn component composition, and Tailwind utility classes only. No data-fetching, validation, or business logic is included — assume `useProjects()`, `useSecrets()`, `useRevealTimer()`, and the auth/form hooks already exist and are wired in.

---

## 0. Design Tokens & Global Conventions

This is a dense, professional developer tool (Vercel / Linear / Doppler register), not a marketing surface. Stay inside the default shadcn **Zinc** theme — do not introduce a custom brand hue. Color is used semantically, not decoratively.

**Color**
- Base: `bg-background` / `text-foreground` (zinc-950 on zinc-50 in light mode, inverted in dark).
- Structure: `border-border`, `bg-muted`, `text-muted-foreground` for secondary text, dividers, and inactive states.
- Semantic only: `text-destructive` / `border-destructive` for delete & errors, `emerald-600` for success/confirmation micro-states (e.g. "copied"), `amber` for the `production` environment badge (a deliberate "be careful here" signal — see §3).
- Never invent a new accent color for primary actions. Primary buttons use the shadcn default `Button` (`variant="default"`).

**Typography**
- UI text: `font-sans` (Inter, shadcn default).
- Anything that is data, not prose — secret values, key names, token prefixes, project slugs, timestamps — uses `font-mono`. This is the single biggest visual cue that separates "labels" from "values" in a secrets manager, lean on it consistently.
- Base size in dense contexts is `text-sm` (14px), dropping to `text-xs` for badges, table headers, and timestamps. Avoid `text-base` anywhere inside the table or sidebar.

**Density & Radius**
- Table rows ~40–44px tall (`py-2.5` on cells). This is a data tool: prioritize rows-per-screen over generous whitespace.
- Consistent `rounded-md` for inputs/buttons/badges, `rounded-lg` as the ceiling for cards/panels. Nothing more rounded than that anywhere in the app.

**Icons**
- Inline icons in dense rows: `h-3.5 w-3.5`. Standalone icon buttons: `h-4 w-4`. Never mix sizes within the same row.

---

## 1. `LoginForm.tsx`

A centered card, no marketing chrome.

```
<div className="flex min-h-screen items-center justify-center bg-background p-4">
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
      {/* Server-level auth error — distinct from per-field zod errors below */}
      {serverError && (
        <Alert variant="destructive" className="mb-4">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Sign in failed</AlertTitle>
          <AlertDescription className="text-sm">{serverError}</AlertDescription>
        </Alert>
      )}

      <Form {...form}>
        <form className="space-y-4">

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
                <a href="#" className="text-xs text-muted-foreground hover:text-foreground hover:underline">
                  Forgot password?
                </a>
              </div>
              <FormControl>
                <Input type="password" autoComplete="current-password" {...field} />
              </FormControl>
              <FormMessage className="text-xs" />
            </FormItem>
          )} />

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? (
              <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Signing in...</>
            ) : "Sign in"}
          </Button>

        </form>
      </Form>
    </CardContent>

    <CardFooter className="justify-center">
      <p className="text-sm text-muted-foreground">
        Don't have an account?{" "}
        <a href="#" className="font-medium text-foreground underline-offset-4 hover:underline">
          Create one
        </a>
      </p>
    </CardFooter>

  </Card>
</div>
```

**Validation states to wire up**
- *Idle*: default `Input` border.
- *Field error* (zod): `FormMessage` renders the message; shadcn's `Form` primitives apply `text-destructive` and a destructive border automatically via `aria-invalid` — no manual className needed.
- *Submitting*: button disabled, label swaps to spinner + "Signing in...", inputs remain editable but submit is blocked.
- *Server error* (401 invalid credentials, account locked, etc.): the `Alert` above the form, not a toast — it should persist until the next submit attempt, since toasts disappear before a user reading slowly has finished.

---

## 2. `AppShell.tsx` & `ProjectSidebar.tsx`

### AppShell.tsx

```
<div className="flex h-screen overflow-hidden bg-background">

  <ProjectSidebar className="hidden md:flex" />

  <div className="flex flex-1 flex-col overflow-hidden">

    <header className="flex h-14 flex-shrink-0 items-center justify-between border-b border-border px-4 md:px-6">
      <div className="flex items-center gap-3">
        {/* Mobile sidebar trigger — Sheet, see note below */}
        <Button variant="ghost" size="icon" className="h-8 w-8 md:hidden">
          <Menu className="h-4 w-4" />
        </Button>

        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink href="#" className="text-sm">{projectName}</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage className="text-sm font-medium">{currentSection}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </div>

      <UserMenu />
    </header>

    {/* Project-level sub-navigation: Secrets / Members / Audit Log / Tokens */}
    <nav className="flex flex-shrink-0 gap-6 border-b border-border px-4 md:px-6">
      {tabs.map(tab => (
        <a
          key={tab.id}
          href={tab.href}
          className={cn(
            "border-b-2 py-3 text-sm transition-colors",
            tab.active
              ? "border-foreground font-medium text-foreground"
              : "border-transparent text-muted-foreground hover:text-foreground"
          )}
        >
          {tab.label}
        </a>
      ))}
    </nav>

    <main className="flex-1 overflow-y-auto p-4 md:p-6">
      {children}
    </main>

  </div>
</div>
```

`UserMenu` (used above):

```
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
    <div className="px-2 py-1.5 text-xs text-muted-foreground">{userEmail}</div>
    <DropdownMenuSeparator />
    <DropdownMenuItem>
      <Settings className="mr-2 h-3.5 w-3.5" /> Settings
    </DropdownMenuItem>
    <DropdownMenuItem>
      <KeySquare className="mr-2 h-3.5 w-3.5" /> API Tokens
    </DropdownMenuItem>
    <DropdownMenuSeparator />
    <DropdownMenuItem className="text-destructive focus:text-destructive">
      <LogOut className="mr-2 h-3.5 w-3.5" /> Log out
    </DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

### ProjectSidebar.tsx

Fixed width, ~260px.

```
<aside className="flex w-[260px] flex-shrink-0 flex-col border-r border-border bg-zinc-50/50 dark:bg-zinc-900/30">

  <div className="flex h-14 items-center gap-2 border-b border-border px-4">
    <KeyRound className="h-4 w-4" />
    <span className="text-sm font-semibold">Vaultify</span>
  </div>

  <div className="px-3 pt-3">
    <Button variant="outline" className="w-full justify-start gap-2 text-sm">
      <Plus className="h-3.5 w-3.5" /> New Project
    </Button>
  </div>

  <div className="flex-1 overflow-y-auto px-3 py-3">
    <p className="px-2 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
      Projects
    </p>
    <nav className="flex flex-col gap-0.5">
      {projects.map(project => (
        <a
          key={project.id}
          href={`/projects/${project.slug}`}
          className={cn(
            "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
            project.active
              ? "bg-zinc-200/70 font-medium text-foreground dark:bg-zinc-800"
              : "text-muted-foreground hover:bg-zinc-100 hover:text-foreground dark:hover:bg-zinc-800/50"
          )}
        >
          <Folder className="h-3.5 w-3.5 flex-shrink-0" />
          <span className="truncate">{project.name}</span>
        </a>
      ))}
    </nav>
  </div>

</aside>
```

**Mobile note:** below `md`, this sidebar is not rendered inline. Mount the same `nav` contents inside a shadcn `Sheet` (`side="left"`) triggered by the `Menu` button in the header. Don't build a second, separate mobile sidebar component — render `<ProjectSidebar>`'s inner content into the `Sheet` to avoid drift between the two.

---

## 3. `SecretsTable.tsx` & `SecretRow.tsx` — the core view

### State machine (authoritative — implement exactly this)

| State | Entered when | Renders | Enabled actions | Exits to |
|---|---|---|---|---|
| **Masked** | initial mount, after Hide, after 30s expiry, after Error auto-resets | `••••••••••••` placeholder, value in state is `null` | Reveal | → Revealing (click Reveal) |
| **Revealing** | Reveal clicked | placeholder + spinner, **all controls in the row disabled** | none | → Revealed (fetch ok) / → Error (fetch fails) |
| **Revealed** | decrypt fetch succeeds | plaintext (`font-mono`), Copy button, countdown | Copy, Hide | → Masked + purge value (30s elapses, or Hide clicked) |
| **Error** | decrypt fetch fails | inline error text + retry | Retry | → Revealing (click Retry) |

The disabled-during-Revealing rule exists specifically to stop double-clicks from firing duplicate decrypt calls and duplicate `SECRET_READ` audit log entries. On exit from Revealed, the component must set the value back to `null` in state — not merely toggle a `hidden` className — so it doesn't persist in memory or show up in React DevTools after hiding.

### SecretsTable.tsx

```
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
      {secrets.length === 0 ? (
        <TableRow>
          <TableCell colSpan={4} className="py-12 text-center">
            <KeyRound className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
            <p className="text-sm font-medium">No secrets yet</p>
            <p className="mb-4 text-sm text-muted-foreground">
              Add your first secret to this environment.
            </p>
            <Button size="sm"><Plus className="mr-1.5 h-3.5 w-3.5" /> Add secret</Button>
          </TableCell>
        </TableRow>
      ) : (
        secrets.map(secret => <SecretRow key={secret.id} secret={secret} />)
      )}
    </TableBody>
  </Table>
</div>
```

### SecretRow.tsx

```
<TableRow className="group border-b border-border last:border-0 hover:bg-zinc-50 dark:hover:bg-zinc-900/40">

  <TableCell className="py-2.5">
    <div className="flex items-center gap-1.5">
      <Key className="h-3.5 w-3.5 text-muted-foreground" />
      <span className="font-mono text-sm font-medium">{secret.keyName}</span>
    </div>
  </TableCell>

  <TableCell className="py-2.5">
    <Badge variant="outline" className={cn("text-xs font-medium uppercase", envBadgeClass(secret.environment))}>
      {secret.environment}
    </Badge>
  </TableCell>

  <TableCell className="py-2.5">
    {/* exactly one of the four blocks below renders, based on row state */}

    {/* Masked */}
    <div className="flex items-center gap-2">
      <span className="select-none font-mono text-sm text-muted-foreground">••••••••••••</span>
      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={reveal}>
        <Eye className="h-3.5 w-3.5" />
      </Button>
    </div>

    {/* Revealing */}
    <div className="flex items-center gap-2">
      <span className="select-none font-mono text-sm text-muted-foreground">••••••••••••</span>
      <Button variant="ghost" size="icon" className="h-7 w-7" disabled>
        <Loader2 className="h-3.5 w-3.5 animate-spin" />
      </Button>
    </div>

    {/* Revealed */}
    <div className="flex items-center gap-3">
      <span className="max-w-[220px] truncate font-mono text-sm">{secret.value}</span>
      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={copy}>
        {justCopied
          ? <Check className="h-3.5 w-3.5 text-emerald-600" />
          : <Copy className="h-3.5 w-3.5" />}
      </Button>
      <div className="flex items-center gap-1.5 text-xs tabular-nums text-muted-foreground">
        <svg viewBox="0 0 16 16" className="h-4 w-4 -rotate-90">
          <circle cx="8" cy="8" r="6.5" fill="none" strokeWidth="2" className="stroke-muted" />
          <circle
            cx="8" cy="8" r="6.5" fill="none" strokeWidth="2"
            strokeDasharray={40.8}
            strokeDashoffset={40.8 * (1 - secondsLeft / 30)}
            className="stroke-foreground"
          />
        </svg>
        <span>Hiding in {secondsLeft}s</span>
      </div>
      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={hide}>
        <EyeOff className="h-3.5 w-3.5" />
      </Button>
    </div>

    {/* Error */}
    <div className="flex items-center gap-2">
      <span className="text-xs text-destructive">Couldn't decrypt</span>
      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={retry}>
        <RotateCcw className="h-3.5 w-3.5" />
      </Button>
    </div>
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
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </TableCell>

</TableRow>
```

**Notes on the above**
- The actions column is hidden until row hover (`opacity-0 group-hover:opacity-100` on the row's `group`) — standard Linear/Vercel convention to keep a dense table visually quiet. The Reveal/Copy/Hide controls in the Value column are *not* hover-gated; they're a primary action, always visible.
- `envBadgeClass` is a small helper, not a hook — it's pure styling, safe to inline: `production` → `border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-400`; `staging` → blue equivalents; anything else (development, etc.) → neutral `border-border bg-muted text-muted-foreground`. Production is the only environment that should visually read as "be careful."
- The countdown ring is plain inline SVG, not a shadcn component — shadcn's `Progress` is a horizontal bar and doesn't fit this affordance. `strokeDasharray={40.8}` is `2 * π * 6.5`; recompute if the radius changes.
- Copy's "just copied" check-icon swap should revert to the `Copy` icon after ~1.5s on its own timer, independent of the 30s reveal countdown — copying doesn't extend or reset the reveal window.
