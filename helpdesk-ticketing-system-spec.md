# Internal Helpdesk & Ticketing System dengan Dashboard Analytics

## 1. Ringkasan Proyek

Sistem internal helpdesk & ticketing lengkap dengan **RBAC (Role-Based Access Control)** dan **dashboard analytics** untuk monitoring performa penyelesaian tiket. Project ini dirancang secara sengaja untuk **overlap ke 3 posisi lamaran sekaligus**: Human Capital (IT), System Administration, dan Data Analyst TREG V — karena benang merah kebutuhan skill di ketiganya adalah kombinasi *sistem internal + RBAC + dashboard/reporting*.

**Positioning penting:** ini BUKAN sekadar CRUD ticketing biasa. Fokus utama ada di tiga hal yang jadi pembeda:
1. **RBAC yang solid** (3 level: admin, staff/agent, end-user) — bukan cuma role field di database, tapi enforcement di level API dan UI
2. **SLA tracking otomatis** — auto-flag tiket yang lewat deadline berdasarkan kategori/prioritas
3. **Dashboard analytics** — statistik tiket, tren mingguan, performa staff, dengan query SQL yang cukup kompleks (agregasi, window function)

Project ini beda domain dari 2 project yang sudah ada (Seismic Monitor = data publik, Fleet Management = logistik) — menambah variasi ke arah **sistem internal perusahaan**, domain paling relevan untuk tiga lamaran ini.

**Deliverable akhir:**
1. Aplikasi web full-stack (repo GitHub, README lengkap, screenshot/demo)
2. Dashboard analytics dengan minimal 5 visualisasi
3. Dokumentasi teknis (technical documentation) + user guide/manual
4. Seed data realistis (dummy tickets, users, kategori) supaya dashboard punya data untuk dianalisis
5. (Opsional) Deploy live demo (Vercel/Railway/Render) supaya bisa ditunjukkan langsung ke interviewer

---

## 2. Kenapa Project Ini Menutupi 3 Posisi

| Fitur yang dibangun | Human Capital | System Administration | Data Analyst TREG V |
|---|---|---|---|
| Modul submit & tracking tiket | Web dev cycle penuh (FE-BE) | "Proses Dokumentasi Ticketing IT" + "Monitoring SLA" | — |
| RBAC (admin, staff, user) | "Mengatur hak akses siapa admin siapa staf" | Mirip "User Access Management" | — |
| Dashboard SLA & performa | Bagian testing/UAT | "Dashboard Ticket Statistic, Ticket Performance Report" | "Dashboard Development" + export laporan mingguan |
| Query & filter data tiket | — | — | Kebutuhan SQL & data processing |
| User manual & dokumentasi | "Go Live" + panduan pengguna | "Technical Documentation, User Guide" | — |

**Strategi framing saat interview** (project sama, sudut cerita beda):
- **Human Capital:** tonjolkan siklus development lengkap (requirement → design → build → UAT → go-live), RBAC sebagai representasi struktur organisasi, dan dokumentasi user guide.
- **System Administration:** tonjolkan ticketing flow, SLA monitoring, user access management, dan technical documentation.
- **Data Analyst TREG V:** tonjolkan dashboard analytics, query SQL kompleks, dan insight yang dihasilkan dari data tiket (pola beban kerja, bottleneck, tren).

---

## 3. Tech Stack

- **Backend:** Go + Fiber (REST API) — konsisten dengan stack Fleet Management System
- **Database:** PostgreSQL — **dijalankan lokal untuk development**, dipindah ke layanan hosted (Neon/Supabase) saat tahap deploy
- **DB Access:** `sqlx` (bukan full ORM) — query SQL analitis (window function, `FILTER`) ditulis eksplisit, relevan langsung untuk posisi Data Analyst
- **Frontend:** Next.js (App Router) + TypeScript + TailwindCSS, dengan desain **corporate/enterprise** (sidebar gelap, aksen biru korporat, tabel padat-informasi — bukan gaya SaaS playful)
- **Tabel/Data Grid:** TanStack Table (React Table v8) — headless, dipakai di semua tabel (tiket, users, categories) untuk sorting, search, dan pagination
- **Notifikasi UI:** react-hot-toast untuk feedback sukses/error di semua form
- **Auth:** JWT (access token), password hashing dengan bcrypt
- **Dashboard/Chart:** Recharts di frontend, agregasi data lewat SQL query di backend (bukan diolah di frontend) — supaya nunjukin kemampuan SQL
- **Testing:** unit test (business logic murni: SLA calculation, status transition) + integration test lewat HTTP request (`httptest` + Fiber `app.Test()`) untuk membuktikan RBAC benar-benar tertegak di endpoint
- **Notifikasi (opsional, nice-to-have):** email notification (SLA breach warning) pakai SMTP sederhana atau in-app notification
- **Deployment:** Backend di Railway/Render, Frontend di Vercel, DB dipindah dari lokal ke Neon atau Supabase (connection string tinggal diganti lewat environment variable, tidak ada perubahan kode)

