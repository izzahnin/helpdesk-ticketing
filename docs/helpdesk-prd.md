# Product Requirement Document (PRD)
## Internal Helpdesk & Ticketing System dengan Dashboard Analytics

Dokumen ini menerjemahkan kebutuhan bisnis di `helpdesk-brd.md` menjadi kebutuhan produk final yang siap diimplementasikan sebagai sistem internal siap produksi v1. Scope final memasukkan refresh token, export CSV, dashboard staff scoped, filter dashboard, audit trail, dan production security baseline sebagai fitur wajib.

---

## 1. Ringkasan Produk

Aplikasi web internal untuk mengelola tiket helpdesk IT dan administratif, dengan empat peran pengguna: `super_admin`, `admin`, `staff`, dan `end_user`. Sistem mendukung submit dan tracking tiket, assignment staff, status workflow, komentar tiket, SLA otomatis, audit trail, dashboard analytics, dan export laporan CSV.

Positioning produk: project ini adalah project portofolio profesional yang ditargetkan sebagai **production-ready v1** untuk sistem internal skala kecil/menengah. Produk ini belum diposisikan sebagai enterprise-grade penuh karena belum mencakup capability enterprise seperti SSO/SAML, high availability multi-region, centralized SIEM integration, distributed tracing, compliance workflow formal, dan disaster recovery dengan RPO/RTO formal.

Stack produk:

- Backend: Go + Gin
- Frontend: Next.js + TypeScript
- Database: PostgreSQL
- Auth: JWT access token + refresh token
- Dashboard: agregasi SQL di backend, visualisasi Recharts di frontend

---

## 2. Persona Pengguna

### 2.1 Admin

Super admin adalah otoritas tertinggi sistem. Admin merepresentasikan IT Manager/Human Capital/System Admin internal. Super admin dan admin membutuhkan kontrol atas user, role, kategori, SLA rules, ticket assignment, dashboard analytics, dan laporan CSV.

Kebutuhan utama:

- Mengelola user dan hak akses
- Melihat semua tiket
- Assign/reassign tiket ke staff
- Melihat dashboard performa tim
- Export laporan periodik

### 2.2 Staff

Staff adalah agent IT yang menyelesaikan tiket yang di-assign kepadanya.

Kebutuhan utama:

- Melihat tiket yang ditugaskan kepadanya
- Update status tiket sesuai workflow
- Berkomunikasi dengan requester melalui komentar
- Melihat dashboard performa dirinya sendiri tanpa melihat performa staff lain

### 2.3 End-user

End-user adalah karyawan/requester yang mengajukan tiket.

Kebutuhan utama:

- Register akun sendiri
- Submit tiket dengan kategori dan prioritas
- Melihat status tiket miliknya sendiri
- Menambahkan komentar pada tiket miliknya sendiri

### 2.4 Analyst/Manajemen

Untuk produksi v1, kebutuhan analyst direpresentasikan oleh akun admin yang mengakses dashboard analytics dan export CSV.

Kebutuhan utama:

- Data historis yang bisa dipercaya
- Filter dashboard berdasarkan tanggal, kategori, dan staff
- Export CSV untuk laporan mingguan/bulanan

---

## 3. User Stories & Acceptance Criteria

### Epic A - Authentication & RBAC

**US-A1: Register end-user**

Sebagai end-user, saya bisa mendaftar akun sendiri supaya bisa submit tiket.

Acceptance criteria:

- Endpoint register publik selalu menghasilkan role `end_user`.
- Client tidak bisa menentukan role saat register publik.
- Role `super_admin`, `admin`, dan `staff` hanya bisa diberikan oleh super admin/admin melalui user management atau command bootstrap internal untuk akun pertama.

**US-A2: Login aman dengan token pair**

Sebagai user, saya bisa login dan mendapat sesi yang aman.

Acceptance criteria:

- Login mengembalikan `access_token`, `refresh_token`, dan `user`.
- Password disimpan dengan bcrypt.
- Password hash tidak pernah dikembalikan di response.
- Access token berumur pendek, target 15 menit.
- Refresh token disimpan sebagai hash di database.
- Refresh token punya expiry dan bisa dicabut saat logout.
- Refresh token expired/revoked ditolak.

**US-A3: Refresh access token**

Sebagai user yang masih punya refresh token valid, saya bisa mendapatkan access token baru tanpa login ulang.

Acceptance criteria:

- `POST /api/auth/refresh` menerima refresh token.
- Jika refresh token valid, endpoint mengembalikan access token baru.
- Jika refresh token expired, revoked, atau tidak dikenal, endpoint mengembalikan `401`.

**US-A4: Logout**

Sebagai user, saya bisa logout sehingga refresh token aktif tidak bisa dipakai lagi.

Acceptance criteria:

- `POST /api/auth/logout` mencabut refresh token aktif.
- Refresh token yang sudah logout tidak bisa dipakai di endpoint refresh.

**US-A5: RBAC backend**

Sebagai sistem, saya harus menolak akses endpoint yang tidak sesuai role.

