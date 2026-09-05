# Technical Specification
## Internal Helpdesk & Ticketing System dengan Dashboard Analytics

Dokumen ini adalah spesifikasi teknis final untuk implementasi sistem internal siap produksi v1. Tidak ada override di akhir dokumen; semua keputusan final sudah ditulis langsung pada bagian terkait.

---

## 1. Ringkasan Sistem

Sistem adalah aplikasi web internal untuk ticketing helpdesk dengan RBAC, SLA tracking, analytics dashboard, dan export CSV.

Positioning teknis: sistem ini adalah project portofolio profesional dengan target production-ready v1 untuk sistem internal skala kecil/menengah. Sistem belum diklaim enterprise-grade penuh karena tidak mencakup SSO/SAML, HA multi-region, SIEM integration, distributed tracing penuh, compliance workflow formal, dan disaster recovery dengan RPO/RTO formal.

Fokus teknis utama:

- RBAC enforced di backend dan dibantu conditional UI di frontend
- SLA deadline dan SLA state dihitung dari kategori + prioritas
- Dashboard memakai query SQL agregasi dan window function
- Refresh token aman dengan penyimpanan hash di database
- Export CSV filterable untuk laporan periodik
- Audit trail untuk status change dan assignment/reassignment
- Production security baseline: secure refresh token cookie, CSRF protection, rate limiting, structured logging, and backup/restore procedure

---

## 1.1 Feature Traceability Notes

Catatan ini menjaga klaim fitur tetap akurat saat implementasi dan review.

- Fitur yang sudah eksplisit: register `end_user`, JWT login, refresh token hash di database, logout revoke, RBAC backend, ticket workflow, comments, category/SLA rules, SLA state 20%, dashboard admin/staff, filter dashboard, CSV export, skeleton/toast, dan admin user/category management.
- Audit trail dipisah: status change dicatat di `ticket_status_log`; assignment/reassignment dicatat di `ticket_activity_log`.
- Refresh token rotation belum menjadi requirement eksplisit. Spec saat ini mewajibkan refresh token hash, expiry, revoke, refresh endpoint, logout, httpOnly cookie production, dan CSRF. Jika rotation dijadikan wajib, tambahkan aturan bahwa refresh sukses menerbitkan refresh token baru dan mencabut token lama.
- TanStack Table tercatat sebagai tech stack tabel. Jika ingin klaim semua tabel memakai TanStack, guide dan implementasi UI harus konsisten memakai TanStack untuk tiket, user, dan kategori.

---

## 2. Tech Stack

- Backend: Go 1.22+ + Gin
- Database: PostgreSQL
- DB access: sqlx
- Auth: JWT access token + refresh token
- Password hashing: bcrypt
- Frontend: Next.js App Router + TypeScript
- Styling: TailwindCSS, corporate/enterprise visual style
- Table: TanStack Table
- Charts: Recharts
- Notification: react-hot-toast
- Testing backend: Go unit test + Gin + `net/http/httptest` integration test
- Deployment target: containerized backend on Railway/Render/Fly.io, frontend on Vercel, managed PostgreSQL on Neon/Supabase
- Logging: structured JSON logs
- Rate limiting: auth and export endpoints
- Runtime: environment-based configuration with no hardcoded secrets

---

## 3. RBAC

| Role | Access |
|---|---|
| `super_admin` | Full system authority, including every admin capability |
| `admin` | Manage users, categories, SLA rules, all tickets, assignment, full dashboard, export CSV |
| `staff` | Assigned tickets, status update, comments, own performance dashboard |
| `end_user` | Register, submit tickets, own tickets, comments on own tickets |

Backend is the source of truth. UI role checks are UX only.

Role assignment policy:

| Action | Allowed actor |
|---|---|
| Public register creates `end_user` | Anyone |
| Promote/demote user to `staff` or `end_user` | `super_admin`, `admin` |
| Promote/demote user to `admin` or `super_admin` | `super_admin` only |
| Create the first `super_admin` | Internal bootstrap command |

---

## 4. Database Schema

### 4.1 Tables

