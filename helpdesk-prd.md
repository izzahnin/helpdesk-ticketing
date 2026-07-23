# Product Requirement Document (PRD)
## Internal Helpdesk & Ticketing System dengan Dashboard Analytics

Dokumen ini menerjemahkan kebutuhan bisnis di `helpdesk-brd.md` menjadi kebutuhan produk yang konkret: siapa penggunanya, apa yang bisa mereka lakukan, dan bagaimana sistem seharusnya berperilaku di tiap skenario.

---

## 1. Ringkasan Produk

Aplikasi web internal untuk mengelola tiket helpdesk (IT & administratif), dengan tiga peran pengguna (Admin, Staff/Agent, End-user), penegakan SLA otomatis, dan dashboard analitik. Dibangun dengan Go + Fiber (backend) dan Next.js + TypeScript (frontend), PostgreSQL sebagai database utama — dijalankan lokal selama development, lalu dipindahkan ke layanan hosted (Neon/Supabase) saat tahap deploy.

---

## 2. Persona Pengguna

### 2.1 Admin (IT Manager / Human Capital)
Bertanggung jawab atas konfigurasi sistem: kelola user, kategori, aturan SLA, dan assignment tiket. Butuh visibilitas penuh terhadap semua tiket dan performa tim.

**Kebutuhan utama:** kontrol penuh, laporan ringkas, kemampuan intervensi cepat (reassign tiket yang macet).

### 2.2 Staff/Agent (System Administrator)
Mengerjakan tiket yang di-assign kepadanya. Butuh antrian kerja yang jelas, prioritas yang mudah dibaca, dan cara cepat update status + komunikasi dengan requester.

**Kebutuhan utama:** kejelasan prioritas, minim friksi saat update status, riwayat komunikasi tidak hilang.

### 2.3 End-user / Karyawan (Requester)
Mengajukan keluhan/permintaan dan ingin tahu progres tanpa harus bertanya manual. Tidak butuh (dan tidak boleh) melihat tiket milik orang lain.

**Kebutuhan utama:** submit gampang, status transparan, notifikasi kalau ada update.

### 2.4 Analyst/Manajemen (Data Analyst)
Tidak submit atau proses tiket, tapi konsumsi laporan/dashboard. Dalam versi awal, peran ini direpresentasikan lewat akun Admin yang mengakses halaman analytics — bukan role terpisah — supaya scope tetap ramping.

**Kebutuhan utama:** angka yang bisa dipercaya, breakdown yang bisa difilter, ekspor data.

---

## 3. User Stories & Acceptance Criteria

### Epic A — Autentikasi & RBAC

**US-A1:** Sebagai end-user, saya bisa mendaftar akun sendiri supaya bisa submit tiket.
- AC: Register hanya menghasilkan role `end_user`. Role lain (`staff`, `admin`) hanya bisa dibuat oleh admin.

**US-A2:** Sebagai user (semua role), saya bisa login dan mendapat sesi yang aman.
- AC: Login mengembalikan JWT access token + refresh token. Password disimpan ter-hash (bcrypt), tidak pernah dikembalikan di response manapun.

**US-A3:** Sebagai sistem, saya harus menolak akses ke endpoint yang tidak sesuai role.
- AC: Endpoint admin-only (misal `DELETE /api/users/:id`) mengembalikan `403 Forbidden` jika diakses oleh staff/end-user, diverifikasi di middleware — bukan hanya disembunyikan di UI.

### Epic B — Manajemen Tiket

**US-B1:** Sebagai end-user, saya bisa submit tiket baru dengan kategori & prioritas.
- AC: Field wajib: judul, deskripsi, kategori. Prioritas defaultnya `Medium` kalau tidak dipilih. SLA deadline dihitung otomatis saat tiket dibuat berdasarkan kategori+prioritas.

**US-B2:** Sebagai admin, saya bisa assign/reassign tiket ke staff tertentu.
- AC: Assignment mengubah `assignee_id` dan mencatat perubahan di `ticket_status_log` (atau log terpisah untuk assignment). Staff yang di-assign otomatis bisa melihat tiket tsb di list-nya.

**US-B3:** Sebagai staff, saya bisa update status tiket dan sistem mencatat riwayatnya.
- AC: Transisi status mengikuti urutan `Open → In Progress → Pending → Resolved → Closed` (tidak bisa lompat mundur tanpa alasan eksplisit, misal reopen dari Closed ke Open dicatat sebagai kasus khusus). Setiap perubahan status masuk ke `ticket_status_log` dengan timestamp dan siapa yang mengubah.