Acceptance criteria:

- Endpoint admin-only menerima `super_admin` dan `admin`, lalu mengembalikan `403` saat diakses staff/end-user.
- Endpoint staff-only mengembalikan `403` saat diakses end-user.
- Enforcement dilakukan di middleware backend, bukan hanya UI.

### Epic B - Ticket Management

**US-B1: Submit ticket**

Sebagai end-user, saya bisa submit tiket baru.

Acceptance criteria:

- Field wajib: title, description, category.
- Priority default `Medium` jika tidak dipilih.
- `sla_deadline` dihitung otomatis saat tiket dibuat.
- Ticket creation mencatat status awal `Open`.

**US-B2: Assign/reassign ticket**

Sebagai admin, saya bisa assign/reassign tiket ke staff tertentu.

Acceptance criteria:

- Assignment mengubah `tickets.assignee_id`.
- Assignment/reassignment dicatat di `ticket_activity_log`.
- Log mencatat actor, action, old value, new value, dan timestamp.
- Staff yang di-assign bisa melihat tiket tersebut di list-nya.
- UI detail tiket admin menyediakan dropdown staff dan tombol assign/reassign.

**US-B3: Update status ticket**

Sebagai staff/admin, saya bisa update status tiket sesuai workflow.

Acceptance criteria:

- Workflow utama: `Open -> In Progress -> Pending -> Resolved -> Closed`.
- Reopen dari `Resolved`/`Closed` ke `Open` boleh sebagai kasus khusus dan wajib tercatat.
- Setiap perubahan status masuk ke `ticket_status_log`.
- Update status dan insert log dilakukan dalam satu database transaction.

**US-B4: Ticket comments**

Sebagai requester/staff/admin, saya bisa menambahkan komentar pada tiket yang boleh saya akses.

Acceptance criteria:

- Komentar terurut ascending berdasarkan waktu.
- Komentar menampilkan `author_name` dan `author_role`.
- End-user hanya bisa komentar pada tiket miliknya sendiri.
- Staff hanya bisa komentar pada tiket yang di-assign kepadanya.
- Admin bisa komentar pada semua tiket.

### Epic C - SLA Tracking

**US-C1: Auto-calculate SLA deadline**

Sebagai sistem, saya harus menghitung SLA deadline otomatis saat tiket dibuat.

Acceptance criteria:

- SLA memakai `sla_rules.resolution_hours` berdasarkan `category_id + priority`.
- Jika rule spesifik tidak ada, fallback ke `categories.default_sla_hours`.
- `sla_deadline = created_at + resolution_hours`.

**US-C2: SLA state**

Sebagai admin/staff, saya ingin melihat tiket yang masih aman, at-risk, atau breached.

Acceptance criteria:

- `breached` jika `now > sla_deadline` dan status belum `Resolved`/`Closed`.
- `at_risk` jika sisa waktu `<= 20%` dari total SLA.
- `ok` untuk kondisi lain.
- Fixed threshold seperti 2 jam tidak dipakai.

### Epic D - Dashboard & Reporting

**US-D1: Dashboard summary**

Sebagai admin, saya ingin melihat ringkasan jumlah tiket per status dan SLA compliance.

Acceptance criteria:

- `GET /api/dashboard/summary` mengembalikan status summary dan SLA compliance percentage.
- Endpoint mendukung query params `from`, `to`, `category_id`, dan `staff_id`.
- Agregasi dihitung di backend/SQL.

**US-D2: Ticket trend**

Sebagai admin, saya ingin melihat tren tiket masuk mingguan/bulanan.

Acceptance criteria:

- `GET /api/dashboard/trend` mengembalikan data trend.
- Endpoint mendukung query params `from`, `to`, `category_id`, dan `staff_id`.
- Frontend merender trend sebagai line chart.

**US-D3: Staff performance**

Sebagai admin, saya ingin melihat ranking performa staff.

Acceptance criteria:

- `GET /api/dashboard/staff-performance` mengembalikan jumlah tiket selesai dan rata-rata waktu penyelesaian per staff.
- Ranking memakai SQL window function `RANK() OVER()`.
- Endpoint mendukung query params `from`, `to`, `category_id`, dan `staff_id`.
- Endpoint menerima `super_admin` dan `admin`.

**US-D4: Staff scoped dashboard**

Sebagai staff, saya ingin melihat performa saya sendiri tanpa melihat performa staff lain.

Acceptance criteria:

- `GET /api/dashboard/my-performance` membaca staff id dari JWT.
- Endpoint tidak menerima `staff_id` dari client.
- Endpoint mendukung `from`, `to`, dan `category_id`.
- Staff tidak bisa mengakses endpoint dashboard admin.

**US-D5: SLA breaches table**

Sebagai admin, saya ingin melihat daftar tiket yang sudah melewati SLA.

Acceptance criteria:

- `GET /api/dashboard/sla-breaches` mengembalikan tiket unresolved yang sudah melewati deadline.
- Endpoint mendukung query params `from`, `to`, `category_id`, dan `staff_id`.