```sql
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'staff', 'end_user')),
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  default_sla_hours INT NOT NULL CHECK (default_sla_hours > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sla_rules (
  id BIGSERIAL PRIMARY KEY,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  priority TEXT NOT NULL CHECK (priority IN ('Low', 'Medium', 'High', 'Critical')),
  resolution_hours INT NOT NULL CHECK (resolution_hours > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (category_id, priority)
);

CREATE TABLE tickets (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  requester_id BIGINT NOT NULL REFERENCES users(id),
  assignee_id BIGINT NULL REFERENCES users(id),
  category_id BIGINT NOT NULL REFERENCES categories(id),
  priority TEXT NOT NULL CHECK (priority IN ('Low', 'Medium', 'High', 'Critical')),
  status TEXT NOT NULL CHECK (status IN ('Open', 'In Progress', 'Pending', 'Resolved', 'Closed')),
  sla_deadline TIMESTAMPTZ NULL,
  resolved_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_comments (
  id BIGSERIAL PRIMARY KEY,
  ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id),
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_status_log (
  id BIGSERIAL PRIMARY KEY,
  ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  old_status TEXT NULL,
  new_status TEXT NOT NULL,
  changed_by BIGINT NOT NULL REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_activity_log (
  id BIGSERIAL PRIMARY KEY,
  ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id BIGINT NOT NULL REFERENCES users(id),
  action TEXT NOT NULL CHECK (action IN ('ticket_created', 'status_changed', 'assigned', 'reassigned', 'comment_added')),
  old_value TEXT NULL,
  new_value TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 4.2 Indexes

```sql
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_tickets_requester_id ON tickets(requester_id);
CREATE INDEX idx_tickets_assignee_id ON tickets(assignee_id);
CREATE INDEX idx_tickets_category_id ON tickets(category_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_created_at ON tickets(created_at);
CREATE INDEX idx_tickets_sla_deadline ON tickets(sla_deadline);
CREATE INDEX idx_ticket_comments_ticket_id ON ticket_comments(ticket_id);
CREATE INDEX idx_ticket_status_log_ticket_id ON ticket_status_log(ticket_id);
CREATE INDEX idx_ticket_activity_log_ticket_id ON ticket_activity_log(ticket_id);
```

---

## 5. API Contract

### 5.1 Auth

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET  /api/auth/me
```

`POST /api/auth/register`

Request:

```json
{ "name": "User", "email": "user@example.com", "password": "password123" }
```

Behavior:

- Public endpoint
- Always creates role `end_user`
- Ignores any role sent by client
- First `super_admin` is created through an internal bootstrap command, not public register.

`POST /api/auth/login`

Response:

```json
{
  "access_token": "jwt",
  "user": { "id": 1, "name": "Admin", "email": "admin@example.com", "role": "admin", "is_active": true }
}
```

Rules:

- Access token expiry target: 15 minutes
- Refresh token is cryptographically random
- Store only refresh token hash in DB
- Return refresh token via secure, httpOnly, sameSite cookie in production
- Response body must not expose refresh token in production
- CSRF token/header is required for refresh and logout when cookies are used

`POST /api/auth/refresh`

Response:

```json
{ "access_token": "jwt" }
```

Production request uses refresh token from httpOnly cookie and CSRF header. Local development may allow JSON body only behind an explicit `AUTH_DEV_TOKEN_BODY=true` flag.

`POST /api/auth/logout`

Behavior:

- Sets `revoked_at`
- Returns success even if token is already revoked
- Clears refresh token cookie

### 5.2 Users

```http
GET   /api/users
PATCH /api/users/:id
```

Super admin/admin only.

`PATCH /api/users/:id` updates `role` and `is_active`.

Only `super_admin` can assign `admin` or `super_admin`. `admin` can assign only `staff` or `end_user`.

### 5.3 Categories & SLA Rules

```http
GET    /api/categories
POST   /api/categories
PATCH  /api/categories/:id
DELETE /api/categories/:id
POST   /api/sla-rules
```

- `GET /api/categories`: authenticated users
- Mutations: super admin/admin only
- `POST /api/sla-rules` upserts by `(category_id, priority)`

### 5.4 Tickets

```http
GET   /api/tickets
GET   /api/tickets/:id
POST  /api/tickets
PATCH /api/tickets/:id/assign
PATCH /api/tickets/:id/status
POST  /api/tickets/:id/comments
```

Role-scoped ticket list:

- Super admin/admin: all tickets
- Staff: assigned tickets
- End-user: own tickets

`GET /api/tickets/:id` includes comments:

```json
{
  "ticket": {},
  "comments": [
    {
      "id": 1,
      "ticket_id": 10,
      "user_id": 3,
      "author_name": "Budi Santoso",
      "author_role": "staff",
      "message": "Sudah dicek.",
      "created_at": "2026-07-23T12:00:00Z"
    }
  ]
}
```

Assignment:

- `PATCH /api/tickets/:id/assign` is super admin/admin only
- Updates `assignee_id`
- Inserts `ticket_activity_log` with `assigned` or `reassigned`
- Runs in one DB transaction

Status update:

- Staff/admin/super admin only
- Validates status transition
- Updates ticket and inserts `ticket_status_log` in one DB transaction

### 5.5 Dashboard

Admin endpoints:

```http
GET /api/dashboard/summary?from=&to=&category_id=&staff_id=
GET /api/dashboard/trend?from=&to=&category_id=&staff_id=
GET /api/dashboard/staff-performance?from=&to=&category_id=&staff_id=
GET /api/dashboard/sla-breaches?from=&to=&category_id=&staff_id=
```

Staff endpoint:

```http
GET /api/dashboard/my-performance?from=&to=&category_id=
```

Filter rules:

- `from`: ISO date `YYYY-MM-DD`, filters `tickets.created_at >= from`
- `to`: ISO date `YYYY-MM-DD`, filters `tickets.created_at < to + 1 day`
- `category_id`: filters `tickets.category_id`
- `staff_id`: super admin/admin endpoints only, filters `tickets.assignee_id`
- Staff endpoint ignores any client-provided staff id and scopes by JWT user id

`GET /api/dashboard/my-performance` response:

```json
{
  "assigned_total": 42,
  "resolved_total": 28,
  "avg_resolution_hours": 6.5,
  "sla_compliance_pct": 91.2,
  "status_summary": [{ "status": "Open", "count": 3 }],
  "sla_breaches": [{ "id": 12, "title": "VPN bermasalah", "priority": "High", "sla_deadline": "2026-07-23T10:00:00Z" }]
}
```

### 5.6 Reports

```http
GET /api/reports/export?from=&to=&category_id=&staff_id=
```

Rules:

- Super admin/admin only
- Uses same filters as admin dashboard
- Response headers:
  - `Content-Type: text/csv`
  - `Content-Disposition: attachment; filename="helpdesk-report.csv"`

CSV columns:

```csv
ticket_id,title,category,priority,status,requester,assignee,created_at,resolved_at,sla_deadline,sla_state
```

---

## 6. Business Logic

### 6.1 SLA Deadline

Resolution hours lookup:

1. Find `sla_rules` by `category_id + priority`
2. If not found, use `categories.default_sla_hours`

Formula:

```text
sla_deadline = created_at + resolution_hours
```

### 6.2 SLA State

```text
breached = now > sla_deadline AND status NOT IN ('Resolved', 'Closed')
at_risk = remaining_time <= total_sla_duration * 0.2
ok = all other cases
```

Ticket query should expose `sla_total_hours` where the UI needs an SLA badge.

### 6.3 Status Transitions

Allowed:

- `Open -> In Progress`
- `In Progress -> Pending`
- `In Progress -> Resolved`
- `Pending -> In Progress`
- `Pending -> Resolved`
- `Resolved -> Closed`
- `Resolved -> Open` for reopen
- `Closed -> Open` for reopen

Invalid transitions return `400`.

---

## 7. Backend Structure

```text
backend/
  cmd/server/main.go
  internal/config/
  internal/database/migrations/001_init.sql
  internal/handlers/
    auth_handler.go
    user_handler.go
    category_handler.go
    ticket_handler.go
    dashboard_handler.go
    report_handler.go
  internal/middleware/
    auth_middleware.go
  internal/dto/
    auth.go
    user.go
    category.go
    ticket.go
  internal/models/
    user.go
    category.go
    ticket.go
    audit_log.go
  internal/repository/
    user_repository.go
    refresh_token_repository.go
    category_repository.go
    ticket_repository.go
    dashboard_repository.go
    report_repository.go
  internal/service/
    auth_service.go
    token_service.go
    sla_service.go
    ticket_service.go
  internal/routes/routes.go
  seed/seed_data.go
  tests/
```

Layering:

```text
handler -> service -> repository -> database
```

Handlers bind request payloads through `internal/dto` and map them to `internal/models` before calling services/repositories. Handlers do not contain SQL. Repositories do not contain business policy.

---

## 8. Production Readiness

Security:

- Refresh token is stored in a secure, httpOnly, sameSite cookie in production.
- Refresh token DB value is hashed with SHA-256 plus server-side pepper.
- CSRF protection is required for cookie-backed `refresh` and `logout`.
- Access token is kept in memory by frontend; do not persist it in localStorage in production.
- CORS allowlist must use `FRONTEND_URL`.
- Password hash uses bcrypt default cost or higher.

Rate limiting:

- `POST /api/auth/login`: 5 attempts per 15 minutes per IP/email pair.
- `POST /api/auth/register`: 10 attempts per hour per IP.
- `POST /api/auth/refresh`: 30 attempts per 15 minutes per IP/user.
- `GET /api/reports/export`: 10 requests per 10 minutes per admin.

Operational endpoints:

```http
GET /api/health      # process is alive
GET /api/ready       # database reachable and migrations available
```

Logging:

- Use structured JSON logs.
- Include request id, method, path, status, latency, user id when authenticated, and role.
- Do not log password, access token, refresh token, or CSV contents.

Database operations:

- Migrations are versioned and run before deployment.
- Production migration must be backward compatible where possible.
- Backup procedure: daily PostgreSQL backup.
- Restore procedure must be documented and tested at least once before demo/deploy.

Configuration:

- All secrets come from environment variables.
- Application must fail fast if required secrets are missing in production.

---

## 9. Frontend Structure

```text
frontend/
  app/
    layout.tsx
    (auth)/login/page.tsx
    (auth)/register/page.tsx
    (dashboard)/layout.tsx
    (dashboard)/tickets/page.tsx
    (dashboard)/tickets/new/page.tsx
    (dashboard)/tickets/[id]/page.tsx
    (dashboard)/analytics/page.tsx
    (dashboard)/admin/layout.tsx
    (dashboard)/admin/users/page.tsx
    (dashboard)/admin/categories/page.tsx
  components/
    charts/StatusPieChart.tsx
    charts/TrendLineChart.tsx
    layout/Sidebar.tsx
    layout/Topbar.tsx
    rbac/RoleGuard.tsx
    tickets/StatusBadge.tsx
    tickets/SlaBadge.tsx
    ui/DataTable.tsx
    ui/Skeleton.tsx
  lib/api.ts
  lib/auth.tsx
  types/
```

Frontend requirements:

- Auth context stores user, access token, and refresh token.
- API interceptor retries one refresh on `401`.
- Sidebar shows analytics to admin and staff.
- Admin analytics shows full dashboard, staff ranking, filters, and export CSV.
- Staff analytics shows only `/api/dashboard/my-performance`.
- Ticket detail admin includes assign/reassign control.
- Ticket comments show author name and role.

---

## 10. Testing Requirements

Unit tests:

- SLA deadline fallback rule
- SLA state `ok`, `at_risk`, `breached`
- Status transition validation

Integration tests:

- Register always creates `end_user`
- Login returns token pair
- Refresh returns new access token
- Logout revokes refresh token
- Admin-only endpoint accepts super admin/admin and rejects staff/end-user
- Staff dashboard endpoint works for staff
- Staff cannot access admin dashboard endpoints
- End-user cannot access dashboard endpoints
- Ticket list is role-scoped
- Assignment writes `ticket_activity_log`
- Comments include author name and role
- Export CSV super admin/admin only
- Dashboard filters affect query result
- Rate limit returns `429` after threshold
- Health endpoint returns process status
- Ready endpoint fails when DB is unavailable
- Production auth mode does not expose refresh token in JSON body

Manual smoke test:

1. Register end-user
2. Login admin, staff, end-user
3. Admin creates category and SLA rule
4. End-user submits ticket
5. Admin assigns ticket to staff
6. Staff updates status and comments
7. Admin checks dashboard with filters
8. Staff checks own dashboard
9. Admin downloads CSV

---

## 11. Seed Data

Seed data minimum:

- 1 super admin
- 1 admin
- 4 staff
- 15 end-users
- 4 categories: Hardware, Software, Network, Access Request
- SLA rules for Low, Medium, High, Critical
- 180 tickets with varied:
  - status
  - priority
  - category
  - requester
  - assignee
  - created_at
  - resolved_at
  - SLA compliance and breach

Dashboard must not show empty or uniform data.

---

## 12. Production Talking Points

- RBAC is enforced at API level and proven by tests.
- Refresh token cookie flow shows production-grade session handling.
- Assignment and status changes are auditable.
- SLA at-risk uses a business-driven 20% rule.
- Staff analytics uses scoped access by identity, not blanket role access.
- Dashboard queries use SQL aggregation and window functions.
- CSV export supports reporting workflow for management.
- Logging, rate limiting, backup/restore, and readiness checks make the project production-ready v1.

---

## 13. Why This Is Not Full Enterprise-Grade Yet

This project is production-ready v1, not full enterprise-grade. Missing enterprise capabilities:

- Enterprise SSO/SAML/OAuth integration.
- HRIS/Active Directory user lifecycle synchronization.
- High availability multi-region deployment.
- Automatic failover and zero-downtime disaster recovery.
- Formal RPO/RTO targets and DR runbooks.
- SIEM/security event integration.
- Distributed tracing with a full observability stack.
- Automated incident alerting and on-call workflow.
- Formal compliance audit workflow.
- Data retention and legal hold policy.
- Load testing and capacity planning evidence for large-scale production traffic.
- Granular dynamic permission matrix beyond the fixed `super_admin`, `admin`, `staff`, `end_user` roles.

These are intentionally kept out of production v1 so the project remains realistic for an individual portfolio while still demonstrating production-oriented engineering practices.
