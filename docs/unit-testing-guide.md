# Unit Testing Guide Backend Helpdesk

Dokumen ini menjelaskan unit test yang sudah dibuat, alasan pemilihannya, cara menjalankannya, dan gap yang sengaja belum masuk unit test.

## 1. Tujuan Unit Test

Unit test dipakai untuk mengetes function kecil secara langsung tanpa menjalankan HTTP server dan tanpa database nyata.

Target unit test pertama di project ini:

- business rule penting
- input dan output jelas
- tidak butuh network
- tidak butuh PostgreSQL
- cepat dijalankan di local dan CI

Unit test berbeda dari integration/API test:

| Jenis test | Yang dites | Butuh server HTTP | Butuh database |
| --- | --- | --- | --- |
| Unit test | function kecil/helper/service logic murni | tidak | tidak |
| Integration/API test | endpoint Gin, middleware, JSON response, RBAC route | biasanya pakai test router | sering butuh test DB/fake repository |
| Repository integration test | SQL dan transaksi repository | tidak | ya |

## 2. Unit Test yang Sudah Ada

### `backend/internal/service/ticket_service_test.go`

Mengetes:

```go
TransitionAllowed(oldStatus, newStatus string) bool
```

Kenapa dites pertama:

- function murni
- tidak butuh DB
- output hanya `true` atau `false`
- rule ticket workflow penting untuk mencegah status loncat sembarangan

Skenario yang dites:

- `Open -> In Progress` valid
- `Open -> Resolved` invalid
- `In Progress -> Pending` valid
- `In Progress -> Resolved` valid
- `Pending -> In Progress` valid
- `Pending -> Resolved` valid
- `Resolved -> Closed` valid
- `Resolved -> Open` valid
- `Closed -> Open` valid
- status unknown/empty invalid

### `backend/internal/service/sla_service_test.go`

Mengetes:

```go
SLAState(deadline *time.Time, totalHours *int, status string, now time.Time) string
```

Skenario yang dites:

- deadline kosong menghasilkan `ok`
- ticket `Resolved` atau `Closed` selalu `ok`
- deadline lewat menghasilkan `breached`
- sisa waktu kurang/sama dengan 20% menghasilkan `at_risk`
- sisa waktu lebih dari 20% menghasilkan `ok`
- total SLA kosong atau nol tidak dihitung `at_risk`

### `backend/internal/service/auth_service_test.go`

Mengetes:

```go
HashPassword(password string) (string, error)
CheckPassword(hash, password string) bool
AccessToken(user *models.User) (string, error)
HashRefreshToken(plain string) string
NewRefreshToken() (plain string, hash string, expiresAt time.Time, err error)
```

Skenario yang dites:

- password hash tidak kosong
- password hash tidak sama dengan password mentah
- password benar lolos `CheckPassword`
- password salah gagal `CheckPassword`
- access token valid dan punya claim `sub`, `role`, dan `exp`
- refresh token hash deterministic untuk input dan pepper yang sama
- refresh token hash berubah jika pepper berubah
- refresh token baru punya plain token, hash, dan expiry sekitar 7 hari

### `backend/internal/repository/ticket_repository_test.go`

Mengetes:

```go
EnsureTicketAccess(t *models.Ticket, userID int64, role string) error
```

Walaupun file-nya berada di package repository, function ini adalah pure access helper dan tidak menyentuh DB.

Skenario yang dites:

- `super_admin` bisa akses semua ticket
- `admin` bisa akses semua ticket
- `staff` hanya bisa akses ticket yang di-assign kepadanya
- `end_user` hanya bisa akses ticket miliknya
- role tidak dikenal ditolak
- staff ditolak jika ticket belum punya assignee

### `backend/internal/handlers/user_handler_test.go`

Mengetes:

```go
validUserRole(role string) bool
canManageRole(actorRole, targetRole string) bool
```

Skenario yang dites:

- role valid: `super_admin`, `admin`, `staff`, `end_user`
- role invalid ditolak
- `super_admin` boleh memberi semua role
- `admin` hanya boleh memberi `staff` dan `end_user`
- `admin` tidak boleh memberi `admin` atau `super_admin`
- `staff`, `end_user`, dan role unknown tidak boleh memberi role

## 3. Cara Menjalankan Unit Test

Dari folder backend:

```powershell
cd backend
go test ./internal/service ./internal/repository ./internal/handlers
```

Menjalankan semua test backend:

```powershell
cd backend
go test ./...
```

Menjalankan satu package dengan output detail:

```powershell
cd backend
go test ./internal/service -v
```

Menjalankan satu test tertentu:

```powershell
cd backend
go test ./internal/service -run TestTransitionAllowed -v
```

## 4. Cara Membaca Pola Table-Driven Test

Sebagian besar test memakai pola table-driven test:

```go
tests := []struct {
	name string
	input string
	want bool
}{
	{name: "case name", input: "value", want: true},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		got := function(tt.input)
		if got != tt.want {
			t.Fatalf("got %v, want %v", got, tt.want)
		}
	})
}
```

Cara bacanya:

- `tests` berisi daftar skenario.
- `name` adalah nama skenario yang muncul di output test.
- `want` adalah hasil yang diharapkan.
- `t.Run` membuat setiap skenario terlihat terpisah.
- `t.Fatalf` menghentikan skenario jika hasil tidak sesuai.

Pola ini cocok untuk rule seperti status transition, SLA state, dan RBAC policy karena banyak kombinasi input-output.

## 5. Kenapa Belum Semua Endpoint Dites

Endpoint seperti:

```text
POST /api/auth/login
GET /api/tickets
PATCH /api/tickets/{id}/status
DELETE /api/categories/{id}
```

lebih tepat disebut integration/API test, bukan unit test.

Alasannya:

- perlu router Gin
- perlu request HTTP palsu
- perlu middleware auth dan role
- sering perlu database test atau fake repository
- perlu cek status code dan response JSON

Jadi endpoint test adalah tahap berikutnya setelah unit test dasar stabil.

## 6. Gap yang Masih Ada

Unit test saat ini belum mencakup:

- handler endpoint penuh dengan `httptest`
- repository SQL dan transaction dengan PostgreSQL test database
- middleware auth/rate-limit dalam request HTTP nyata
- dashboard/report query dengan data dummy
- CI/CD yang menjalankan test otomatis

Ini bukan bug unit test, tetapi batas scope. Untuk menutup gap CV berikutnya, tahap lanjut yang disarankan:

1. Tambahkan GitHub Actions untuk `go test ./...`, `go vet ./...`, dan `go build ./cmd/server`.
2. Perluas integration test endpoint auth dan RBAC memakai `httptest`.
3. Tambahkan repository integration test dengan test database untuk transaksi yang lebih detail.

Panduan integration test otomatis ada di `docs/integration-testing-guide.md`.

## 7. Standar Penambahan Unit Test Baru

Tambahkan unit test baru jika function memenuhi salah satu kondisi:

- punya business rule penting
- output bisa diprediksi dari input
- pernah menjadi sumber bug
- menyangkut security atau authorization
- menyangkut kalkulasi SLA/status/report

Jangan memaksakan unit test untuk function yang tugasnya hanya meneruskan request ke database. Untuk function seperti itu, biasanya integration test lebih bernilai.

## 8. Checklist Sebelum Menganggap Unit Test Selesai

- Test bisa jalan dengan `go test ./...`.
- Test tidak bergantung pada urutan eksekusi.
- Test tidak butuh server backend hidup.
- Test tidak butuh database lokal.
- Nama skenario jelas.
- Ada case sukses dan gagal.
- Security rule seperti role dan access check punya negative case.
- Dokumentasi test diperbarui jika ada test baru.