---

## 4. Role & Hak Akses (RBAC)

| Role | Hak Akses |
|---|---|
| **Admin** | Full access: manage users, manage kategori tiket, manage SLA rules, lihat semua tiket & dashboard, assign/reassign tiket ke staff |
| **Staff/Agent** | Lihat tiket yang di-assign ke dia, update status tiket, tambah komentar/response, lihat dashboard performa dirinya sendiri |
| **End-user (Requester)** | Submit tiket baru, lihat status tiket miliknya sendiri, tambah komentar di tiketnya, tidak bisa lihat tiket orang lain |

Enforcement RBAC dilakukan di 2 layer:
- **Middleware backend** (cek role dari JWT claim sebelum request diproses ke handler)
- **UI conditional rendering** (frontend sembunyikan menu/aksi yang tidak relevan dengan role, tapi backend tetap jadi source of truth keamanan)

---

## 5. Fitur Inti (Functional Requirements)

### 5.1 Autentikasi & User Management
- Register (khusus dibuat admin untuk staff/agent; end-user bisa self-register)
- Login/logout dengan JWT
- Admin bisa CRUD user, assign role, nonaktifkan akun

### 5.2 Manajemen Tiket
- Submit tiket baru (judul, deskripsi, kategori, prioritas, lampiran opsional)
- Auto-assign atau manual assign ke staff oleh admin
- Update status: `Open → In Progress → Pending → Resolved → Closed`
- Komentar/thread di dalam tiket (komunikasi requester ↔ staff)
- Riwayat perubahan status (audit trail sederhana)

### 5.3 Kategori & Prioritas
- Admin bisa CRUD kategori tiket (misal: Hardware, Software, Network, Access Request)
- Prioritas: Low, Medium, High, Critical — masing-masing punya SLA target berbeda (jam/hari)

### 5.4 SLA Tracking
- Setiap tiket punya deadline otomatis berdasarkan kombinasi kategori + prioritas
- Auto-flag (badge/warna) kalau tiket mendekati atau sudah lewat deadline
- Job/cron sederhana (atau dihitung on-the-fly saat query) untuk update status "SLA Breached"

### 5.5 Dashboard Analytics
- Total tiket per status (pie/donut chart)
- Tren jumlah tiket masuk per minggu/bulan (line chart)
- Rata-rata waktu penyelesaian tiket (overall & per kategori)
- Ranking staff berdasarkan jumlah tiket selesai & rata-rata waktu penyelesaian
- Tabel tiket yang SLA breach (butuh perhatian)
- Filter dashboard: rentang tanggal, kategori, staff

### 5.6 Reporting (Pengembangan Lanjutan — di luar implementasi inti)
- Export laporan mingguan/bulanan ke CSV atau PDF sederhana (jumlah tiket, SLA compliance rate, dll)
- Tidak termasuk di guide implementasi inti (lihat `helpdesk-step-by-step-guide.md`) karena 4 endpoint dashboard (summary, trend, staff-performance, sla-breaches) sudah cukup untuk mendemokan kemampuan analitis; export CSV bisa ditambahkan belakangan sebagai enhancement

### 5.7 Dokumentasi
- README teknis (setup, arsitektur, ERD, API docs)
- User guide (cara submit tiket, cara staff handle tiket, cara admin manage user) — bisa dalam bentuk PDF/markdown dengan screenshot

---

## 6. ERD (Entity Relationship Diagram)

