# Plik Webapp — Architecture & Gotchas

> Non-obvious details, design decisions, and pitfalls that agents should know before iterating on this codebase. For system-wide overview, see the root [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## Tech Stack

| Layer       | Tech                          |
|-------------|-------------------------------|
| Framework   | Vue 3 (Composition API, `<script setup>`) |
| Router      | Vue Router 4, hash history (`#/`) |
| Styling     | Tailwind CSS v4 (via `@import "tailwindcss"`) with custom `@utility` and `@theme` blocks |
| Code Editor | CodeMirror 6 (`@codemirror/language-data` for syntax, `@codemirror/theme-one-dark`) |
| Build       | Vite                          |
| HTTP        | `fetch()` for JSON APIs, `XMLHttpRequest` for file uploads (progress tracking) |
| Backend     | Go (Plik server, serves the SPA from `webapp/dist/` via `http.FileServer`) |

---

## Routing & URL Format

All routes use hash-history (`#/`):

| Route          | View            | Purpose                                   |
|----------------|-----------------|-------------------------------------------|
| `/#/`          | `RootView`      | Upload (no query) or Download (`?id=...`) |
| `/#/login`     | `LoginView`     | Local + OAuth login                       |
| `/#/home`      | `HomeView`      | User dashboard (uploads, tokens, account) |
| `/#/admin`     | `AdminView`     | Admin panel (stats, users, all uploads)   |
| `/#/clients`   | `ClientsView`   | CLI client downloads                      |
| `/#/upload/:id`| (redirect)      | Legacy URL → `/?id=:id`                   |

Admin link (upload-level): `/#/?id=<uploadId>&uploadToken=<token>`

`RootView.vue` checks `route.query.id` — if present, renders `DownloadView`; otherwise `UploadView`.

> **Gotcha**: The router uses `createWebHashHistory()`, so all URLs include `#/`. The `base` in `api.js` is computed from `window.location.origin + pathname` (without hash), so API calls go to the correct backend path.

### Auth Navigation Guard

When `config.feature_authentication` is `"forced"`, a `router.beforeEach` guard redirects unauthenticated users to `/#/login`. Exceptions:
- The login page itself (`to.name === 'login'`)
- CLI client downloads (`to.name === 'clients'`) — so users can get the CLI without logging in
- Download pages (`to.name === 'root' && to.query.id`) — so shared links still work

---

## Upload Token (Admin Auth)

### How it works

The Plik server generates an `uploadToken` when an upload is created. This token grants admin access (add/remove files, delete upload). It is returned as part of the upload JSON response.

### Token lifecycle

```
UploadView creates upload → server returns { id, uploadToken, ... }
  ↓
Token stored in memory (tokenStore.js) via setToken(id, token)
  ↓
Router navigates to /?id=<id>  (NO token in URL)
  ↓
DownloadView reads token from getToken(id) → sends as X-UploadToken header
```

### Admin URL sharing

The **Admin URL** (shown in DownloadView sidebar) is the only way to share admin access:
```
http://host/#/?id=<id>&uploadToken=<token>
```

When someone opens an admin URL:
1. `onMounted` in DownloadView reads `uploadToken` from `route.query`
2. Stores it in memory via `setToken(id, token)`
3. **Immediately strips it from the URL** via `router.replace()` to prevent accidental sharing

### Key rules

- **Never persist tokens** to `localStorage`, `sessionStorage`, or cookies
- **Never leave tokens in the URL** after initial load — strip immediately
- Tokens are sent as `X-UploadToken` header, never in the request body
- Token is **per-tab, per-session** — refreshing the page loses admin access (by design)
- `getUpload()` and other API calls pass the token so the server returns `admin: true` in the response

---

## File Status Values

Files in the Plik API have a `status` field with 5 possible values:

| Status      | Meaning                                           | Displayed? |
|-------------|---------------------------------------------------|------------|
| `missing`   | File entry created, waiting to be uploaded         | ✅ Yes (uploading UI) |
| `uploading` | File is currently being uploaded                   | ✅ Yes (progress bar) |
| `uploaded`  | File has been uploaded and is ready for download   | ✅ Yes      |
| `removed`   | File has been removed by user, not yet cleaned up  | ❌ No       |
| `deleted`   | File has been deleted from the data backend        | ❌ No       |

### activeFiles computed property

Files with `removed` or `deleted` status are filtered out. All other statuses are displayed:

```js
const activeFiles = computed(() => {
  return upload.value.files.filter(f => f.status !== 'removed' && f.status !== 'deleted')
})
```

> **Gotcha**: After deleting a file via the API, the server returns `"ok"` (plain text, not JSON). The file's status changes to `removed` server-side but the API **does not return the updated file object**. You must call `fetchUpload()` again to refresh the list.

> **Gotcha**: If `activeFiles.length === 0` after fetching, DownloadView **redirects to home** (`/`). This handles the case where all files have been deleted or consumed (one-shot).

---

## API Error Handling

### The two-pass text→JSON pattern

The Plik server returns errors as **either JSON or plain text** depending on the endpoint. The `apiCall` function handles this with a two-pass approach:

```js
// 1. Read as text first (always works)
const text = await resp.text()
// 2. Try to parse as JSON
try {
    const body = JSON.parse(text)
    message = body.message || body || message
} catch {
    // 3. Fall back to raw text (e.g., "upload abc123 not found")
    message = text || message
}
```

> **Why**: Calling `resp.json()` first consumes the body stream. If it fails (plain text response), `resp.text()` would then also fail with "body stream already read". The text-first approach avoids this.

### Success responses

Some endpoints return **plain text** on success:
- `DELETE /upload/:id` → `"ok"`
- `DELETE /file/:uploadId/:fileId/:fileName` → `"ok"`

The `apiCall` function handles this too:
```js
const text = await resp.text()
if (!text) return null
try { return JSON.parse(text) } catch { return text }
```

### Error display

Errors from `fetchUpload` are displayed inline (not redirecting). This shows the actual server message like `"upload feafea not found"` instead of a generic error.

---

## File Upload Mechanics

### Two URL patterns for uploading files

| Scenario                   | URL pattern                                  |
|----------------------------|----------------------------------------------|
| Initial upload (has fileId from `createUpload`) | `POST /file/:uploadId/:fileId/:fileName` |
| Adding files to existing upload (no fileId)     | `POST /file/:uploadId`                   |

The `api.js` `uploadFile` function picks the right pattern:
```js
if (fileEntry.id) {
    url = `${base}/${mode}/${upload.id}/${fileEntry.id}/${fileEntry.fileName}`
} else {
    url = `${base}/${mode}/${upload.id}`
}
```

### Stream vs File mode

The URL prefix changes based on whether the upload uses streaming:
- Normal: `/file/...`
- Streaming: `/stream/...`

### Upload flow (UploadView)

1. `createUpload(params)` → server returns upload with `id`, `uploadToken`, and pre-created file entries (with IDs)
2. For each file: `uploadFile(upload, fileEntry, onProgress)` → sends FormData with the `file` field
3. On success: `setToken(id, token)`, then `router.push({ path: '/', query: { id } })`

### Staged upload flow (DownloadView)

When adding files to an existing upload:
1. `onFilesSelected` stages files in `pendingFiles` ref (NOT uploaded yet)
2. User sees staged files with remove buttons, can review before uploading
3. Clicking "Upload" runs `uploadPendingFiles()` which uploads each file with progress
4. After all uploads: `pendingFiles` cleared, `fetchUpload()` refreshes the list
5. Files added to existing uploads have **no pre-created fileId** — server assigns one

---

## Staged File Object Shape

Files stored locally before upload use this shape (NOT the server shape):

```js
{
  reference: 'ref-1707123456-1',  // Local unique ID (from generateRef())
  fileName: 'photo.jpg',
  size: 1048576,
  file: File,                      // The browser File object
  status: 'toUpload',             // 'toUpload' | 'uploading' | 'uploaded' | 'error'
  progress: 0,                    // 0-100 upload progress
}
```

> **Gotcha**: Local files use `reference` as a key, not `id`. The `id` is only assigned by the server after upload. The `size` field is `size` locally but `fileSize` in server responses.

---

## Feature Flags (Config)

The server exposes feature flags via `GET /config`:

| Value       | Meaning                                     |
|-------------|---------------------------------------------|
| `enabled`   | Feature is available, default off            |
| `disabled`  | Feature is hidden entirely                   |
| `forced`    | Feature is on, user cannot toggle it off     |
| `default`   | Feature is available, default on             |

The config object keys use the pattern `feature_<name>` (e.g., `feature_one_shot`, `feature_stream`).

Helper functions (in `config.js`):
- `isFeatureEnabled(name)` → returns `true` unless value is `"disabled"`
- `isFeatureForced(name)` → returns `true` only if value is `"forced"`
- `isFeatureDefaultOn(name)` → returns `true` if value is `"default"` or `"forced"` (controls initial toggle state)

### Other Config Keys

The `GET /config` response also includes:

| Key | Purpose |
|-----|---------||
| `maxFileSize` | Max file size in bytes (shown in upload drop zone) |
| `maxUserSize` | Max total size per user |
| `maxTTL` | Max TTL in seconds |
| `googleAuthentication` | `true` if Google OAuth is configured → shows Google login button |
| `ovhAuthentication` | `true` if OVH OAuth is configured → shows OVH login button |
| `localAuthentication` | `true` if local login is enabled (auth enabled AND `DisableLocalLogin` is `false`) |
| `oidcAuthentication` | `true` if OIDC is configured → shows OIDC login button |
| `oidcProviderName` | Display name for OIDC button (e.g. `"Keycloak"`, defaults to `"OpenID"`) |
| `downloadDomain` | Alternate domain for download URLs (set in `api.js` via `setDownloadDomain`) |
| `abuseContact` | Abuse contact email → displayed in global footer (`App.vue`) |

---

## Size & TTL Limit Precedence

The server enforces layered limits: **user-specific → server config**. The special values `0` (use default) and `-1` (unlimited) are key.

### Value Semantics

| Value | Meaning |
|-------|---------|
| `> 0` | Explicit limit (bytes for size, seconds for TTL) |
| `0`   | Use server default |
| `-1`  | Unlimited (no limit enforced) |

### Precedence Rules (from `server/context/upload.go`)

**MaxFileSize** (`GetMaxFileSize()`):
```
if user != nil && user.MaxFileSize != 0 → user.MaxFileSize
else → config.MaxFileSize
```

**MaxUserSize** (`GetUserMaxSize()`):
```
if user == nil → unlimited (-1)                    // anonymous = no user quota
if user.MaxUserSize > 0 → user.MaxUserSize          // explicit user limit
if user.MaxUserSize < 0 → unlimited (-1)            // user explicitly unlimited
if user.MaxUserSize == 0 → config.MaxUserSize        // fall back to server default
```

**MaxTTL** (inside `setTTL()`):
```
maxTTL = config.MaxTTL
if user != nil && user.MaxTTL != 0 → maxTTL = user.MaxTTL
if maxTTL > 0 → enforce (reject infinite or over-limit TTL)
if maxTTL <= 0 → no limit enforced
```

### Effective Limit Calculation (Client-Side)

`UploadView.vue` computes effective limits via `auth.user` with fallback to `config`:

```js
const effectiveMaxFileSize = computed(() => {
  const user = auth.user
  if (user && user.maxFileSize !== 0 && user.maxFileSize !== undefined) return user.maxFileSize
  return config.maxFileSize
})
```

The same pattern applies for `effectiveMaxTTL`, which is passed as a prop to `UploadSidebar`.

### Size Unit Convention (SI / 1000-based)

> [!IMPORTANT]
> All size formatting uses **SI units** (1 GB = 1,000,000,000 bytes), matching the server's `go-humanize` library (`humanize.ParseBytes` / `humanize.Bytes`). The `GB` constant in edit modals is `1000³`, not `1024³`.

This means:
- Config `MaxFileSizeStr = "10GB"` → 10,000,000,000 bytes → displays "10.00 GB" everywhere
- Admin enters "1" GB in edit modal → stores 1,000,000,000 bytes → shows "1.00 GB"

> [!CAUTION]
> Never use 1024-based division with "GB" labels. If you need binary units, use "GiB" labels with 1024-based math.

---

## Component Architecture

```
App.vue
├── AppHeader.vue          — top nav bar (Upload, CLI, Source, user/admin links)
├── RootView.vue           — switches between Upload/Download based on query.id
│   ├── UploadView.vue     — file staging, settings, upload execution
│   │   ├── UploadSidebar  — upload settings (one-shot, stream, TTL, etc.)
│   │   ├── FileRow        — individual file display
│   │   └── CodeEditor     — text paste mode with syntax highlighting
│   └── DownloadView.vue   — file list, download links, admin actions
│       ├── DownloadSidebar — upload info, admin URL, action buttons
│       ├── FileRow         — download/QR/copy/remove per file + View button
│       ├── CodeEditor      — inline file viewer (read-only)
│       ├── QrCodeDialog    — QR code modal
│       ├── CopyButton      — clipboard copy with feedback
│       └── ConfirmDialog   — confirmation modal
├── LoginView.vue          — local login form + OAuth/OIDC buttons
├── HomeView.vue           — user dashboard (uploads/tokens/account)
│   └── CopyButton         — clipboard copy for tokens
├── AdminView.vue          — admin panel (stats/users/uploads)
└── ClientsView.vue        — CLI client downloads (from embedded build info)
```

---

## Authenticated Pages

### Auth State (`authStore.js`)

Reactive singleton holding `auth.user` (set on login, cleared on logout). Checked by `main.js` on app load via `GET /me`. The header shows user/admin links when `auth.user` is set.

### LoginView (`/#/login`)

- Local login form (username + password → `POST /auth/local/login`) — **hidden** when `config.localAuthentication` is `false` (i.e. `DisableLocalLogin = true` on the server)
- Conditional OAuth buttons (Google, OVH) based on `config.googleAuthentication` / `config.ovhAuthentication`
- OIDC button (label from `config.oidcProviderName`) → calls `GET /auth/oidc/login` to get the authorization URL, then `window.location.href` redirects to the OIDC provider
- "or continue with" divider only shown when both local login and at least one OAuth/OIDC provider are enabled
- Redirects to `/#/home` on success

### HomeView (`/#/home`)

Sidebar + main content layout (same pattern as download view).

**Sidebar**: user avatar, login/provider, name, email, admin badge, stats (uploads/files/size). Buttons: Upload files, Uploads, Tokens, Sign out, Edit account, Delete uploads, Delete account.

**Uploads tab**: paginated user uploads via `GET /me/uploads`. Supports token-based filtering. Each upload shows files, date, size, with clickable token labels.

**Tokens tab**: list/create/revoke tokens via `GET|POST|DELETE /me/token`. Token comment displayed above UUID. Click token to filter uploads by it.

**Edit Account modal**: name, email, password (local only). Admin users additionally see maxFileSize, maxUserSize, maxTTL, admin toggle. Saves via `POST /user/{id}`.

> **Gotcha**: Non-admin users cannot change quota fields or admin status — the server enforces this; the UI hides those fields.

### AdminView (`/#/admin`)

Admin-only page. Redirects non-admins to `/` on mount.

**Sidebar**: server version/build info (release + mint badges), nav buttons (Stats, Uploads, Users), Create User button.

**Stats tab**: server config (maxFileSize, maxUserSize, defaultTTL, maxTTL) + server statistics (users, uploads, files, totalSize, anonymous counts).

**Users tab**: paginated user list via `GET /users`. Each row shows login, provider, name, email, quotas, admin badge. Actions: Impersonate (👤), Edit (opens modal with full quota controls), Delete (with confirmation). Delete disabled for self. Impersonate disabled for self.

**Uploads tab**: paginated all-uploads via `GET /uploads`. Sort by date/size, order asc/desc. Filter by user/token (clickable links in each row). Each row shows upload ID (link), dates, user, token, files with sizes, Remove button.

**Create User modal**: provider (select), login, password (local only), name, email, quotas (maxFileSize, maxUserSize, maxTTL), admin toggle. Creates via `POST /user`.

**Edit User modal**: same as HomeView edit but with full admin quota controls always visible.

### Impersonation

Allows an admin to "become" another user to browse their uploads, test their quotas, or manage their account. The feature spans four files:

**Flow:**
1. Admin clicks 👤 on a user row in AdminView
2. `authStore.impersonate(user)` stores the target user and calls `api.setImpersonateUser(userId)`
3. `api.js` injects `X-Plik-Impersonate: <userId>` header on **every** subsequent API request
4. Server middleware (`server/middleware/impersonate.go`) detects the header, verifies the caller is an admin, and switches the request context to the impersonated user
5. `GET /me` now returns the impersonated user — `authStore.user` updates accordingly
6. A yellow banner in `AppHeader.vue` shows "⚠️ Impersonating **username**" with a **Stop** button

**State management (`authStore.js`):**
- `auth.originalUser` — preserved real admin identity (never changes during impersonation)
- `auth.impersonatedUser` — the user object being impersonated (null when not impersonating)
- `auth.user` — switches to the impersonated user during impersonation
- `clearImpersonate()` — resets header, restores `auth.user` to `auth.originalUser`

**API layer (`api.js`):**
- `setImpersonateUser(userId)` — sets/clears a module-level `_impersonateUserId`
- `apiCall()` — if `_impersonateUserId` is set, adds `X-Plik-Impersonate` header


### API Endpoints (Auth/Admin)

| Endpoint              | Method | Purpose                        | Auth       |
|-----------------------|--------|--------------------------------|------------|
| `/auth/local/login`   | POST   | Local login                    | —          |
| `/auth/oidc/login`    | GET    | Get OIDC authorization URL     | —          |
| `/auth/oidc/callback` | GET    | OIDC callback (sets session)   | —          |
| `/auth/google/login`  | GET    | Get Google authorization URL   | —          |
| `/auth/ovh/login`     | GET    | Get OVH authorization URL      | —          |
| `/auth/logout`        | GET    | Logout                         | Session    |
| `/me`                 | GET    | Get current user               | Session    |
| `/me`                 | DELETE | Delete account                 | Session    |
| `/me/uploads`         | GET    | User uploads (paginated)       | Session    |
| `/me/uploads`         | DELETE | Delete all user uploads        | Session    |
| `/me/token`           | GET    | List tokens                    | Session    |
| `/me/token`           | POST   | Create token                   | Session    |
| `/me/token/{token}`   | DELETE | Revoke token                   | Session    |
| `/user/{id}`          | POST   | Update user                    | Session    |
| `/stats`              | GET    | Server statistics              | Admin only |
| `/users`              | GET    | List all users (paginated)     | Admin only |
| `/user`               | POST   | Create user                    | Admin only |
| `/user/{id}`          | DELETE | Delete user                    | Admin only |
| `/uploads`            | GET    | All uploads (paginated, filterable) | Admin only |

> **Gotcha**: XSRF token (from `plik-xsrf` cookie) must be sent as `X-XSRFToken` header on all mutating requests (POST, DELETE). This is handled automatically in `apiCall()`.

---

## Responsive Layout

The layout uses a **mobile-first stacking pattern**:

```
Mobile (<768px):     [Sidebar]     (full width, stacked on top)
                     [Main Content] (full width, below)

Desktop (≥768px):   [Sidebar | Main Content]  (side by side)
```

Key classes:
- Containers: `flex flex-col md:flex-row`
- Sidebars: `w-full md:w-72 md:shrink-0`
- Outer wrapper: `overflow-x-hidden` (prevents long URLs from causing horizontal scroll)
- FileRow: inner container uses `flex-wrap`, "Download" text hidden on mobile (`hidden md:inline`)

---

## CSS Custom Utilities

The `style.css` file defines custom utility classes via `@utility` (Tailwind v4 syntax):

| Utility           | Description                                      |
|--------------------|--------------------------------------------------|
| `glass-card`       | Semi-transparent card with backdrop blur          |
| `btn`              | Base button styles                                |
| `btn-primary`      | Accent-colored button (cyan)                      |
| `btn-success`      | Green button                                      |
| `btn-danger`       | Red button                                        |
| `btn-ghost`        | Transparent hover button                          |
| `toggle-switch`    | Toggle switch base                                |
| `toggle-dot`       | Toggle switch dot (animated)                      |
| `input-field`      | Styled text input                                 |
| `sidebar-section`  | Glass-card styled sidebar section                 |
| `file-row`         | Glass-card styled file row with hover effect      |

> **Gotcha**: These are `@utility` blocks, NOT traditional CSS classes or Tailwind `@apply`. They follow Tailwind v4's custom utility syntax and generate single utility classes.

---

## Code Editor & File Viewer

### CodeEditor Component

Reusable CodeMirror 6 wrapper (`CodeEditor.vue`) used in two contexts:

| Context | View | Mode | Purpose |
|---------|------|------|---------|
| Text paste | UploadView | Read-write | Paste/edit text before uploading as a file |
| File viewer | DownloadView | Read-only | Preview uploaded text files inline |

**Props**: `modelValue` (v-model), `filename` (drives syntax highlighting), `readonly`, `placeholder`

**Language switching**: Uses a `Compartment` to reconfigure the language extension dynamically when `filename` changes — no editor destruction/recreation needed, preserving cursor position and undo history.

**Content-based language detection**: Uses `highlight.js` (lazy-loaded via dynamic `import()` on first detection call) for accurate auto-detection of ~190 languages. Detection fires via a 1s debounce on content changes. In UploadView, auto-detection only updates the filename when it still matches the default `paste.*` pattern.

**JSON prettify / validate**: When the detected language is JSON, two action buttons appear in the editor header bar. **Validate** (`JSON.parse()` only) checks syntax and shows a brief green "Valid" flash on success or a dismissable red error banner on failure — it never changes the content. **Prettify** (`JSON.parse()` → `JSON.stringify(…, null, 2)`) validates *and* reformats the content with 2-space indentation. In read-only mode (DownloadView file viewer) prettify updates the displayed view only — it does not modify the file on the server.

### Text-File Detection

The `isTextFile()` utility in `utils.js` determines if a file can be viewed in the code editor based on:
1. **Size**: Max 5 MB (`MAX_VIEWABLE_SIZE`)
2. **MIME type**: `text/*` prefix or known application types (JSON, XML, YAML, etc.)
3. **Extension**: ~60 common text/code extensions

`FileRow.vue` uses this to conditionally show a "View" button on uploaded files in download mode.

---

## Build & Release Process

### Development

```bash
cd webapp && npm install && npm run dev    # Vite dev server on :5173, proxies API to :8080
cd server && go run . --config ./plikd.cfg # Go backend on :8080
```

Vite proxy is configured in `vite.config.js` — all `/api`, `/auth`, `/file`, `/stream`, `/config`, `/me`, etc. calls are forwarded to the Go backend.

### Production Build

```bash
make frontend   # cd webapp && npm ci && npm run build → webapp/dist/
make server     # cd server && go build → server/plikd
```

The Go server serves `webapp/dist/` via `http.FileServer`. Default config: `WebappDirectory = "../webapp/dist"`.

### Makefile Targets

| Target            | Purpose                                           |
|-------------------|---------------------------------------------------|
| `all`             | `clean clean-frontend frontend clients server`    |
| `frontend`        | `npm ci && npm run build` in `webapp/`            |
| `server`          | Build Go binary `server/plikd`                    |
| `client`          | Build Go CLI client `client/plik`                 |
| `clients`         | Cross-compile clients for all architectures       |
| `docker`          | Build Docker image `rootgg/plik:dev`              |
| `release`         | Create release archives via `releaser/release.sh` |
| `clean`           | Remove server/client binaries                     |
| `clean-frontend`  | Remove `webapp/dist/`                             |
| `clean-all`       | Clean everything including `node_modules`         |

### Build Info & Client Downloads

The server binary embeds a JSON blob (via `server/gen_build_info.sh`) containing a client list discovered from the `clients/` directory. The `ClientsView` page displays download links from this embedded build info.

For full details on the Docker multi-stage build and release packaging, see [releaser/ARCHITECTURE.md](../releaser/ARCHITECTURE.md).

---

## Common Pitfalls

1. **Don't call `resp.json()` then `resp.text()`** — the body stream can only be read once. Always read as text first.

2. **File IDs are server-assigned** — when adding files to existing uploads, don't pass a `fileId` in the URL. The server creates one.

3. **`uploadToken` must be in `X-UploadToken` header** — not in the request body or URL query for API calls.

4. **Filter out `removed`/`deleted`, show everything else** — use a blacklist, not a whitelist. Files with `missing` or `uploading` status must be shown (they represent ongoing uploads).

5. **Refreshing the page loses admin access** — tokens are in-memory only. The only way to regain access is to open the Admin URL again.

6. **Delete responses are plain text `"ok"`** — don't try to parse `.message` from them. Always `fetchUpload()` after mutations.

7. **One-shot files disappear after download** — their status changes server-side; re-fetching will show them as removed/missing.

8. **The Admin URL sidebar truncation uses `overflow-hidden` + `min-w-0`** — without this, long URLs push the entire mobile layout wider than the viewport.

9. **`generateRef()` is for local tracking only** — it creates monotonically increasing IDs that are never sent to the server.

10. **Vite dev server runs on port 5173/5174** — the Go backend runs on port 8080. During dev, Vite proxies API calls to the backend via `vite.config.js`.

11. **`webapp/dist/` is gitignored** — never commit build artifacts. The CI/Docker build produces them fresh.

