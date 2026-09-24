# Incident Report

1. **Problem:** 
   Setelah deployment via CI/CD, aplikasi web berstatus "Restarting" dan tidak dapat diakses.
2. **Investigation:** 
   Melakukan pengecekan log container menggunakan perintah `docker compose logs app` di VM. Ditemukan error: `failed to connect to user=candidate database=candidate-app: 127.0.0.1:5432 (localhost): dial error: connection refused`.
3. **Root Cause:** 
   Aplikasi mencoba terhubung ke database menggunakan alamat IP `127.0.0.1` (localhost). Di dalam jaringan Docker (Docker network), localhost merujuk pada container aplikasi itu sendiri, bukan container database.
4. **Fix:** 
   Mengubah nilai host pada variabel `DATABASE_URL` di dalam file `.env` VM, dari `localhost` menjadi nama service database Docker yaitu `postgres`.
5. **Verification:** 
   Melakukan restart container dengan `docker compose down` dan `docker compose up -d`, lalu memastikan log aplikasi menampilkan status berhasil terkoneksi ke database.
6. **Prevention:** 
   Menambahkan dokumentasi yang jelas di file `.env.example` bahwa koneksi database di dalam Docker harus menggunakan nama service, serta menyederhanakan CI/CD untuk mencegah *race condition* infrastruktur.