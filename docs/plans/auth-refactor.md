# Auth Refactor Plan — ropefish

Branch: `refactor-auth-phased`  
Status: Phase 1 — planning only (no code changes in this phase)

## Goals

1. **Single source of truth** for roles and route permissions (backend + admin UI derive from one matrix).
2. **Revocable sessions** — refresh tokens stored server-side; logout and role/disable take effect immediately.
3. **No stale client state** — remove the redundant `admin_user` cookie; user profile comes from `/v1/auth/me` or login/refresh response.
4. **One refresh path** — centralized token refresh in the admin BFF (no duplicate logic in hooks + `fetchWithAuth`).
5. **Fewer DB round-trips** — middleware validates JWT + session record instead of full user lookup on every request.
6. **Test coverage** — unit and integration tests for auth service, handlers, middleware, and admin BFF.

---

## Current State (baseline)

| Layer | Implementation | Pain |
|-------|----------------|------|
| Backend tokens | HS256 JWT access (15 min) + refresh (7 d), stateless | No revocation, no rotation |
| Backend middleware | `RequireAuth` → `GetUserByID` on every protected request | DB round-trip per request |
| Backend routes | `staff` (admin+manager), `adm` (admin only) | Duplicated with frontend guards |
| Backend logout | **Missing** | Client-only cookie delete |
| Admin cookies | `admin_access_token`, `admin_refresh_token`, `admin_user` | `admin_user` never updated on refresh |
| Admin guards | `ALLOWED_ROLES`, `ADMIN_ONLY_PATHS` in `hooks.server.ts` + layout | Drift from backend route groups |
| Admin refresh | `hooks.server.ts` (expiry) + `api.server.ts` (missing token, 401) | Scattered, inconsistent |
| Tests | `routes_auth_test.go` only (route-level mocks) | No service/handler/middleware/UI tests |

---

## Target Session Model

### Overview

Move from **stateless dual-JWT** to **JWT access token + server-side refresh session**.

```
┌─────────────┐     login      ┌──────────────┐
│  Admin BFF  │ ──────────────►│  POST /login │
│  (SvelteKit)│◄───────────────│              │
└──────┬──────┘  access+refresh└──────┬───────┘
       │  cookies (2)                 │ creates refresh_sessions row
       │                              ▼
       │                       ┌──────────────┐
       │  Bearer access        │ refresh_token│  opaque UUID (or JWT jti)
       └──────────────────────►│  in DB       │  linked to user_id, expires_at,
                               │              │  revoked_at, rotated_from
                               └──────────────┘
```

### Access token (JWT, unchanged algorithm)

- **Algorithm:** HS256  
- **TTL:** 15 minutes (configurable via `JWT_EXPIRATION_MINUTES`)  
- **Claims:** `user_id`, `email`, `role`, `type: "access"`, `sid` (session ID — links to refresh session row)  
- **Validation:** signature + expiry + `type == "access"` + session not revoked (lightweight lookup by `sid`, not full user fetch)

### Refresh token (opaque, server-validated)

- **Format:** cryptographically random UUID (or JWT with `jti` only — prefer opaque for simplicity)  
- **Storage:** `refresh_sessions` table (see migration below)  
- **TTL:** 7 days (configurable via `JWT_REFRESH_EXPIRATION_HOURS`)  
- **Rotation:** each `POST /auth/refresh` issues a new refresh token and **revokes** the previous one (`rotated_from` chain for audit)  
- **Revocation triggers:** logout, password change, role change, `is_active = false`, admin force-revoke (future)

### Middleware behavior (target)

`RequireAuth`:

1. Parse Bearer access JWT.  
2. Validate signature, expiry, `type == "access"`.  
3. Check `refresh_sessions` row by `sid`: exists, not revoked, not expired.  
4. Build `*models.User` from **JWT claims** (id, email, role) — no `GetUserByID` on hot path.  
5. Optional: async or periodic full user revalidation (out of scope for initial refactor; document as future enhancement).

`RequireRole(roles...)`: unchanged semantics — check `user.Role` from locals.

### Database: `refresh_sessions`

```sql
CREATE TABLE refresh_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,  -- SHA-256 of opaque refresh token
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    rotated_from UUID REFERENCES refresh_sessions(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_agent  TEXT,
    ip_address  INET
);
CREATE INDEX idx_refresh_sessions_user_id ON refresh_sessions(user_id);
CREATE INDEX idx_refresh_sessions_expires_at ON refresh_sessions(expires_at) WHERE revoked_at IS NULL;
```

Housekeeping: cron or startup job deletes expired/revoked rows older than 30 days.

---

## API Contract

Base path: `/v1/auth` (unchanged).

### POST `/v1/auth/login`

**Request:**
```json
{ "email": "user@example.com", "password": "..." }
```

