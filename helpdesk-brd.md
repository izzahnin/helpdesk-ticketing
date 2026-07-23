# Business Requirement Document (BRD)
## Internal Helpdesk & Ticketing System dengan Dashboard Analytics

---

## 1. Latar Belakang

Perusahaan (dalam konteks ini disimulasikan sebagai instansi/BUMN dengan divisi IT internal) membutuhkan sistem untuk mengelola permintaan/keluhan internal karyawan — mulai dari masalah hardware, software, akses sistem, sampai request administratif — secara terstruktur. Saat ini proses semacam ini biasanya berjalan lewat chat/email tanpa pencatatan status yang jelas, tanpa SLA yang terukur, dan tanpa data historis yang bisa dianalisis untuk perbaikan layanan.

Project ini dibangun sebagai **simulasi sistem internal perusahaan**, dirancang untuk merepresentasikan kebutuhan nyata di 3 area fungsi sekaligus: **Human Capital (people & access management)**, **System Administration (IT service management)**, dan **Data Analyst (reporting & insight)** — karena ketiganya sama-sama berkepentingan terhadap proses ticketing ini dari sudut pandang berbeda.

---

## 2. Tujuan Bisnis (Business Objectives)

1. Menyediakan satu kanal resmi bagi karyawan untuk mengajukan permintaan/keluhan IT & administratif, menggantikan komunikasi informal via chat/email.
2. Memberikan visibilitas atas beban kerja staff/agent penyelesai tiket, termasuk siapa mengerjakan apa dan berapa lama.
3. Menegakkan **SLA (Service Level Agreement)** — target waktu penyelesaian berdasarkan kategori dan prioritas — supaya ada standar layanan yang terukur, bukan sekadar "dikerjakan kalau sempat".
4. Menyediakan data historis tiket yang bisa dianalisis untuk mengidentifikasi pola (kategori apa paling sering, bottleneck di mana, staff mana yang overload) sebagai dasar pengambilan keputusan operasional.
5. Mengatur **hak akses berlapis (RBAC)** sehingga informasi dan aksi yang tersedia sesuai dengan tanggung jawab masing-masing peran (admin, staff, end-user).

---

## 3. Ruang Lingkup (Scope)

### 3.1 Termasuk dalam Scope
- Submit, tracking, dan penyelesaian tiket internal
- Manajemen kategori dan prioritas tiket beserta aturan SLA-nya
- RBAC 3 level (Admin, Staff/Agent, End-user)
- Dashboard analitik untuk monitoring performa dan SLA compliance
- Ekspor laporan periodik (mingguan/bulanan)
- Dokumentasi teknis dan user guide

### 3.2 Di Luar Scope (versi awal)
- Integrasi dengan sistem HR/Active Directory sungguhan (di project ini user dibuat manual oleh admin, bukan sync otomatis)
- Live chat real-time antara requester dan staff (komunikasi dilakukan lewat komentar tiket, bukan chat instan)
- Notifikasi push/mobile app
- Multi-tenant (project ini disimulasikan untuk satu organisasi/perusahaan saja)
- Integrasi pembayaran/procurement (di luar konteks helpdesk internal)

---

## 4. Pemangku Kepentingan (Stakeholders) — Disimulasikan

| Peran | Kepentingan terhadap sistem |
|---|---|
| **Admin/IT Manager** (analog Human Capital & Sysadmin) | Mengelola user & akses, memastikan SLA terpenuhi, melihat performa tim |
| **Staff/Agent IT** (analog System Administration) | Menyelesaikan tiket sesuai antrian & prioritas, mendokumentasikan penyelesaian |
| **Karyawan/End-user** | Mengajukan keluhan/permintaan, memantau status penyelesaian |
| **Analyst/Manajemen** (analog Data Analyst TREG V) | Membutuhkan laporan & insight dari data tiket untuk evaluasi layanan |

---

## 5. Kebutuhan Bisnis (Business Requirements)

| ID | Kebutuhan | Justifikasi |
|---|---|---|
| BR-01 | Sistem harus mencatat setiap tiket dengan kategori, prioritas, dan status yang jelas | Menghilangkan ambiguitas proses manual via chat/email |
| BR-02 | Sistem harus punya aturan SLA per kombinasi kategori+prioritas, dan menandai otomatis tiket yang melewati SLA | Standar layanan terukur, bukan subjektif |
| BR-03 | Sistem harus membatasi akses data & aksi berdasarkan role pengguna | Kepatuhan terhadap prinsip least privilege dan pemisahan tanggung jawab |
| BR-04 | Sistem harus menyediakan dashboard yang menampilkan tren, beban kerja staff, dan SLA compliance | Mendukung pengambilan keputusan berbasis data, bukan asumsi |
| BR-05 | Sistem harus bisa mengekspor laporan periodik | Kebutuhan pelaporan rutin ke manajemen |
| BR-06 | Sistem harus punya dokumentasi teknis dan user guide | Kesiapan operasional & knowledge transfer |

---

## 6. Batasan (Constraints)

- Dibangun sebagai project portofolio individu — data adalah dummy/seed data, bukan data karyawan sungguhan
- Tidak ada anggaran untuk software berbayar — semua tools yang dipakai gratis/open-source. Database menggunakan PostgreSQL yang di-install lokal untuk kebutuhan development, dan baru dipindahkan ke layanan hosted gratis (Neon atau Supabase) pada tahap deploy
- Waktu pengerjaan terbatas (target ±1 minggu kerja efektif), sehingga fitur dibatasi ke inti (lihat Section 3.2)

---

## 7. Kriteria Keberhasilan (Success Criteria)

Project dianggap berhasil dari sisi bisnis jika:
1. Alur end-to-end (submit → assign → resolve → analisis) berjalan tanpa manual intervention di database
2. RBAC benar-benar membatasi apa yang bisa dilihat/dilakukan tiap role (bisa didemokan dengan 3 akun berbeda)
3. Dashboard menghasilkan angka yang masuk akal dan bisa dijelaskan alur perhitungannya (SLA compliance %, rata-rata waktu penyelesaian, dll)
4. Ada dokumentasi yang cukup bagi orang lain (atau interviewer) untuk memahami sistem tanpa membaca kode

---

## 8. Risiko

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Scope creep (menambah fitur di luar rencana karena "kelihatan gampang") | Waktu pengerjaan molor, project tidak selesai rapi | Ikuti Definition of Done yang sudah ditetapkan di PRD/guide |
| Dashboard dengan data seed yang tidak realistis (misal semua tiket selesai instan) | Insight dashboard jadi tidak meyakinkan saat didemokan | Seed data dibuat dengan variasi waktu penyelesaian, status, dan distribusi kategori yang realistis |
| RBAC hanya diterapkan di frontend (security theater) | Tidak valid secara teknis, mudah ketahuan saat ditanya interviewer | RBAC wajib di-enforce di middleware backend, frontend hanya UX layer |

---

## 9. Referensi Dokumen Terkait
- `helpdesk-prd.md` — kebutuhan produk & fungsional detail
- `helpdesk-ticketing-system-spec.md` — spesifikasi teknis (ERD, API, struktur folder)
- `helpdesk-step-by-step-guide.md` — panduan implementasi end-to-end