```
┌─────────────────┐       ┌──────────────────┐       ┌─────────────────┐
│      users       │       │     tickets       │       │   categories    │
├─────────────────┤       ├──────────────────┤       ├─────────────────┤
│ id (PK)          │       │ id (PK)           │       │ id (PK)         │
│ name             │       │ title             │       │ name            │
│ email (unique)   │       │ description       │       │ default_sla_hrs │
│ password_hash    │◄─────┤ requester_id (FK) │       └─────────────────┘
│ role             │  1:N │ assignee_id (FK)  │──────►         ▲
│ is_active        │      │ category_id (FK)  │───────────────┘
│ created_at       │      │ priority          │
└─────────────────┘      │ status            │       ┌─────────────────┐
        ▲                 │ sla_deadline      │       │ ticket_comments │
        │                 │ resolved_at       │       ├─────────────────┤
        │                 │ created_at        │       │ id (PK)         │
        │                 │ updated_at        │◄─────┤ ticket_id (FK)  │
        │                 └──────────────────┘  1:N  │ user_id (FK)    │──┐
        │                          ▲                  │ message         │  │
        │                          │                  │ created_at      │  │
        │                          │                  └─────────────────┘  │
        │                          │                                       │
        │                 ┌──────────────────┐                            │
        └─────────────────┤ ticket_status_log│                            │
                  1:N      ├──────────────────┤                            │
                           │ id (PK)          │                            │
                           │ ticket_id (FK)   │                            │
                           │ old_status       │                            │
                           │ new_status       │                            │
                           │ changed_by (FK)  │◄───────────────────────────┘
                           │ changed_at       │
                           └──────────────────┘
```

**Tabel `sla_rules`** (aktif dipakai — bukan opsional, jadi sumber SLA spesifik per kombinasi kategori+prioritas sebelum fallback ke `default_sla_hours` kategori):
```
sla_rules (id, category_id FK, priority, resolution_hours, UNIQUE(category_id, priority))
```

**Tabel tambahan (opsional untuk laporan/export):**
```
attachments (id, ticket_id FK, file_url, uploaded_at)
```

---

## 7. API Endpoints

### Auth
```
POST   /api/auth/register        # self-register, role SELALU dipaksa 'end_user' di backend
POST   /api/auth/login            # mengembalikan JWT access token (expiry 2 jam)
GET    /api/auth/me               # profil user yang sedang login (butuh token)
```

> Catatan: versi awal tidak memakai refresh token terpisah — access token JWT berlaku 2 jam, cukup untuk sesi kerja harian dan menyederhanakan implementasi. Refresh token bisa ditambahkan sebagai pengembangan lanjutan.

### Users (admin only)
```
GET    /api/users                 # list semua user
PATCH  /api/users/:id             # update role & is_active (nonaktifkan = soft-delete, bukan DELETE)
```

> User baru (staff) dibuat lewat endpoint register publik lalu role-nya diubah admin ke `staff` via PATCH — bukan endpoint `POST /users` terpisah, supaya alur pembuatan user konsisten satu jalur.

### Categories & SLA Rules
```
GET    /api/categories                   # semua role bisa lihat
POST   /api/categories                   # admin
PATCH  /api/categories/:id               # admin
DELETE /api/categories/:id               # admin
POST   /api/sla-rules                    # admin — upsert aturan SLA per kombinasi kategori+prioritas
```

### Tickets
```
GET    /api/tickets                     # list (terfilter otomatis by role: admin=all, staff=assigned, end_user=own)
GET    /api/tickets/:id                 # detail tiket + daftar komentar
POST   /api/tickets                     # submit tiket baru (semua role terautentikasi)
PATCH  /api/tickets/:id/assign          # assign/reassign ke staff (admin only)
PATCH  /api/tickets/:id/status          # update status (staff & admin only, validasi transisi, trigger log)
POST   /api/tickets/:id/comments        # tambah komentar (semua role terautentikasi, dibatasi ke tiketnya sendiri untuk end_user)
```

### Dashboard (admin only)
```
GET    /api/dashboard/summary            # total per status + SLA compliance rate
GET    /api/dashboard/trend              # tren tiket masuk mingguan
GET    /api/dashboard/staff-performance  # ranking staff (window function RANK())
GET    /api/dashboard/sla-breaches       # daftar tiket yang sudah lewat SLA
```

