# User Guide Helpdesk Ticketing System

Panduan ini menjelaskan cara memakai sistem Helpdesk Ticketing berdasarkan role, workflow, dan endpoint yang tersedia di Swagger.

Base URL local:

```text
http://localhost:4000
```

Swagger UI:

```text
http://localhost:4000/swagger/index.html
```

## 1. Role dan Hak Akses

Sistem memakai 4 role:

| Role | Hak akses utama |
| --- | --- |
| `super_admin` | Semua otoritas sistem, termasuk semua kemampuan admin |
| `admin` | Manage users, categories, SLA rules, assignment ticket, dashboard organisasi, export report |
| `staff` | Melihat ticket yang di-assign kepadanya, update status, komentar, dashboard performa dirinya |
| `end_user` | Register, submit ticket, melihat ticket miliknya, komentar pada ticket yang boleh diakses |

Backend adalah sumber kebenaran akses. Tampilan frontend boleh menyembunyikan menu tertentu, tetapi backend tetap mengecek token dan role.

## 1.1 Siapa yang Membuat Akun dan Memberi Role

Ada dua cara akun masuk ke sistem:

1. User melakukan register sendiri melalui endpoint publik.
2. Super admin/admin mengubah role user melalui user management.

Register publik selalu membuat akun dengan role:

```text
end_user
```

Artinya karyawan/requester boleh membuat akun sendiri, tetapi mereka tidak bisa memilih menjadi `staff`, `admin`, atau `super_admin`.

Policy pemberian role:

| Aksi | Yang boleh melakukan |
| --- | --- |
| Membuat akun `end_user` lewat register publik | siapa pun |
| Mengubah user menjadi `staff` | `super_admin` atau `admin` |
| Mengubah user menjadi `end_user` | `super_admin` atau `admin` |
| Mengubah user menjadi `admin` | hanya `super_admin` |
| Mengubah user menjadi `super_admin` | hanya `super_admin` |
| Membuat super admin pertama | command bootstrap internal |

Alur praktisnya:

```text
karyawan register -> role end_user
admin/super_admin review -> jika perlu, ubah menjadi staff
super_admin review -> jika perlu, ubah menjadi admin atau super_admin
```

Kenapa begitu:

- register publik tidak boleh memberi privilege tinggi
- admin biasa tidak boleh menciptakan super admin baru
- super admin menjadi otoritas tertinggi untuk role sensitif

## 2. Menjalankan Backend

Dari folder backend:

```powershell
cd backend
air
```

Atau tanpa live reload:

```powershell
cd backend
go run ./cmd/server
```

Jika muncul error port `4000` sudah dipakai, cek prosesnya:

```powershell
netstat -ano | findstr :4000
Get-Process -Id <PID> | Select-Object Id,ProcessName,Path
Stop-Process -Id <PID>
```

## 3. Membuat Super Admin Pertama

Register publik tidak bisa membuat `super_admin`. Akun super admin pertama dibuat melalui command internal.

Pastikan migration super admin sudah dijalankan jika database dibuat sebelum role ini ditambahkan:

```powershell
cd backend
psql "postgresql://helpdesk_user:helpdesk_pass@localhost:5432/helpdesk_db?sslmode=disable" -f migration/002_add_super_admin_role.sql
```

Pastikan `.env` berisi:

```env
SUPER_ADMIN_NAME=Super Admin
SUPER_ADMIN_EMAIL=superadmin@example.com
SUPER_ADMIN_PASSWORD=change-me-super-admin
```

Jalankan bootstrap:

```powershell
cd backend
go run ./cmd/create-super-admin
```

Jika sukses, login memakai email dan password tersebut melalui endpoint login.

## 4. Login dan Authorize di Swagger

Di Swagger UI:

1. Buka `POST /api/auth/login`.
2. Isi body:

```json
{
  "email": "superadmin@example.com",
  "password": "change-me-super-admin"
}
```

3. Copy nilai `access_token` dari response.
4. Klik tombol **Authorize**.
5. Pada `BearerAuth`, isi:

```text
Bearer <access_token>
```

`BearerAuth` dipakai untuk endpoint protected seperti tickets, users, categories, dashboard, dan reports.

`RefreshCookie` dipakai oleh endpoint refresh/logout. Setelah login, browser biasanya menyimpan cookie `refresh_token` otomatis. Jadi saat mencoba `POST /api/auth/refresh`, kamu tidak perlu mengetik refresh token manual selama Swagger UI dibuka dari origin backend yang sama.

## 5. Auth Flow

### Register End User

Endpoint:

```http
POST /api/auth/register
```

Body:

```json
{
  "name": "Budi User",
  "email": "budi@example.com",
  "password": "password123"
}
```

Hasil:

- akun dibuat dengan role `end_user`
- client tidak boleh menentukan role sendiri
- jika email sudah ada, response `409`

### Login

Endpoint:

```http
POST /api/auth/login
```

Body:

```json
{
  "email": "budi@example.com",
  "password": "password123"
}
```

Hasil:

- response berisi `access_token`
- response berisi data user
- refresh token dikirim lewat cookie `refresh_token`

### Melihat User yang Sedang Login

Endpoint:

```http
GET /api/auth/me
```

Butuh:

```text
BearerAuth
```

Hasil:

- data user dari token yang sedang dipakai

### Refresh Access Token

Endpoint:

```http
POST /api/auth/refresh
```

Butuh:

```text
RefreshCookie
```

Hasil:

- access token baru

Pakai ini ketika access token lama expired.

### Logout

Endpoint:

```http
POST /api/auth/logout
```

Butuh:

```text
RefreshCookie
```

Hasil:

- refresh token direvoke di database
- cookie refresh token dibersihkan

## 6. Skenario End User

End user adalah karyawan/requester yang membuat ticket.

### Melihat Category

Endpoint:

```http
GET /api/categories
```

Butuh:

```text
BearerAuth
```

Gunakan category `id` dari response untuk submit ticket.

### Submit Ticket

Endpoint:

```http
POST /api/tickets
```

Body:

```json
{
  "title": "Laptop tidak bisa connect WiFi",
  "description": "Sejak pagi laptop tidak bisa terhubung ke jaringan kantor.",
  "category_id": 1,
  "priority": "High"
}
```

Catatan:

- `title`, `description`, dan `category_id` wajib.
- `priority` boleh dikirim. Jika kosong, backend memakai default `Medium`.
- `requester_id`, `status`, dan `sla_deadline` ditentukan backend.

Hasil:

- ticket dibuat dengan status `Open`
- SLA deadline dihitung dari category dan priority
- activity log `ticket_created` dibuat

### Melihat Ticket Milik Sendiri

Endpoint:

```http
GET /api/tickets
```

Untuk `end_user`, hasilnya hanya ticket yang dia buat sendiri.

### Melihat Detail Ticket dan Komentar

Endpoint:

```http
GET /api/tickets/{id}
```

End user hanya bisa melihat ticket miliknya sendiri. Jika mencoba ticket user lain, backend mengembalikan `403`.

### Menambahkan Komentar

Endpoint:

```http
POST /api/tickets/{id}/comments
```

Body:

```json
{
  "message": "Saya sudah coba restart laptop, tapi masih belum bisa."
}
```

End user hanya bisa komentar pada ticket yang boleh dia akses.

## 7. Skenario Staff

Staff adalah agent/helpdesk yang mengerjakan ticket.

### Melihat Ticket yang Ditugaskan

Endpoint:

```http
GET /api/tickets
```

Untuk `staff`, hasilnya hanya ticket dengan `assignee_id` staff tersebut.

### Melihat Detail Ticket

Endpoint:

```http
GET /api/tickets/{id}
```

Staff hanya bisa melihat ticket yang di-assign kepadanya.

### Update Status Ticket

Endpoint:

```http
PATCH /api/tickets/{id}/status
```

Body:

```json
{
  "status": "In Progress"
}
```

Status yang tersedia:

```text
Open
In Progress
Pending
Resolved
Closed
```

Transition yang valid:

| Dari | Ke |
| --- | --- |
| `Open` | `In Progress` |
| `In Progress` | `Pending`, `Resolved` |
| `Pending` | `In Progress`, `Resolved` |
| `Resolved` | `Closed`, `Open` |
| `Closed` | `Open` |

Jika transition tidak valid, backend mengembalikan `400`.

### Menambahkan Komentar

Endpoint:

```http
POST /api/tickets/{id}/comments
```

Body:

```json
{
  "message": "Ticket sedang dicek oleh tim IT."
}
```

Komentar akan menyimpan `author_name` dan `author_role`.

### Melihat Dashboard Performa Sendiri

Endpoint:

```http
GET /api/dashboard/my-performance
```

Optional query:

```text
from=YYYY-MM-DD
to=YYYY-MM-DD
category_id=1
```

Contoh:

```http
GET /api/dashboard/my-performance?from=2026-08-01&to=2026-08-31&category_id=1
```

Staff tidak boleh mengirim `staff_id`. Scope staff diambil dari JWT.

Response berisi:

- total assigned ticket
- total resolved ticket
- rata-rata resolution hours
- SLA compliance
- summary status
- daftar SLA breach

