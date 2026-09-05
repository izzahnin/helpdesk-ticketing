# Integration Testing Guide Backend Helpdesk

Dokumen ini menjelaskan integration/API test otomatis untuk backend Helpdesk. Test ini bukan test manual Swagger/Postman, tetapi test kode Go yang dijalankan dengan `go test`.

## 1. Apa Itu Integration/API Test Otomatis

Integration/API test otomatis mengetes beberapa komponen sekaligus:

```text
HTTP request -> Gin router -> middleware -> handler -> service -> repository -> PostgreSQL test database
```

Bedanya dengan unit test:

| Jenis test | Cara kerja | Contoh |
| --- | --- | --- |
| Unit test | panggil function langsung | `TransitionAllowed("Open", "Resolved")` |
| Integration/API test | kirim HTTP request palsu ke router Gin | `POST /api/auth/register` |
| Manual test | klik Swagger/Postman/browser | coba endpoint satu per satu |

Integration test otomatis tetap berjalan dari terminal:

```powershell
cd backend
go test ./...
```

Tidak perlu membuka Swagger, Postman, atau browser.

## 2. Test Database

Integration test backend ini memakai PostgreSQL test database khusus melalui env:

```env
TEST_DATABASE_URL=postgresql://helpdesk_user:helpdesk_pass@localhost:5432/helpdesk_test_db?sslmode=disable
```

Nama database wajib mengandung kata:

```text
test
```

Guard ini sengaja dibuat agar test tidak tidak sengaja menjalankan reset/drop table ke database development atau production.

Contoh membuat database test lokal:

```powershell
createdb -U helpdesk_user -h localhost helpdesk_test_db
```

Atau lewat psql:

```powershell
psql -U postgres -h localhost
CREATE DATABASE helpdesk_test_db OWNER helpdesk_user;
```

## 3. Cara Test Menyiapkan Data

File test:

```text
backend/tests/api_integration_test.go
```

Setiap test membuat app baru dengan alur:

```text
baca TEST_DATABASE_URL
cek nama database mengandung "test"
connect ke PostgreSQL
drop tabel test
run migration 001 dan 002
seed user dasar
reset in-memory rate limiter
buat Gin router asli
tembak endpoint dengan httptest
cek status code dan response
```

User seed yang dibuat:

| Email | Password | Role |
| --- | --- | --- |
| `super@test.local` | `password123` | `super_admin` |
| `admin@test.local` | `password123` | `admin` |
| `staff@test.local` | `password123` | `staff` |
| `requester@test.local` | `password123` | `end_user` |
| `other.user@test.local` | `password123` | `end_user` |

## 4. Skenario yang Sudah Dites

### Auth dan Role

Test:

```go
TestAuthAndRoleIntegration
```

Skenario:

- `POST /api/auth/register` membuat user baru.
- User register selalu mendapat role `end_user`.
- Admin biasa tidak boleh promote user menjadi `super_admin`.
- Super admin boleh promote user menjadi `admin`.

### Ticket Workflow

Test:

```go
TestTicketWorkflowIntegration
```

Skenario:

- End user login.
- Super admin login.
- Staff login.
- Super admin membuat category.
- End user submit ticket.
- Ticket baru berstatus `Open`.
- Super admin assign ticket ke staff.
- Staff update status ke `In Progress`.
- Staff menambahkan komentar.
- End user lain ditolak saat membuka ticket yang bukan miliknya.

### Dashboard dan Report

Test:

```go
TestDashboardAndReportIntegration
```

Skenario:

- Super admin boleh akses dashboard summary.
- Staff ditolak dari dashboard admin.
- Staff boleh akses `my-performance`.
- Staff ditolak jika mengirim `staff_id` ke `my-performance`.
- Super admin boleh export CSV.
- Response report punya `Content-Type` CSV dan header kolom yang benar.

## 5. Cara Menjalankan

Menjalankan semua test:

```powershell
cd backend
go test ./...
```

Menjalankan hanya integration test:

```powershell
cd backend
go test ./tests -v
```

Menjalankan satu skenario:

```powershell
cd backend
go test ./tests -run TestTicketWorkflowIntegration -v
```

Jika `TEST_DATABASE_URL` belum diset, test integration akan `skip`, bukan gagal. Ini berguna agar unit test tetap bisa jalan di mesin yang belum punya PostgreSQL test database.

## 6. Kenapa Memakai `httptest`

Package `net/http/httptest` membuat request HTTP palsu tanpa membuka port server.

Flow sederhananya:

```go
req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
res := httptest.NewRecorder()

router.ServeHTTP(res, req)

if res.Code != http.StatusCreated {
	t.Fatalf("got %d, want %d", res.Code, http.StatusCreated)
}
```

Keuntungannya:

- cepat
- otomatis
- tidak perlu `air`
- tidak perlu Swagger/Postman
- bisa masuk CI/CD
- tetap mengetes route, middleware, handler, service, repository, dan SQL

## 6.1 Temuan dari Integration Test

Integration test boleh menemukan bug kecil yang tidak kelihatan dari unit test.

Contoh dari project ini:

- Response submit ticket harus mengembalikan status `Open`, bukan string kosong, karena kontrak API menjanjikan model ticket yang baru dibuat.
- Rate limiter in-memory menyimpan state global, jadi test perlu reset helper sebelum membuat app test baru agar skenario sebelumnya tidak memengaruhi skenario berikutnya.

## 7. Gap yang Masih Ada

Integration test saat ini belum mencakup semua endpoint dan semua edge case.

Gap berikutnya:

- refresh/logout cookie flow lebih detail
- endpoint category update/delete
- invalid JSON payload untuk semua handler
- invalid ticket status transition lewat API
- dashboard filter `from`, `to`, `category_id`, dan `staff_id`
- rate limit `429`
- readiness saat database down
- repository transaction rollback case

Prioritas lanjutan yang disarankan:

1. Tambah test invalid status transition.
2. Tambah test category update/delete.
3. Tambah test dashboard filter.
4. Tambah CI/CD agar `go test ./...` jalan otomatis di GitHub Actions.

## 8. Standar Project

Mulai sekarang, perubahan endpoint penting sebaiknya punya minimal satu dari dua test:

- unit test untuk business rule kecil
- integration/API test untuk behavior HTTP dan RBAC

Jika perubahan menyentuh role, auth, ticket access, atau report, wajib pikirkan test sukses dan test gagal.
