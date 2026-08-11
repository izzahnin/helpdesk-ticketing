# Business Requirement Document (BRD)
## Internal Helpdesk & Ticketing System dengan Dashboard Analytics

Dokumen ini menjelaskan kebutuhan bisnis final untuk project Internal Helpdesk & Ticketing System. Dokumen ini sudah diselaraskan dengan `helpdesk-prd.md`, `helpdesk-ticketing-system-spec.md`, dan `helpdesk-step-by-step-guide.md`.

---

## 1. Latar Belakang

Perusahaan membutuhkan kanal resmi untuk mengelola permintaan dan keluhan internal karyawan, terutama terkait IT dan administrasi: masalah hardware, software, jaringan, akses sistem, dan request operasional lain. Tanpa sistem ticketing, proses biasanya berjalan lewat chat/email sehingga status pekerjaan tidak jelas, SLA tidak terukur, histori penyelesaian sulit ditelusuri, dan manajemen tidak memiliki data yang cukup untuk mengevaluasi layanan.

Project ini dibuat sebagai simulasi sistem internal perusahaan/BUMN. Fokusnya bukan sekadar CRUD tiket, tetapi sistem operasional yang menunjukkan:

- tata kelola akses berbasis role
- proses ticketing end-to-end
- SLA tracking otomatis
- dashboard analytics berbasis data historis
- export laporan CSV untuk kebutuhan monitoring rutin
- audit trail atas perubahan penting

Catatan positioning: project ini adalah project portofolio profesional dengan target **production-ready v1** untuk sistem internal skala kecil/menengah. Project ini belum diklaim sebagai sistem enterprise-grade penuh karena beberapa capability enterprise lanjutan masih di luar scope, seperti SSO/SAML, high availability multi-region, distributed tracing, SIEM integration, dan compliance audit formal.

---

## 2. Tujuan Bisnis

1. Menyediakan satu kanal resmi bagi karyawan untuk mengajukan permintaan dan keluhan internal.
2. Mengurangi proses informal lewat chat/email yang tidak punya status, ownership, dan histori jelas.
3. Memberikan visibilitas atas beban kerja staff/agent dan performa penyelesaian tiket.
4. Menegakkan SLA berdasarkan kategori dan prioritas tiket.
5. Menyediakan dashboard analytics untuk memantau status tiket, tren tiket, SLA compliance, SLA breach, dan performa staff.
6. Memberikan akses dashboard yang tepat guna: admin melihat performa organisasi, staff melihat performa dirinya sendiri.
7. Menyediakan export CSV untuk laporan periodik.
8. Menerapkan RBAC agar setiap role hanya bisa melihat dan melakukan aksi yang sesuai tanggung jawabnya.
9. Menyediakan audit trail untuk perubahan status dan assignment/reassignment tiket.

---

## 3. Scope Produksi V1

### 3.1 In Scope

- Register end-user
- Login dengan access token + refresh token
- Refresh access token
- Logout dengan revoke refresh token
- RBAC 3 role: `admin`, `staff`, `end_user`
- User management oleh admin
- Category management oleh admin
- SLA rule management oleh admin
- Submit tiket
- Role-scoped ticket list
- Ticket detail
- Assignment/reassignment tiket oleh admin
- Status transition oleh staff/admin
- Komentar tiket dengan author name dan author role
- Status log untuk perubahan status
- Activity log untuk assignment/reassignment
- SLA deadline otomatis
- SLA state: `ok`, `at_risk`, `breached`
- Dashboard admin dengan filter `from`, `to`, `category_id`, `staff_id`
- Dashboard staff scoped via `/api/dashboard/my-performance`
- SLA breach table
- Export CSV oleh admin
- Seed data realistis
- Technical documentation dan user guide
- Production-ready security baseline
- Structured application logging
- Rate limiting untuk endpoint auth
- Health check dan readiness check
- Database migration strategy
- Backup dan restore procedure
- Environment-based configuration

### 3.2 Out of Scope