**Success (200):**
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "user": {
    "id": "...",
    "email": "...",
    "full_name": "...",
    "role": "admin|manager",
    "is_active": true,
    "notify_leads": false,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

**Errors:** 400 (validation), 401 (bad credentials), 403 (disabled account).

**Side effects:** create `refresh_sessions` row; set `sid` in access JWT claims.

---

### POST `/v1/auth/refresh`

**Request:**
```json
{ "refresh_token": "<opaque>" }
```

**Success (200):** same shape as login (new access + new refresh + current `user` from DB).

**Errors:** 401 (invalid/expired/revoked refresh token, disabled user).

**Side effects:** revoke old session; create new session; rotate refresh token.

---

### GET `/v1/auth/me`

**Headers:** `Authorization: Bearer <access_token>`

**Success (200):**
```json
{
  "user": { /* full User object, fresh from DB or claims + notify_leads lookup */ }
}
```

**Errors:** 401 (missing/invalid token, revoked session).

**Note:** This remains the canonical source for `notify_leads` and post-login profile updates. Admin BFF should call `/me` once per request (or cache in `event.locals` for the request lifecycle) instead of parsing `admin_user` cookie.

---

### POST `/v1/auth/logout` *(new)*

**Headers:** `Authorization: Bearer <access_token>` (optional but recommended)

**Request:**
```json
{ "refresh_token": "<opaque>" }
```

**Success (204):** no body.

**Side effects:** revoke refresh session (by token hash or `sid` from access JWT). Idempotent — already-revoked token returns 204.

**Admin BFF:** call logout before clearing cookies so server-side session is invalidated.

---

## Role Matrix (Single Source of Truth)

Define once in **`internal/auth/permissions.go`** (new file) and expose read-only metadata for the admin UI via **`GET /v1/auth/permissions`** (optional in Phase 2; admin can import generated JSON in Phase 3).

### Roles

| Role | Constant | Description |
|------|----------|-------------|
| Admin | `models.RoleAdmin` (`"admin"`) | Full access |
| Manager | `models.RoleManager` (`"manager"`) | Content + leads; no users/settings/backups/AI models |

### API permission groups

| Group | Roles | Route prefix / endpoints |
|-------|-------|--------------------------|
| `public` | — | `/v1/leads`, `/v1/public/*`, `/v1/auth/login`, `/v1/auth/refresh` |
| `authenticated` | admin, manager | `/v1/auth/me`, `/v1/auth/logout` |
| `staff` | admin, manager | All content, uploads, files, leads, `GET /languages`, `POST /translate` |
| `admin` | admin only | `/users`, `/settings`, `/settings/*`, language CRUD, `/ai-models`, `/backups` |

### Admin UI route permissions

| UI path pattern | Required group | Backend equivalent |
|-----------------|----------------|-------------------|
| `/`, `/projects/**`, `/services/**`, `/stack/**`, `/seo/**`, `/pages/**`, `/legal/**`, `/files/**`, `/leads/**` | `staff` | `staff` group |
| `/settings/**` | `admin` | `adm` group |
| `/login`, `/logout` | — | public |

**Rule:** Admin UI must **not** maintain a parallel `ADMIN_ONLY_PATHS` list. Instead:

- Phase 2: generate `web/admin/src/lib/permissions.generated.ts` from `permissions.go` (build step or `go generate`).  
- Interim (Phase 2 UI): import constants from a hand-synced JSON until codegen lands.

### Matrix maintenance

- Adding a new admin-only API route → update `permissions.go` + `routes.go` (same PR).  
- Adding a new admin UI section → add path pattern to matrix; hooks derive guard from matrix.  
- CI check (Phase 3): test that every `adm` route in `routes.go` appears in matrix and vice versa.

---

## Cookie Policy (Admin BFF)

### Target: 2 cookies (remove `admin_user`)

| Cookie | Purpose | maxAge | httpOnly | secure | sameSite |
|--------|---------|--------|----------|--------|----------|
| `admin_access_token` | Bearer JWT for API calls | 15 min (match JWT TTL) | true | prod only | `lax` |
| `admin_refresh_token` | Opaque refresh for BFF refresh calls | 7 days | true | prod only | `lax` |

**Removed:** `admin_user` — user profile loaded via `/v1/auth/me` into `event.locals.user` during the handle hook (once per request).

### BFF session flow

1. **Login:** set 2 cookies; store user in `event.locals` from login response.  
2. **Handle hook (every request):**  
   - Read cookies.  
   - If access valid → optional lightweight expiry check; attach token to locals.  
   - If access expired/missing but refresh present → call centralized `ensureValidSession(event)` (single module).  
   - Call `GET /v1/auth/me` (via internal fetch) to populate `event.locals.user`.  
   - Apply route guard from permissions matrix (not hardcoded paths).  
3. **Logout:** `POST /v1/auth/logout` + delete both cookies.  
4. **`fetchWithAuth`:** only attaches Bearer from locals; **does not** refresh (handle hook guarantees valid access token before loaders run).

### Security notes

- Refresh token never exposed to client JS (httpOnly).  
- `secure: true` when `url.protocol === 'https:'` (keep current behavior).  
- SameSite=Lax sufficient for same-site admin panel.  
- Cookie `maxAge` for access token should match JWT TTL to avoid holding expired tokens.

---

## Phased Migration

### Phase 2 — Backend

**Scope:** session storage, token rotation, logout, middleware optimization, permissions matrix, `/auth/permissions` (optional).

**Steps:**

1. Add migration `00N_refresh_sessions.sql`.  
2. Add `internal/repository/sessions.go` (create, get by hash, revoke, rotate, purge expired).  
3. Refactor `internal/services/auth/auth.go`:  
   - Access JWT includes `sid`.  
   - Generate opaque refresh token; store hash in DB.  
   - Validate refresh via DB lookup.  
4. Update `internal/handlers/auth.go`: login/refresh/logout; refresh returns fresh `user`.  
5. Update `internal/middleware/middleware.go`: session check by `sid`; build user from claims.  
6. Add `internal/auth/permissions.go` + wire `routes.go` groups to matrix constants.  
7. Revoke all sessions for user on: deactivate, role change, password update (`internal/handlers/users.go`).  
8. Extend `internal/server/routes_auth_test.go` + add service/handler tests.

**Backward compatibility:** During rollout, accept old stateless refresh JWTs for one release with deprecation log (optional flag `AUTH_LEGACY_REFRESH=true`); remove in following release.

**Files touched:**

| File | Change |
|------|--------|
| `migrations/00N_refresh_sessions.sql` | **new** |
| `internal/auth/permissions.go` | **new** |
| `internal/services/auth/auth.go` | session-aware tokens |
| `internal/services/auth/auth_test.go` | **new** |
| `internal/repository/sessions.go` | **new** |
| `internal/repository/sessions_test.go` | **new** |
| `internal/handlers/auth.go` | logout, rotation |
| `internal/handlers/auth_test.go` | **new** |
| `internal/handlers/users.go` | revoke sessions on user update |
| `internal/middleware/middleware.go` | sid check, no GetUserByID |
| `internal/middleware/middleware_test.go` | **new** |
| `internal/middleware/user_lookup.go` | simplify or remove if unused |
| `internal/server/routes.go` | register logout, use permissions |
| `internal/server/routes_auth_test.go` | extend |
| `internal/config/config.go` | session TTL env vars if needed |
| `internal/models/models.go` | session model if needed |
| `cmd/piplos/main.go` | wire session repo |

**Acceptance criteria — Phase 2:**

- [ ] Login creates DB session; access JWT contains `sid`.  
- [ ] Refresh rotates token; old refresh token returns 401.  
- [ ] Logout revokes session; subsequent refresh with old token fails.  
- [ ] Deactivated user cannot refresh or use access token (session revoked on deactivate).  
- [ ] Role change revokes all user sessions.  
- [ ] Protected routes do **not** call `GetUserByID` in middleware (verify via test spy or query log).  
- [ ] `permissions.go` documents all staff/admin routes; matches `routes.go`.  
- [ ] All existing `routes_auth_test.go` tests pass; new tests cover login/refresh/logout/rotation.  
- [ ] No breaking change to login/refresh JSON shape (fields unchanged).

---

### Phase 3 — Admin UI (BFF)

**Scope:** 2-cookie model, centralized refresh, permissions-driven guards, logout API call.

**Steps:**

1. Remove `COOKIE_USER` / `admin_user` from `auth.server.ts`.  
2. Implement `ensureValidSession(event)` in `api.server.ts` (sole refresh entry point).  
3. Refactor `hooks.server.ts`:  
   - Call `ensureValidSession` then `/v1/auth/me` to fill `event.locals.user`.  
   - Replace `ADMIN_ONLY_PATHS` / `ALLOWED_ROLES` with permissions module.  
4. Update `login/+page.server.ts`: set 2 cookies only; access cookie maxAge = 15 min.  
5. Update `logout/+page.server.ts`: POST `/v1/auth/logout` then clear cookies.  
6. Simplify `fetchWithAuth`: attach Bearer only; remove 401-refresh-retry (handle hook covers freshness).  
7. Update `(auth)/+layout.server.ts`: remove redundant `/me` if hooks already populate user (keep if needed for layout-specific data).  
8. Update `(auth)/+layout.svelte`: use permissions helper for nav visibility.  
9. Add `permissions.generated.ts` or hand-synced `permissions.ts` from backend matrix.  
10. Update `app.d.ts`, `types.ts` for locals shape.

**Files touched:**

| File | Change |
|------|--------|
| `web/admin/src/lib/auth.server.ts` | remove user cookie; cookie maxAge |
| `web/admin/src/lib/api.server.ts` | centralized session; simplify fetchWithAuth |
| `web/admin/src/lib/permissions.ts` | **new** (from matrix) |
| `web/admin/src/hooks.server.ts` | me + permissions guard |
| `web/admin/src/routes/login/+page.server.ts` | 2 cookies |
| `web/admin/src/routes/logout/+page.server.ts` | backend logout |
| `web/admin/src/routes/(auth)/+layout.server.ts` | use locals.user |
| `web/admin/src/routes/(auth)/+layout.svelte` | permissions-based nav |
| `web/admin/src/app.d.ts` | locals types |
| `web/admin/src/lib/types.ts` | AdminUser type |
| `web/admin/src/routes/(auth)/settings/users/+page.svelte` | use permissions helper |

**Consumers (verify only, no logic change unless cookie/API shape breaks):**

- `web/admin/src/lib/{toggle,settings,seo,lists,languages}.server.ts`  
- All `web/admin/src/routes/(auth)/**/*.{server.ts,ts}` using `fetchWithAuth` (~25 files)

**Acceptance criteria — Phase 3:**

- [ ] Only 2 cookies set after login; `admin_user` absent.  
- [ ] After refresh, `event.locals.user` reflects current DB state (role/name changes visible without re-login).  
- [ ] Manager blocked from `/settings/**`; admin allowed — matches backend 403 behavior.  
- [ ] Logout clears cookies and invalidates server session (refresh fails after logout).  
- [ ] Single code path performs token refresh (grep confirms no refresh outside `ensureValidSession`).  
- [ ] Access token expiry mid-session: next navigation refreshes transparently.  
- [ ] `fetchWithAuth` no longer contains refresh/retry logic.  
- [ ] Manual smoke: login → edit content → wait 16 min → navigate → still authenticated.

---

### Phase 4 — Tests & CI

**Scope:** fill coverage gaps; add drift-prevention checks.

**Steps:**

1. Backend integration test: full login → me → refresh → logout flow with test DB.  
2. Backend: middleware session-revoked scenarios.  
3. Admin: unit tests for `ensureValidSession`, hooks guard logic (vitest + mocked fetch/cookies).  
4. Optional e2e: Playwright login/logout/role redirect.  
5. CI script: compare `permissions.go` routes vs `routes.go` admin/staff groups.

**Files touched:**

| File | Change |
|------|--------|
| `internal/services/auth/auth_test.go` | expand |
| `internal/handlers/auth_test.go` | expand |
| `internal/middleware/middleware_test.go` | expand |
| `internal/server/routes_auth_test.go` | integration-style tests |
| `internal/auth/permissions_test.go` | **new** — matrix/route parity |
| `web/admin/src/lib/api.server.test.ts` | **new** |
| `web/admin/src/hooks.server.test.ts` | **new** |
| `scripts/check-auth-permissions.sh` | **new** (optional) |

**Acceptance criteria — Phase 4:**

- [ ] Auth service test coverage ≥ 80% for `internal/services/auth`.  
- [ ] Handler tests cover login failure modes (bad password, inactive, missing fields).  
- [ ] Middleware tests cover revoked session, wrong role, missing Bearer.  
- [ ] Admin unit tests cover hooks redirect for manager on `/settings` and expired-token refresh.  
- [ ] CI fails if permissions matrix drifts from `routes.go`.  
- [ ] All tests pass in CI without flaky timing issues.

---

## Rollout & Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Active sessions invalidated on deploy | Accept forced re-login once; communicate in deploy notes |
| Legacy refresh JWTs in flight | Optional `AUTH_LEGACY_REFRESH` grace period (Phase 2) |
| Extra `/me` call per admin request | One fetch in handle hook; cheaper than current DB lookup on every API call from backend |
| Cookie maxAge mismatch | Set access cookie maxAge = JWT TTL exactly |

**Deploy order:** Phase 2 (backend) → deploy → Phase 3 (admin UI) → deploy → Phase 4 (tests/CI, can overlap with Phase 3).

---

## Out of Scope (document for future)

- RS256 / asymmetric JWT  
- Multi-device session management UI  
- OAuth / SSO  
- Rate limiting on login/refresh  
- Periodic background user revalidation in middleware  
- Public site auth (unchanged)

---

## Summary Checklist

| Deliverable | Location |
|-------------|----------|
| Target session model | This doc § Target Session Model |
| API contract | This doc § API Contract |
| Role matrix | This doc § Role Matrix + `internal/auth/permissions.go` (Phase 2) |
| Cookie policy | This doc § Cookie Policy |
| Phase file lists | This doc § Phased Migration |
| Acceptance criteria | Per-phase sections above |