## 8. Skenario Admin dan Super Admin

`super_admin` memiliki semua akses `admin`. Dalam guide ini, setiap bagian admin juga berlaku untuk super admin.

### Melihat Semua User

Endpoint:

```http
GET /api/users
```

Butuh role:

```text
super_admin atau admin
```

### Melihat List Staff Aktif

Endpoint:

```http
GET /api/users/staff
```

Dipakai saat assignment ticket ke staff.

### Update Role dan Status User

Endpoint:

```http
PATCH /api/users/{id}
```

Body:

```json
{
  "role": "staff",
  "is_active": true
}
```

Role yang valid:

```text
super_admin
admin
staff
end_user
```

Catatan:

- hanya `super_admin` yang boleh memberi role `admin` atau `super_admin`
- `admin` hanya boleh mengubah user menjadi `staff` atau `end_user`
- user inactive sebaiknya tidak bisa dipakai untuk workflow aktif
- register publik tetap tidak bisa membuat role selain `end_user`

### Membuat Category

Endpoint:

```http
POST /api/categories
```

Body:

```json
{
  "name": "Network",
  "default_sla_hours": 24
}
```

`default_sla_hours` dipakai jika tidak ada SLA rule khusus untuk priority ticket.

### Update Category

Endpoint:

```http
PATCH /api/categories/{id}
```

Body:

```json
{
  "name": "Network & Internet",
  "default_sla_hours": 12
}
```

### Delete Category

Endpoint:

```http
DELETE /api/categories/{id}
```

Gunakan hati-hati. Jika category masih dipakai ticket, database bisa menolak tergantung relasi yang ada.

### Membuat atau Update SLA Rule

Endpoint:

```http
POST /api/sla-rules
```

Body:

```json
{
  "category_id": 1,
  "priority": "High",
  "resolution_hours": 8
}
```

Priority yang valid:

```text
Low
Medium
High
Critical
```

Jika rule untuk kombinasi `category_id + priority` sudah ada, backend melakukan update.

### Melihat Semua Ticket

Endpoint:

```http
GET /api/tickets
```

Untuk `super_admin` dan `admin`, hasilnya semua ticket.

### Assign atau Reassign Ticket ke Staff

Endpoint:

```http
PATCH /api/tickets/{id}/assign
```

Body:

```json
{
  "assignee_id": 2
}
```

Hasil:

- `assignee_id` ticket berubah
- activity log mencatat `assigned` atau `reassigned`
- proses berjalan dalam transaction

### Update Status Ticket

Endpoint:

```http
PATCH /api/tickets/{id}/status
```

Admin dan super admin juga boleh update status memakai transition yang sama seperti staff.

### Melihat Dashboard Summary

Endpoint:

```http
GET /api/dashboard/summary
```

Optional query:

```text
from=YYYY-MM-DD
to=YYYY-MM-DD
category_id=1
staff_id=2
```

Response berisi:

- total ticket
- total open ticket
- total resolved ticket
- SLA compliance percentage
- status summary

### Melihat Trend Ticket

Endpoint:

```http
GET /api/dashboard/trend
```

Response berisi jumlah ticket per minggu.

### Melihat Staff Performance

Endpoint:

```http
GET /api/dashboard/staff-performance
```

Response berisi:

- rank
- staff id
- staff name
- jumlah resolved ticket
- rata-rata resolution hours

### Melihat SLA Breaches

Endpoint:

```http
GET /api/dashboard/sla-breaches
```

Response berisi ticket yang melewati SLA deadline dan belum selesai.

### Export CSV Report

Endpoint:

```http
GET /api/reports/export
```

Optional query:

```text
from=YYYY-MM-DD
to=YYYY-MM-DD
category_id=1
staff_id=2
```

CSV columns:

```csv
ticket_id,title,category,priority,status,requester,assignee,created_at,resolved_at,sla_deadline,sla_state
```

Endpoint ini rate limited. Jika terlalu sering dipanggil, backend bisa mengembalikan `429`.

## 9. System Check

### Health

Endpoint:

```http
GET /api/health
```

Dipakai untuk mengecek proses backend hidup.

### Readiness

Endpoint:

```http
GET /api/ready
```

Dipakai untuk mengecek backend siap melayani request dan database bisa diakses.

Jika database mati atau connection salah, endpoint ini bisa mengembalikan `503`.

## 10. Skenario Testing Manual di Swagger

Urutan testing yang disarankan:

1. Jalankan backend.
2. Buka Swagger UI.
3. Jalankan seeder jika ingin data demo siap pakai: `go run ./cmd/seed`.
4. Buat super admin dengan command bootstrap jika belum ada dan tidak memakai akun demo seeder.
5. Login sebagai super admin.
6. Copy `access_token` ke `BearerAuth`.
7. Cek `GET /api/auth/me`.
8. Buat category.
9. Buat SLA rule untuk category tersebut.
10. Register end user.
11. Update user lain menjadi `staff` jika belum ada staff.
12. Login sebagai end user.
13. Submit ticket.
14. Login ulang sebagai super admin/admin.
15. Assign ticket ke staff.
16. Login sebagai staff.
17. Cek list ticket staff.
18. Update status ticket dari `Open` ke `In Progress`.
19. Tambahkan komentar sebagai staff.
20. Login sebagai end user dan tambahkan komentar balasan.
21. Login sebagai super admin/admin.
22. Cek dashboard summary, trend, staff performance, dan SLA breaches.
23. Export CSV report.
24. Test refresh token dengan `POST /api/auth/refresh`.
25. Test logout dengan `POST /api/auth/logout`.

## 11. Expected Access Matrix

| Endpoint | end_user | staff | admin | super_admin |
| --- | --- | --- | --- | --- |
| `POST /api/auth/register` | yes | yes | yes | yes |
| `POST /api/auth/login` | yes | yes | yes | yes |
| `POST /api/auth/refresh` | yes | yes | yes | yes |
| `POST /api/auth/logout` | yes | yes | yes | yes |
| `GET /api/auth/me` | yes | yes | yes | yes |
| `GET /api/categories` | yes | yes | yes | yes |
| `POST/PATCH/DELETE /api/categories` | no | no | yes | yes |
| `POST /api/sla-rules` | no | no | yes | yes |
| `GET /api/tickets` | own | assigned | all | all |
| `GET /api/tickets/{id}` | own only | assigned only | all | all |
| `POST /api/tickets` | yes | yes | yes | yes |
| `POST /api/tickets/{id}/comments` | allowed ticket only | assigned only | all | all |
| `PATCH /api/tickets/{id}/status` | no | assigned role route | yes | yes |
| `PATCH /api/tickets/{id}/assign` | no | no | yes | yes |
| `GET /api/dashboard/my-performance` | no | yes | yes | yes |
| `GET /api/dashboard/summary` | no | no | yes | yes |
| `GET /api/dashboard/trend` | no | no | yes | yes |
| `GET /api/dashboard/staff-performance` | no | no | yes | yes |
| `GET /api/dashboard/sla-breaches` | no | no | yes | yes |
| `GET /api/reports/export` | no | no | yes | yes |
| `GET /api/health` | yes | yes | yes | yes |
| `GET /api/ready` | yes | yes | yes | yes |

Catatan: `POST /api/tickets` saat ini authenticated. Secara produk biasanya dipakai end user untuk submit ticket, tetapi backend tidak membatasi hanya end user.

## 12. Troubleshooting

### `401 Unauthorized`

Kemungkinan:

- belum klik **Authorize**
- format token salah
- access token expired

Format yang benar:

```text
Bearer <access_token>
```

Jika token expired, panggil:

```http
POST /api/auth/refresh
```

Lalu copy access token baru ke `BearerAuth`.

### `403 Forbidden`

Kemungkinan:

- role tidak punya akses endpoint
- user mencoba melihat ticket yang bukan miliknya
- staff mencoba akses dashboard admin

### `400 Bad Request`

Kemungkinan:

- body JSON tidak sesuai DTO
- path id bukan angka valid
- status transition tidak valid
- query date salah format

### `409 Conflict`

Biasanya email register sudah dipakai.

### `500 Internal Server Error`

Kemungkinan:

- database error
- constraint database menolak data
- query repository gagal

### `503 Service Unavailable`

Biasanya muncul dari `/api/ready` ketika database tidak bisa diakses.

### Cookie Refresh Tidak Muncul

Cek di browser:

```text
DevTools -> Application -> Cookies -> http://localhost:4000 -> refresh_token
```

Jika tidak ada:

- pastikan login dilakukan dari Swagger di origin backend yang sama
- pastikan backend berjalan di `localhost:4000`
- cek konfigurasi cookie/SameSite/Secure jika mode production

## 13. Catatan Produksi

- `air` hanya untuk development, bukan production.
- `.env` lokal tidak boleh di-commit.
- `.env.example` hanya template.
- Super admin pertama dibuat lewat bootstrap command, bukan endpoint publik.
- Password bootstrap production sebaiknya berasal dari secret manager dan segera dirotasi setelah akun siap.
- Jangan simpan access token atau refresh token di log.
- Swagger membantu testing, tetapi bukan pengganti test otomatis.