- Real-time chat
- Push notification/mobile app
- Integrasi HRIS/Active Directory
- Multi-tenant
- Procurement/payment
- PDF export
- Attachment upload untuk produksi v1
- SSO/SAML/OAuth enterprise
- High availability multi-region
- Full observability stack dengan distributed tracing
- SIEM/security event integration
- Advanced audit compliance workflow
- Data retention/legal hold policy
- Disaster recovery dengan RPO/RTO formal
- Load testing dan capacity planning formal
- Role hierarchy/permission matrix dinamis di luar 3 role utama
- Email notification SLA breach

---

## 4. Stakeholders

| Stakeholder | Kepentingan |
|---|---|
| Admin / IT Manager | Mengelola user, role, kategori, SLA rules, assignment, dashboard organisasi, dan export laporan |
| Staff / Agent IT | Menangani tiket yang di-assign, update status, memberi komentar, dan melihat performa dirinya sendiri |
| End-user / Karyawan | Submit tiket, melihat progres tiket miliknya sendiri, dan berkomunikasi lewat komentar |
| Analyst / Manajemen | Melihat dashboard dan laporan CSV untuk evaluasi layanan |

Untuk produksi v1, kebutuhan analyst/manajemen direpresentasikan oleh akun admin. Tidak ada role analyst terpisah.

---

## 5. Business Requirements

| ID | Requirement | Justifikasi |
|---|---|---|
| BR-01 | Sistem harus mencatat setiap tiket dengan title, description, requester, category, priority, status, assignee, SLA deadline, dan timestamp | Menghilangkan ambiguitas proses manual |
| BR-02 | Sistem harus menghitung SLA deadline otomatis berdasarkan category + priority dengan fallback default SLA kategori | Menjadikan standar layanan terukur |
| BR-03 | Sistem harus menandai SLA state `ok`, `at_risk`, dan `breached` | Membantu prioritisasi tiket yang hampir/melewati deadline |
| BR-04 | Sistem harus menerapkan RBAC di backend | Menjamin least privilege dan mencegah akses data lintas role |
| BR-05 | Sistem harus mendukung refresh token dan logout | Memberikan session handling yang lebih production-grade |
| BR-06 | Sistem harus mencatat audit trail untuk status change dan assignment/reassignment | Mendukung accountability dan penelusuran perubahan |
| BR-07 | Sistem harus menyediakan dashboard admin dengan filter tanggal, kategori, dan staff | Mendukung monitoring organisasi dan analisis operasional |
| BR-08 | Sistem harus menyediakan dashboard staff scoped berdasarkan identitas JWT | Staff bisa memantau performa dirinya tanpa melihat data staff lain |
| BR-09 | Sistem harus menyediakan export CSV sesuai filter dashboard admin | Mendukung laporan periodik ke manajemen |
| BR-10 | Sistem harus menyediakan seed data realistis | Dashboard bisa didemokan dengan data yang masuk akal |
| BR-11 | Sistem harus memiliki technical documentation dan user guide | Mendukung knowledge transfer dan kesiapan demo/interview |

---

## 6. Role & Business Rules

### 6.1 Admin

Admin dapat:

- melihat semua tiket
- assign/reassign tiket ke staff
- update status tiket
- komentar di semua tiket
- manage user dan role
- manage kategori dan SLA rules
- melihat dashboard organisasi
- memakai semua filter dashboard
- export CSV

### 6.2 Staff

Staff dapat:

- melihat tiket yang di-assign ke dirinya
- update status tiket assigned
- komentar di tiket assigned
- melihat dashboard performa dirinya sendiri

Staff tidak dapat:

- melihat performa staff lain
- memakai `staff_id` bebas untuk mengubah scope dashboard
- manage user/kategori/SLA rules
- export CSV

### 6.3 End-user

End-user dapat:

- register sendiri
- submit tiket
- melihat tiket miliknya sendiri
- komentar di tiket miliknya sendiri

End-user tidak dapat:

- melihat tiket user lain
- update status
- assign tiket
- melihat dashboard
- export CSV

