# Vaultify Frontend — Architecture Spec

Stack: React 19, Vite, Tailwind v4, shadcn/ui, TanStack Query v5, Axios, react-hook-form + Zod.
Scope: design only. No code. Intended for hand-off to Antigravity for implementation.

---

## 1. Architectural Critique of the Constraints (read first)

Before the directory structure, three things in your constraints have second-order
consequences you should decide on now, not mid-build:

**1.1 — "JWT in memory" actually means "JWT in a React context that survives route
changes but not refreshes."** A `useState` in the top-level `App` component is not
enough by itself — every component tree remount (e.g. a hard navigation, or React
Query's error boundary resetting) will lose it. You need a single `AuthProvider`
that owns the access token in a ref/state pair, and a non-React module (`tokenStore.js`)
that Axios's interceptor can read synchronously, because Axios interceptors run
outside React's render cycle and can't call `useContext`. This is a common mistake:
people put the token in context only, then can't reach it from the interceptor
without prop-drilling a setter into a singleton. Decide now: **token lives in a
plain JS module-level variable, mirrored into React context for re-renders.**

**1.2 — The PRD says JWT TTL is 10 minutes, and refresh rotates the token on every
use.** This means on a hard page reload, the access token is gone (by design) and
the app must silently call `/auth/refresh` once on boot using the HttpOnly cookie
before rendering anything that needs auth — otherwise every reload bounces the user
to login even though their session is valid. This is a "bootstrap" concern your
router needs to model explicitly (an `isBootstrapping` state, not just `isAuthenticated`).

**1.3 — Refresh-token rotation under concurrent requests is a real race condition,
not a hypothetical one.** If a user has 5 secrets loading in parallel (table +
audit log + members panel) and the access token expires mid-flight, you'll get 5
parallel 401s, and if your interceptor naively calls `/auth/refresh` 5 times, the
backend's rotation logic (delete-old-issue-new) will treat requests 2–5 as **replay
attacks** per your own spec (Doc 2, section "Refresh Token Rotation"), and nuke the
user's entire session. This is the single most important thing to get right in the
interceptor design, and it's detailed in section 3 below.

**1.4 — Masked secrets + 30-second reveal-then-revert is a UI state machine, not a
boolean.** "Revealed" has at least 4 states you need to model: `masked` →
`revealing` (decrypt request in flight, write audit log server-side) → `revealed`
(value shown + clipboard write + 30s timer running) → `masked` (timer fired or user
manually re-hid it). Treat this as a finite state machine per-row, not a `useState(false)`,
or you'll get bugs where switching tabs or re-rendering the table resets the timer
incorrectly. Detailed in section 5.

---

## 2. Directory Structure

```
/src
├── api/
│   ├── client.js                 # Axios instance, base URL = /api/v1
│   ├── interceptors.js           # 401 → refresh → retry queue (see section 3)
│   ├── tokenStore.js             # module-level in-memory access token, outside React
│   ├── endpoints/
│   │   ├── auth.api.js           # login, register, refresh, logout, me
│   │   ├── projects.api.js       # CRUD + members
│   │   ├── secrets.api.js        # CRUD + reveal
│   │   ├── tokens.api.js         # API token CRUD
│   │   └── audit.api.js          # paginated audit log
│   └── queryKeys.js              # centralized TanStack Query key factory
│
├── app/
│   ├── App.jsx                   # Router root, AuthProvider + QueryClientProvider wrap
│   ├── router.jsx                # Route tree (see section 4)
│   └── ProtectedRoute.jsx        # Gate on isBootstrapping / isAuthenticated
│
├── auth/
│   ├── AuthProvider.jsx          # Owns auth state, exposes login/logout/bootstrap
│   ├── useAuth.js                # Consumer hook
│   └── permissions.js            # Pure functions: canDeleteSecret(role, scope), etc.
│
├── features/
│   ├── projects/
│   │   ├── ProjectsListPage.jsx
│   │   ├── ProjectDashboardPage.jsx       # Master-detail shell (section 4)
│   │   ├── components/
│   │   │   ├── ProjectSidebar.jsx         # Master list pane
│   │   │   ├── ProjectCard.jsx
│   │   │   ├── CreateProjectDialog.jsx
│   │   │   ├── ProjectSettingsPanel.jsx   # Rename / delete (Owner only)
│   │   │   └── MembersPanel.jsx           # Invite / remove (Owner only)
│   │   └── hooks/
│   │       ├── useProjects.js             # list query
│   │       ├── useProject.js              # single project + members query
│   │       └── useProjectMutations.js     # create/rename/delete/invite/remove
│   │
│   ├── secrets/
│   │   ├── components/
│   │   │   ├── SecretsTable.jsx           # Critical component, section 5
│   │   │   ├── SecretRow.jsx              # One row, owns its own reveal FSM
│   │   │   ├── RevealButton.jsx           # Triggers decrypt + clipboard + timer
│   │   │   ├── SecretValueCell.jsx        # Masked/revealed rendering only
│   │   │   ├── CountdownRing.jsx          # Visual 30s countdown (svg or css)
│   │   │   ├── CreateSecretDialog.jsx     # react-hook-form + Zod
│   │   │   ├── EditSecretDialog.jsx
│   │   │   ├── DeleteSecretConfirm.jsx    # AlertDialog, Owner-gated
│   │   │   └── EnvironmentTabs.jsx        # Filter by ?env=
│   │   └── hooks/
│   │       ├── useSecrets.js              # list (names only) query
│   │       ├── useRevealSecret.js         # decrypt mutation, no caching of plaintext
│   │       ├── useSecretMutations.js      # create/update/delete
│   │       └── useRevealTimer.js          # the 30s FSM, see section 5
│   │
│   ├── audit/
│   │   ├── AuditLogPage.jsx
│   │   └── components/
│   │       ├── AuditTable.jsx             # paginated
│   │       ├── ActionBadge.jsx            # color-coded by action type
│   │       └── AuditFilters.jsx           # by user / action / date range
│   │
│   ├── api-tokens/
│   │   ├── ApiTokensPage.jsx
│   │   └── components/
│   │       ├── TokensTable.jsx
│   │       ├── CreateTokenDialog.jsx      # Shows raw token ONCE, copy + dismiss warning
│   │       └── RevokeTokenConfirm.jsx
│   │
│   └── auth/
│       ├── LoginPage.jsx
│       ├── RegisterPage.jsx
│       └── components/
│           ├── LoginForm.jsx              # react-hook-form + Zod
│           └── RegisterForm.jsx
│
├── components/
│   ├── ui/                       # shadcn/ui generated primitives (button, dialog, etc.)
│   └── layout/
│       ├── AppShell.jsx          # Top nav + sidebar + content slot
│       ├── TopNav.jsx
│       ├── EmptyState.jsx
│       ├── ErrorBoundaryFallback.jsx
│       └── LoadingSkeletons.jsx  # Table/row/card skeletons, matched to real layout
│
├── hooks/
│   ├── useClipboard.js           # navigator.clipboard wrapper + fallback + toast
│   ├── useDebounce.js
│   └── usePermission.js          # role+scope → boolean, wraps permissions.js
│
├── lib/
│   ├── queryClient.js            # QueryClient instance, default options
│   ├── cn.js                     # shadcn's class merge helper
│   └── constants.js               # roles, scopes, audit action enums, TTLs
│
└── types/                        # JSDoc typedefs or .d.ts if you add TS later
    ├── auth.d.ts
    ├── project.d.ts
    └── secret.d.ts
```

**Why this shape, specifically:**

- `api/` vs `features/*/hooks/` is a deliberate seam: `api/` is pure HTTP (knows
  about URLs, headers, status codes), `features/*/hooks/` is TanStack Query glue
  (knows about cache keys, invalidation, optimistic updates). Mixing them means
  every component that needs a different cache strategy has to also know the URL
  shape. Keep them separate even though it's two files instead of one.
- `permissions.js` is plain functions, not hooks, deliberately — RBAC display logic
  (hide a Delete button) and RBAC enforcement (backend middleware) are two different
  systems. Keeping permission checks as pure functions makes it obvious in code
  review that **the frontend check is cosmetic only** — the real gate is the backend
  middleware per Doc 4. Don't let anyone on the team (including future-you) think
  hiding a button is the security boundary.
- `SecretRow.jsx` owns its own reveal state rather than the table owning an array
  of "which rows are revealed" — this avoids one row's countdown timer triggering
  a full table re-render on every tick, which matters once a project has 50+ secrets.

---

## 3. Axios Interceptor — Refresh Queue Design

This is the part most portfolio projects get wrong, and it's also your best
interview talking point on the frontend side ("how did you handle concurrent 401s
without triggering your own replay-attack detection?").

**The failure mode to avoid:** 5 components mount, each fires a request, the access
token is already expired, you get 5 simultaneous 401s. If each one independently
calls `/auth/refresh`, only the first will succeed — by the time the 2nd-5th hit the
backend, the refresh token has already been rotated out from under them (Doc 2:
"On every single hit to `/auth/refresh`, the server immediately deletes the presented
refresh token UUID"). Requests 2–5 will look like replay attacks and the backend will
nuke the entire session (Doc 2: "invalidates the entire user session set"). You'd
log a real, legitimate user out by accident, every time two tabs load simultaneously.

**The fix — a single in-flight refresh promise, shared across all callers:**

- A module-level variable holds either `null` or a pending Promise for the current
  refresh call.
- Response interceptor catches a 401. Before retrying, it checks: is a refresh
  already in flight? If yes, attach to that existing promise instead of starting
  a new one. If no, start one and store the promise.
- Every queued request `await`s the shared promise, then retries with whatever
  new token the refresh resolved to.
- Once the promise resolves (or rejects), clear the module-level variable so the
  *next* expiry cycle starts fresh.
- One exception flag: requests to `/auth/login`, `/auth/register`, and
  `/auth/refresh` itself must never enter this retry loop — a 401 on those means
  "actually not authenticated," not "token expired," and retrying would infinite-loop.
- On a refresh failure (refresh token itself invalid/expired/replayed): clear the
  in-memory token, redirect to `/login`, and importantly — **don't auto-retry the
  original failed requests**; let them reject so the calling components can show
  appropriate error states rather than hanging.

**Diagram of the flow:**

```
Request A ──┐
Request B ──┼──► 401 401 401 ──► [is refresh in flight?]
Request C ──┘                          │
                                   No ──┴── Yes
                                   │          │
                          start refresh   await existing promise
                                   │          │
                          refresh resolves────┘
                                   │
                     retry A, B, C with new token
```

This single design decision — "collapse concurrent 401s into one refresh call" —
is worth more in an interview than almost anything else in the frontend, because
it shows you understood the backend's rotation/replay defense well enough to not
accidentally trip it from your own client.

---

## 4. Routing & The Master-Detail Dashboard

```
/login                              public
/register                           public
/                                   redirect → /projects
/projects                           ProjectsListPage (grid/list of project cards)
/projects/:projectId                ProjectDashboardPage — master-detail shell
  ?tab=secrets (default)            SecretsTable
  ?tab=members                      MembersPanel (Owner sees invite form, Member sees read-only list)
  ?tab=audit                        AuditTable
  ?tab=settings                     ProjectSettingsPanel (rename/delete, Owner only — 403-aware)
/tokens                             ApiTokensPage (account-level, not project-scoped)
*                                   404 / NotFound
```

**Master-detail shell layout** (`ProjectDashboardPage.jsx`):

- **Master pane (left, ~280px, sticky):** `ProjectSidebar` — list of all projects the
  user belongs to, current one highlighted, a role badge (Owner/Member) per project
  so the user always knows their permission level without opening it, "+ New Project"
  at the bottom. Clicking a project updates the URL param, doesn't remount the shell.
- **Detail pane (right, fluid):** Tab strip (Secrets / Members / Audit Log / Settings)
  using `?tab=` query param so the state is shareable/bookmarkable/back-button-able,
  rendering whichever feature panel is active.
- Why query params over local state for the tab: this is a portfolio project meant
  to look production-grade — being able to send someone a deep link straight to a
  project's audit log is the kind of detail that reads as senior-level thinking, not
  over-engineering, because it costs nothing extra here.

**Route-level RBAC awareness:** `ProjectSettingsPanel` and the "remove member"
action in `MembersPanel` should consult `usePermission()` to decide whether to
render the controls at all — but every mutating call still needs to handle a 403
gracefully (toast: "You don't have permission to do that" rather than a generic
error), because role can change server-side between page load and click (Doc 4:
"Role changes take effect immediately — there is no caching of roles"). Don't
just hide the button and assume that's sufficient defense-in-depth on the UI side;
handle the 403 response explicitly too.

---

## 5. The Secrets Table — Reveal/Mask State Machine

This is the component you flagged as most critical, so it gets the most detail.

**State machine per row** (lives in `SecretRow.jsx`, exposed via `useRevealTimer` hook):

```
masked ──(click Reveal)──► revealing ──(decrypt success)──► revealed ──(30s elapsed
  ▲                              │                              │       OR user clicks
  │                       (decrypt fails)                       │       "Hide now")
  └──────────────────────────────┴──────────────────────────────┘
                                                                  back to masked
```

- **`masked`**: default. Cell renders `••••••••` (fixed-width dots, not the actual
  ciphertext length — never leak value length via dot count). Reveal button enabled
  if the user's role/scope permits read (per Doc 4 matrix); disabled+tooltip otherwise.
- **`revealing`**: Reveal button shows a spinner, is disabled to prevent double-fire.
  This state exists specifically because the GET on `/secrets/:secretId` is not free
  — it's a decrypt operation server-side AND it writes an audit log entry (per
  api-endpoints.md). Don't let a user accidentally fire 3 reveals by double-clicking;
  that's 3 audit log entries and 3 decrypt operations for one intent.
- **`revealed`**: Plaintext rendered, value also written to clipboard (via
  `useClipboard`), a small inline countdown (`CountdownRing`, a 30→0 ring or text)
  shown next to the value, and a "Hide now" link for users who don't want to wait.
  **The 30-second timer is owned by `useRevealTimer`, using `useEffect` with a
  `setTimeout`, cleaned up on unmount AND on manual hide AND on row identity change**
  — three cleanup paths, not one, because: (a) user navigates away mid-countdown
  (unmount), (b) user clicks Hide (manual), (c) the table re-sorts/re-filters and
  this row's key changes (identity change). Missing any of these three leaks a
  dangling timer that tries to update state on an unmounted component or, worse,
  re-masks the *wrong* row if you're not careful with closures over the secret ID.
- **Plaintext is never put in TanStack Query's cache.** The reveal mutation result
  is held in **local component state only**, never `queryClient.setQueryData`'d into
  the secrets-list query — because Query's cache is inspectable in devtools and can
  persist longer than 30 seconds depending on `gcTime`/`staleTime` config. Treat the
  decrypted value as radioactive: it exists in one `useState` in one row component,
  for at most 30 seconds, and nowhere else. This is worth stating explicitly in your
  README as a deliberate security decision, mirroring the backend's "zero the
  plaintext from memory after use" philosophy from the PRD.
- **Clipboard write happens once, at the moment of reveal**, not re-written every
  re-render. Use a ref to guard against the effect firing twice in React 19 strict
  mode dev double-invocation.
- **Tab visibility edge case:** if the user backgrounds the tab during the 30s
  window, the `setTimeout` still fires on schedule (timers aren't paused by tab
  visibility in modern browsers), so no special handling needed there — but worth
  a comment in code so the next person doesn't "fix" something that isn't broken.

**Table-level concerns** (`SecretsTable.jsx`):

- Columns: Key Name, Environment (badge), Value (masked/revealed cell), Last Updated,
  Updated By, Actions (Reveal / Edit / Delete — Delete hidden/disabled for Members
  per the RBAC matrix in api-endpoints.md).
- `EnvironmentTabs` filters via the `?env=` query param the GET endpoint already
  supports — don't filter client-side after fetching all environments; request
  only what's needed.
- Skeleton loading state matches the real table's column widths exactly (defined
  in `LoadingSkeletons.jsx`) so there's no layout shift on data arrival — small
  detail, reads as polish.
- Empty state ("No secrets yet" + a CTA to open `CreateSecretDialog`) instead of
  a blank table — this is the kind of thing that separates a portfolio piece from
  a backend-engineer's-afterthought UI.

---

## 6. Forms & Validation

- `react-hook-form` + `Zod` resolver for: Login, Register, Create/Edit Secret,
  Create Project, Invite Member, Create API Token.
- Zod schemas live colocated with their dialog component (e.g.
  `CreateSecretDialog.jsx` exports its own schema) rather than centralized — these
  schemas are rarely reused across components and centralizing them just adds an
  import hop for no benefit at this project's size.
- Secret key name validation should mirror backend constraints (uppercase,
  underscores, no spaces — match whatever convention `DATABASE_URL`/`STRIPE_KEY`
  implies) so the user gets instant feedback instead of a round-trip 400.

---

## 7. Query Key Strategy (TanStack Query v5)

Centralize in `api/queryKeys.js` as a factory, not scattered string literals:

```
projects.all()              → ['projects']
projects.detail(id)         → ['projects', id]
secrets.list(projectId,env) → ['projects', projectId, 'secrets', { env }]
audit.list(projectId,page)  → ['projects', projectId, 'audit', { page }]
tokens.all()                → ['tokens']
```

Hierarchical keys matter here specifically because of project deletion: invalidating
`['projects', projectId]` should cascade-invalidate everything nested under it
(secrets, audit, members) in one `queryClient.invalidateQueries({ queryKey: ['projects', projectId] })`
call, matching the backend's own `ON DELETE CASCADE` semantics. Don't invent a key
shape where you have to remember to invalidate 4 separate queries by hand on every
delete — that's where stale-cache bugs live.

---

## 8. What NOT to Build (mirroring the PRD's own scope discipline)

Your PRD explicitly cuts scope to keep this a focused, finishable project. Carry
that discipline to the frontend:

- No client-side secret value caching/offline support — values are sensitive,
  request fresh every reveal.
- No optimistic updates on secret create/edit — the encrypt round-trip is the
  whole point of the product; faking instant UI feedback undersells what's
  actually happening server-side. A brief loading state is more honest here.
- No drag-and-drop, no dark-mode toggle, no multi-language support — these read
  as scope creep on a security tool's portfolio piece, not polish.
- No client-side role caching beyond the current session — re-fetch project
  membership/role on dashboard mount, don't persist it, since Doc 4 is explicit
  that role changes are immediate and uncached server-side; the frontend shouldn't
  be the place that introduces staleness the backend deliberately avoided.
