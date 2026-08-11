# Issue Log

Catatan issue tertunda dan selesai untuk project helpdesk ticketing. Gunakan file ini sebagai reminder sebelum commit, push, atau deploy.

## Pending

### 2026-08-08 00:10:03 +08:00 - Force push rewrite history untuk menghapus guide dari GitHub

- Status: pending
- Issue: `helpdesk-step-by-step-guide.md` sudah dihapus dari Git tracking dan local history sudah di-rewrite, tetapi push ke GitHub gagal karena host `github.com` tidak bisa di-resolve.
- Dampak: GitHub remote kemungkinan masih menyimpan commit lama yang pernah berisi `helpdesk-step-by-step-guide.md`.
- Perintah yang perlu dijalankan saat koneksi GitHub normal:

```bash
git push --force-with-lease origin main
```

- Verifikasi setelah push:

```bash
git log --all -- helpdesk-step-by-step-guide.md
git ls-files -- helpdesk-step-by-step-guide.md
```

- Catatan keamanan: jika file pernah berisi secret/API key/password, secret tetap harus di-rotate walaupun history sudah dibersihkan.

## Done

Belum ada issue selesai yang dicatat.