---

## 7. SLA Business Rules

- SLA ditentukan dari kombinasi category + priority di `sla_rules`.
- Jika rule spesifik tidak ada, sistem memakai `categories.default_sla_hours`.
- `sla_deadline = created_at + resolution_hours`.
- Tiket `breached` jika waktu sekarang melewati `sla_deadline` dan status belum `Resolved`/`Closed`.
- Tiket `at_risk` jika sisa waktu kurang dari atau sama dengan 20% dari total SLA.
- Tiket `ok` untuk kondisi lainnya.

---

## 8. Dashboard & Reporting Requirements

Dashboard admin harus menampilkan:

- summary jumlah tiket per status
- SLA compliance percentage
- tren tiket masuk
- ranking performa staff
- SLA breach table

Filter dashboard admin:

- `from`
- `to`
- `category_id`
- `staff_id`

Dashboard staff harus menampilkan:

- total tiket assigned
- total tiket resolved
- rata-rata waktu penyelesaian
- SLA compliance pribadi
- status summary pribadi
- SLA breaches pada tiket miliknya

Export CSV:

- admin-only
- memakai filter yang sama dengan dashboard admin
- bisa diunduh dari browser
- format minimal kolom: `ticket_id`, `title`, `category`, `priority`, `status`, `requester`, `assignee`, `created_at`, `resolved_at`, `sla_deadline`, `sla_state`

---

## 9. Constraints

- Project dibuat sebagai sistem internal siap produksi v1, dengan data dummy/seed data untuk development dan demo.
- Data demo bukan data karyawan sungguhan.
- Tools development boleh memakai open-source atau layanan development murah, tetapi desain sistem harus tetap valid saat dipindahkan ke environment produksi berbayar/managed.
- PostgreSQL digunakan lokal saat development dan bisa dipindah ke Neon/Supabase saat deploy.
- Scope produksi v1 dibatasi pada fitur yang tercantum di Section 3.1.

---

## 10. Success Criteria

Project dianggap berhasil jika:

1. End-to-end flow berjalan: register/login -> submit ticket -> assign -> update status -> comment -> dashboard -> export CSV.
2. RBAC terbukti lewat 3 akun berbeda: admin, staff, end-user.
3. Staff hanya melihat tiket dan dashboard miliknya sendiri.
4. Admin dapat melihat dashboard organisasi dan export CSV.
5. Refresh token, refresh endpoint, dan logout berfungsi.
6. SLA deadline dan SLA state dapat dijelaskan dan diuji.
7. Assignment/reassignment dan status change memiliki audit trail.
8. Dashboard menampilkan data realistis dari seed data.
9. Dokumentasi teknis dan user guide cukup untuk memahami sistem tanpa membaca kode.

---

## 11. Risks & Mitigation

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Scope creep | Project molor dan tidak selesai rapi | Ikuti in-scope produksi v1 di Section 3.1 |
| RBAC hanya diterapkan di frontend | Keamanan tidak valid | Enforcement wajib di backend middleware dan diuji |
| Seed data tidak realistis | Dashboard tidak meyakinkan | Buat variasi status, SLA, assignee, kategori, dan waktu |
| Dashboard filter dibuat di frontend saja | Query tidak scalable dan tidak valid sebagai analytics backend | Filter dan agregasi wajib di SQL/backend |
| Refresh token disimpan plaintext | Risiko keamanan | Simpan hash refresh token di database |
| Assignment tidak tercatat | Audit trail tidak lengkap | Gunakan `ticket_activity_log` transactionally |
| Tidak ada backup/restore | Risiko kehilangan data saat produksi | Dokumentasikan backup harian dan restore drill |
| Endpoint auth brute-forced | Risiko account abuse | Terapkan rate limiting pada login/register/refresh |

---

## 12. Referensi Dokumen

- `helpdesk-prd.md`
- `helpdesk-ticketing-system-spec.md`
- `helpdesk-step-by-step-guide.md`
