# Seeding Guide Backend Helpdesk

Seeder mengisi database development/demo dengan data siap pakai. Seeder berbeda dari migration.

```text
Migration -> membuat struktur database
Seeder    -> mengisi data demo
```

Seeder project ini dibuat sebagai command terpisah:

```text
backend/cmd/seed/main.go
```

## 1. Kapan Seeder Dipakai

Gunakan seeder untuk:

- demo dashboard agar tidak kosong
- manual testing lewat Swagger/frontend
- membuat akun demo semua role
- mengisi category dan SLA rule awal
- membuat ticket dengan variasi status, priority, requester, assignee, dan tanggal

Jangan gunakan seeder untuk production runtime. Command seed akan menolak jalan jika:

```env
APP_ENV=production
```

## 2. Data yang Dibuat

Seeder membuat user demo:

| Email | Password | Role |
| --- | --- | --- |
| `superadmin@test.local` | `password123` | `super_admin` |
| `admin@test.local` | `password123` | `admin` |
| `budi@test.local` | `password123` | `staff` |
| `sari@test.local` | `password123` | `staff` |
| `andi@test.local` | `password123` | `staff` |
| `rina@test.local` | `password123` | `staff` |
| beberapa `end_user` | `password123` | `end_user` |

Seeder juga membuat:

- 4 category: `Hardware`, `Software`, `Network`, `Access Request`
- SLA rule untuk priority `Low`, `Medium`, `High`, `Critical`
- 80 ticket demo dengan title prefix `[DEMO]`
- ticket comments
- ticket status log
- ticket activity log

## 3. Cara Menjalankan

Pastikan backend `.env` mengarah ke database development:

```env
APP_ENV=development
DATABASE_URL=postgresql://helpdesk_user:helpdesk_pass@localhost:5432/helpdesk_db?sslmode=disable
```

Jalankan:

```powershell
cd backend
go run ./cmd/seed
```

Jika sukses:

```text
seed complete
```

Setelah itu login memakai akun demo, misalnya:

```json
{
  "email": "superadmin@test.local",
  "password": "password123"
}
```

## 4. Sifat Seeder

Seeder bersifat idempotent untuk:

- user demo
- category
- SLA rule

Artinya data tersebut akan dibuat jika belum ada, atau diupdate jika sudah ada.

Untuk ticket demo, seeder akan menghapus ticket dengan prefix:

```text
[DEMO]
```

Lalu membuat ulang 80 ticket demo. Tujuannya agar menjalankan seeder berkali-kali tidak membuat data ticket demo menumpuk.

Jangan membuat ticket penting/manual dengan title prefix `[DEMO]`, karena akan ikut dihapus oleh seeder.

## 5. Relasi dengan Test

Seeder ini berbeda dari integration test seed.

```text
cmd/seed                  -> untuk database development/demo
backend/tests setup seed  -> untuk PostgreSQL test database
```

Integration test tetap memakai `TEST_DATABASE_URL` dan membuat seed sendiri agar test terisolasi.

## 6. Catatan Best Practice

- Seeder tidak boleh otomatis dipanggil saat server start.
- Seeder tidak boleh dipakai untuk data production dummy.
- Password demo hanya untuk development.
- Data demo harus mudah dikenali, karena itu ticket memakai prefix `[DEMO]`.
- Jika schema berubah, update seeder dan dokumentasi ini bersama migration.