**US-D6: Export CSV**

Sebagai admin, saya ingin mengekspor laporan periodik ke CSV.

Acceptance criteria:

- `GET /api/reports/export` menghasilkan file CSV.
- Endpoint mendukung query params `from`, `to`, `category_id`, dan `staff_id`.
- Endpoint menerima `super_admin` dan `admin`.
- CSV bisa diunduh langsung dari browser.
- Kolom minimal: `ticket_id`, `title`, `category`, `priority`, `status`, `requester`, `assignee`, `created_at`, `resolved_at`, `sla_deadline`, `sla_state`.

---

## 4. Functional Requirements Summary

| FR ID | Requirement | Terkait BRD |
|---|---|---|
| FR-01 | Register, login, refresh token, logout | BR-01, BR-03 |
| FR-02 | RBAC middleware 4 role | BR-03 |
| FR-03 | CRUD/read ticket, status transition, comments | BR-01 |
| FR-04 | Assignment/reassignment dengan audit trail | BR-01, BR-03 |
| FR-05 | CRUD categories dan SLA rules | BR-01, BR-02 |
| FR-06 | Auto-calculate SLA deadline dan SLA state | BR-02 |
| FR-07 | Dashboard summary, trend, staff performance, my-performance, SLA breaches | BR-04 |
| FR-08 | Dashboard filter `from`, `to`, `category_id`, `staff_id` | BR-04 |
| FR-09 | Export CSV | BR-05 |
| FR-10 | Technical documentation dan user guide | BR-06 |

---

## 5. Non-Functional Requirements

| Aspek | Requirement |
|---|---|
| Security | Password bcrypt, access token pendek, refresh token hashed, backend RBAC enforced |
| Performance | Dashboard aggregation dihitung di SQL/backend |
| Reliability | Status update, ticket creation, dan assignment log dilakukan transactionally |
| Auditability | Status changes dan assignment changes punya audit trail |
| Maintainability | Backend layered: handler -> service -> repository |
| Usability | UI role-aware, form punya loading state dan feedback toast |
| Reporting | CSV export stabil, filterable, dan bisa diunduh dari browser |
| Operational Readiness | Health check, readiness check, structured logging, migration strategy, backup/restore procedure |
| Abuse Protection | Rate limiting untuk login, register, refresh token, dan export CSV |

---

## 6. Out of Scope

- Real-time chat
- Push notification/mobile app
- Integrasi HRIS/Active Directory
- Multi-tenant
- Procurement/payment
- PDF export
- Attachment upload untuk produksi v1
- SSO/SAML/OAuth enterprise
- High availability multi-region
- Distributed tracing penuh
- Automated incident alerting
- SIEM/security event integration
- Advanced audit compliance workflow
- Data retention/legal hold policy
- Disaster recovery dengan RPO/RTO formal
- Load testing dan capacity planning formal
- Dynamic permission matrix di luar role tetap `super_admin`, `admin`, `staff`, dan `end_user`

### 6.1 Batasan Enterprise-Grade

Alasan produk ini belum diklaim enterprise-grade penuh:

- Belum ada SSO/SAML/OAuth enterprise untuk integrasi identitas perusahaan.
- Belum ada integrasi HRIS/Active Directory untuk lifecycle user otomatis.
- Belum ada high availability multi-region dan failover otomatis.
- Belum ada disaster recovery plan dengan target RPO/RTO formal.
- Belum ada centralized SIEM/security event integration.
- Belum ada distributed tracing dan observability stack penuh.
- Belum ada compliance workflow formal untuk audit approval, data retention, dan legal hold.
- Belum ada load testing formal untuk membuktikan kapasitas produksi pada skala besar.
- Permission model masih memakai role tetap, belum dynamic permission matrix granular.

---

## 7. Definition of Done

- Semua user story utama bisa didemokan dengan akun super admin, admin, staff, dan end-user.
- Register publik selalu membuat `end_user`.
- Login, refresh, dan logout berfungsi.
- RBAC backend terbukti lewat test `401`, `403`, dan `200`.
- End-user hanya melihat tiket miliknya sendiri.
- Staff hanya melihat tiket assigned.
- Admin bisa assign/reassign tiket dan activity log tercatat.
- Komentar menampilkan nama dan role author.
- SLA deadline dan SLA state benar, termasuk at-risk 20%.
- Dashboard admin mendukung filter `from`, `to`, `category_id`, `staff_id`.
- Staff dashboard memakai `/api/dashboard/my-performance`.
- Export CSV bisa diunduh dan sesuai filter.
- Seed data realistis membuat dashboard tidak kosong/seragam.
- Dokumentasi teknis dan user guide tersedia di folder `docs/`.
- Health/readiness endpoint tersedia.
- Structured logging aktif di backend.
- Rate limit auth dan export endpoint aktif.
- Migration, backup, restore, dan deployment procedure terdokumentasi.

---

## 8. Referensi

- `helpdesk-brd.md`
- `helpdesk-ticketing-system-spec.md`
- `helpdesk-step-by-step-guide.md`