**US-B4:** Sebagai requester atau staff, saya bisa menambahkan komentar di tiket sebagai riwayat komunikasi.
- AC: Komentar terurut berdasarkan waktu, menampilkan nama & role penulis. End-user hanya bisa komentar di tiketnya sendiri.

### Epic C — SLA Tracking

**US-C1:** Sebagai sistem, saya harus menghitung SLA deadline otomatis saat tiket dibuat.
- AC: `sla_deadline = created_at + resolution_hours` (dari tabel `sla_rules` berdasarkan kategori+prioritas). Jika kombinasi kategori+prioritas tidak punya rule spesifik, pakai default SLA kategori.

**US-C2:** Sebagai admin/staff, saya ingin melihat tiket mana yang sudah/hampir melewati SLA.
- AC: Tiket dengan `now() > sla_deadline` dan status belum `Resolved`/`Closed` ditandai "SLA Breached". Tiket yang sisa waktunya < 20% dari total SLA ditandai "At Risk" (warna kuning), breach ditandai merah.

### Epic D — Dashboard & Reporting

**US-D1:** Sebagai admin/analyst, saya ingin melihat ringkasan jumlah tiket per status.
- AC: Endpoint `/api/dashboard/summary` mengembalikan hitungan per status + SLA compliance rate (%), dapat difilter rentang tanggal.

**US-D2:** Sebagai admin/analyst, saya ingin melihat tren tiket masuk dari waktu ke waktu.
- AC: Endpoint `/api/dashboard/trend` mengembalikan agregasi mingguan/bulanan, dirender sebagai line chart di frontend.

**US-D3:** Sebagai admin/analyst, saya ingin melihat ranking performa staff.
- AC: Endpoint `/api/dashboard/staff-performance` mengembalikan jumlah tiket selesai & rata-rata waktu penyelesaian per staff, diurutkan pakai `RANK() OVER()`.

**US-D4:** Sebagai admin, saya ingin mengekspor laporan periodik.
- AC: Endpoint export menghasilkan file CSV berisi ringkasan tiket dalam rentang tanggal yang dipilih, bisa diunduh langsung dari browser.

---

## 4. Functional Requirements Summary (Traceability ke BRD)

| FR ID | Requirement | Terkait BRD |
|---|---|---|
| FR-01 | Register & login dengan JWT | BR-01 |
| FR-02 | RBAC middleware 3 level | BR-03 |
| FR-03 | CRUD tiket + status transition + log | BR-01 |
| FR-04 | CRUD kategori & SLA rules oleh admin | BR-01, BR-02 |
| FR-05 | Auto-kalkulasi & auto-flag SLA breach | BR-02 |
| FR-06 | Dashboard summary, trend, staff performance | BR-04 |
| FR-07 | Export laporan CSV | BR-05 |
| FR-08 | Dokumentasi teknis & user guide | BR-06 |

---

## 5. Non-Functional Requirements

| Aspek | Requirement |
|---|---|
| **Security** | Password di-hash bcrypt; JWT expiry pendek untuk access token (misal 15 menit) + refresh token; RBAC enforced di backend |
| **Performance** | Endpoint dashboard tidak melakukan agregasi di frontend — semua dihitung via SQL di backend agar skalabel |
| **Usability** | UI membedakan tampilan berdasarkan role secara jelas (tidak menampilkan menu yang tidak relevan) |
| **Reliability** | Perubahan status tiket tidak boleh hilang — dicatat di log meskipun request berikutnya gagal (idealnya dalam 1 transaction) |
| **Maintainability** | Kode terstruktur layered (handler → service → repository), konsisten dengan pola di Fleet Management System |
| **Auditability** | Semua perubahan status & assignment tiket harus punya jejak (siapa, kapan, dari status apa ke apa) |

---

## 6. Out of Scope (ditegaskan ulang dari BRD)
- Real-time chat, push notification, integrasi HRIS/AD, multi-tenant — lihat `helpdesk-brd.md` Section 3.2

## 7. Definition of Done (Produk)
- Semua User Story di Section 3 punya endpoint yang berfungsi dan bisa didemokan dengan 3 akun berbeda (admin, staff, end-user)
- Dashboard menampilkan data dari seed data yang realistis (bukan angka kosong/seragam)
- RBAC terbukti ditegakkan di backend (bisa ditunjukkan lewat test: request staff ke endpoint admin-only mengembalikan 403)
- Dokumentasi teknis & user guide tersedia di folder `docs/`

## 8. Referensi
- `helpdesk-brd.md`
- `helpdesk-ticketing-system-spec.md`
- `helpdesk-step-by-step-guide.md`