> Endpoint export CSV (`/api/reports/export`) termasuk pengembangan lanjutan opsional — tidak ada di implementasi inti Section 11-13 pada guide, karena data yang sama sudah bisa dianalisis lewat 4 endpoint dashboard di atas.

---

## 8. Struktur Folder

```
helpdesk-ticketing-system/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   │   └── migrations/
│   │   ├── handlers/
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── ticket_handler.go
│   │   │   ├── category_handler.go
│   │   │   └── dashboard_handler.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go         # RequireAuth + RequireRole
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── ticket.go
│   │   │   └── category.go                # termasuk struct SLARule
│   │   ├── repository/
│   │   │   ├── user_repository.go
│   │   │   ├── category_repository.go
│   │   │   ├── ticket_repository.go       # termasuk method transaction-based (CreateWithLog, UpdateStatusWithLog)
│   │   │   └── dashboard_repository.go    # query SQL analitis (agregasi, window function)
│   │   ├── service/
│   │   │   ├── sla_service.go             # hitung deadline & SLAState (ok/at_risk/breached)
│   │   │   └── ticket_service.go          # validasi transisi status, ListForUser (RBAC di level data)
│   │   └── routes/
│   │       └── routes.go
│   ├── seed/
│   │   └── seed_data.go                   # 1 admin, 4 staff, 15 end-user, 180 tiket
│   ├── tests/
│   │   ├── rbac_integration_test.go       # integration test lewat HTTP (httptest + Fiber app.Test)
│   │   └── helpers_test.go
│   ├── go.mod
│   └── .env.example
│
├── frontend/
│   ├── app/
│   │   ├── layout.tsx                     # AuthProvider + Toaster
│   │   ├── (auth)/
│   │   │   ├── login/page.tsx
│   │   │   └── register/page.tsx
│   │   └── (dashboard)/
│   │       ├── layout.tsx                 # Sidebar + Topbar shell
│   │       ├── tickets/
│   │       │   ├── page.tsx               # list, dipakai bersama end_user & staff
│   │       │   ├── new/page.tsx
│   │       │   └── [id]/page.tsx
│   │       ├── analytics/page.tsx
│   │       └── admin/
│   │           ├── layout.tsx             # guard tambahan: redirect non-admin
│   │           ├── users/page.tsx
│   │           └── categories/page.tsx
│   ├── components/
│   │   ├── ui/
│   │   │   ├── DataTable.tsx              # TanStack Table generic (sort, search, pagination)
│   │   │   └── Skeleton.tsx               # TableSkeleton untuk loading state
│   │   ├── charts/
│   │   │   ├── StatusPieChart.tsx
│   │   │   └── TrendLineChart.tsx
│   │   ├── tickets/
│   │   │   ├── StatusBadge.tsx
│   │   │   └── SlaBadge.tsx
│   │   ├── layout/
│   │   │   ├── Sidebar.tsx
│   │   │   └── Topbar.tsx
│   │   └── rbac/
│   │       └── RoleGuard.tsx              # conditional render berdasarkan role
│   ├── lib/
│   │   ├── api.ts                         # axios instance + interceptor JWT & 401 handler
│   │   └── auth.tsx                       # AuthProvider, useAuth hook
│   ├── types/
│   └── package.json
│
├── docs/
│   ├── technical-documentation.md
│   ├── user-guide.md
│   └── screenshots/
│
└── README.md
```

---

## 9. Contoh Query SQL Analitis (untuk dashboard)

Contoh-contoh ini penting untuk showcase ke posisi Data Analyst — pakai window function & agregasi, bukan cuma `COUNT(*)`:

```sql
-- Rata-rata waktu penyelesaian per kategori
SELECT c.name AS category,
       ROUND(AVG(EXTRACT(EPOCH FROM (t.resolved_at - t.created_at)) / 3600)::numeric, 2) AS avg_resolution_hours,
       COUNT(*) AS total_resolved
FROM tickets t
JOIN categories c ON t.category_id = c.id
WHERE t.status = 'Closed'
GROUP BY c.name
ORDER BY avg_resolution_hours DESC;

-- Ranking staff berdasarkan tiket selesai (window function)
SELECT u.name AS staff,
       COUNT(*) AS tickets_resolved,
       RANK() OVER (ORDER BY COUNT(*) DESC) AS rank
FROM tickets t
JOIN users u ON t.assignee_id = u.id
WHERE t.status = 'Closed'
GROUP BY u.name;

-- Tren mingguan jumlah tiket masuk
SELECT DATE_TRUNC('week', created_at) AS week,
       COUNT(*) AS ticket_count
FROM tickets
GROUP BY 1
ORDER BY 1;

-- SLA compliance rate
SELECT
  COUNT(*) FILTER (WHERE resolved_at <= sla_deadline) * 100.0 / COUNT(*) AS sla_compliance_pct
FROM tickets
WHERE status = 'Closed';
```

---

## 10. Rencana Kerja (perkiraan: 4-5 hari kerja)

**Hari 1 — Backend Foundation:**
- Setup PostgreSQL lokal, project Go + Fiber, migrations
- Model & repository: users, categories, sla_rules, tickets, comments, status log
- Auth (register/login/JWT) + middleware RBAC (`RequireAuth`, `RequireRole`)

**Hari 2 — Backend Fitur Inti:**
- CRUD tiket, assign, update status transisi (dengan DB transaction) + status log
- SLA service (hitung deadline, `SLAState` untuk badge ok/at_risk/breached)
- Dashboard endpoints (query SQL analitis: `RANK()`, `FILTER`, agregasi)
- Seed data dummy (1 admin, 4 staff, 15 end-user, 180 tiket dengan variasi realistis)
- Unit test (business logic) + integration test (RBAC lewat HTTP request sungguhan)

**Hari 3 — Frontend Fondasi & Auth:**
- Setup Next.js + Tailwind dengan design token corporate/enterprise
- API client, auth context, layout sidebar/topbar, RoleGuard
- Komponen reusable: DataTable (TanStack Table), Skeleton loading, Toaster
- Halaman login & register

**Hari 4 — Frontend Fitur Inti:**
- List tiket (DataTable dengan sort/search/pagination), form submit tiket
- Detail tiket + thread komentar + panel update status (staff/admin)
- Halaman admin: manajemen user, kategori & SLA rules

**Hari 5 — Dashboard, Dokumentasi, Deploy:**
- Dashboard analytics (charts Recharts + tabel performa staff)
- Technical documentation + user guide (dengan screenshot UI corporate)
- Pindah database dari lokal ke Neon/Supabase, deploy backend & frontend, smoke test production

---

## 11. Yang Perlu Diceritakan di Portofolio/Interview

- Satu project, tiga sudut cerita — menunjukkan kemampuan melihat kebutuhan dari perspektif berbeda (HR/People, teknis infrastruktur, dan analitis data) dari sistem yang sama
- RBAC yang di-enforce di 2 layer (bukan sekadar kolom `role` di database) — **dibuktikan lewat integration test** yang mengecek response 401/403/200 sungguhan, bukan cuma diklaim
- Perubahan status & pembuatan tiket bersifat atomic lewat DB transaction — paham konsep data integrity, bukan sekadar CRUD
- SLA tracking otomatis — menunjukkan kemampuan menerjemahkan business rule (target waktu penyelesaian) jadi logic sistem, teruji lewat unit test
- Dashboard dengan query SQL analitis (window function, agregasi) — bukan cuma `COUNT(*)`, tapi ranking, rata-rata per kategori, compliance rate
- UI corporate/enterprise dengan tabel data yang scalable (TanStack Table — sort, search, pagination) — relevan langsung untuk sistem internal yang datanya akan terus bertambah
- Strategi database lokal-ke-cloud (PostgreSQL lokal saat development, dipindah ke Neon/Supabase saat deploy lewat environment variable) — menunjukkan pemahaman konfigurasi environment-based, bukan hardcode
- Dokumentasi lengkap (technical + user guide) — menunjukkan kesiapan kerja di lingkungan enterprise/BUMN yang biasanya menuntut dokumentasi rapi
- Domain baru (sistem internal perusahaan) yang melengkapi variasi portofolio dari 2 project sebelumnya (data publik & logistik)
