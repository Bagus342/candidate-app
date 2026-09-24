# DevOps Technical Test

Repository berisi *source code*, konfigurasi Docker, pipeline CI/CD, dan dokumentasi infrastruktur untuk menyelesaikan DevOps Technical Test.

---

## Task 1 (LINUX SERVER)

### 1. SSH Security
Konfigurasi file `/etc/ssh/sshd_config` telah disesuaikan agar:
- `PermitRootLogin no` (Direct SSH login sebagai root dinonaktifkan).
- `PasswordAuthentication no` (Login SSH menggunakan password dinonaktifkan).
- `PubkeyAuthentication yes` (Login menggunakan SSH Public Key diaktifkan).

**Pembuktian:**
- *Login SSH Key Berhasil:* Mengeksekusi `ssh devops@<IP_VM>` langsung masuk tanpa password.
- *Login Password Ditolak:* Mencoba SSH tanpa private key yang valid akan menghasilkan error `Permission denied (publickey)`.
- *Login Root Ditolak:* `ssh root@<IP_VM>` langsung menghasilkan `Permission denied (publickey)`.

**Kesalahan:**
- Mengira kalau login ssh via root, jadi stuck konfigurasi ssh key.

### 2. User & Sudo
- **User:** Dibuat user baru bernama `devops`.
- **Group:** User ditambahkan ke grup `sudo` (agar dapat menjalankan command administratif) dan grup `docker` (agar bisa menjalankan container tanpa root).
- **Alasan Pembatasan Root:** Membatasi akses root langsung sangat krusial untuk mencegah serangan *brute-force* yang selalu menargetkan user `root`. Selain itu, menggunakan user terpisah memberikan jejak audit (*audit trail*) yang jelas di log sistem mengenai siapa yang mengeksekusi sebuah command.
- **Cara Membatasi Privilege (Full Sudo):** Jika user tidak butuh full sudo, konfigurasinya bisa dibatasi melalui `visudo` (file `/etc/sudoers.d/devops`). Contoh: `devops ALL=(ALL) NOPASSWD: /bin/systemctl restart candidate-app.service` (hanya bisa restart service tertentu saja).

### 3. Firewall
Konfigurasi menggunakan UFW (*Uncomplicated Firewall*).
- **Status & Rule Aktif:**
  `sudo ufw status verbose`
  Hasil *Output*: Port 22, 80, 443, 8080 (TCP) berstatus ALLOW. Default incoming rule adalah DENY (menolak koneksi lain yang tidak diperlukan).
- **Cara Cek Port Listening:**
  `sudo ss -tuln` atau `sudo netstat -tuln`

### 4. Linux Service
Systemd service telah dibuat di `/etc/systemd/system/candidate-app.service` untuk mengelola aplikasi di `/opt/candidate/app.sh`.
- Konfigurasi kunci pada unit file:
  - `ExecStart=/opt/candidate/app.sh`
  - `Restart=always` (Restart otomatis jika aplikasi crash).
  - `WantedBy=multi-user.target` (Start otomatis saat VM boot).

**Command yang Digunakan:**
- Menjalankan service: `sudo systemctl start candidate-app`
- Autostart saat boot: `sudo systemctl enable candidate-app`
- Cek status service: `sudo systemctl status candidate-app`
- Memeriksa log aplikasi: `sudo journalctl -u candidate-app -f`

---

## Task 2

### 1. Cara Menjalankan Application (Tanpa Docker)
Aplikasi ini dibangun menggunakan Go. Untuk menjalankannya secara lokal di mesin Anda:
1. Pastikan Go (minimal v1.22) sudah terinstal.
2. Jalankan perintah untuk mengunduh dependensi:
   `go mod download`
3. Jalankan aplikasi:
   `go run .`
   *(Atau lakukan build menggunakan `go build -o bin/candidate-app .` lalu jalankan `./bin/candidate-app`)*

### 2. Cara Menjalankan Docker
Sistem ini menggunakan Docker Compose untuk menjalankan *Application* dan *PostgreSQL* dalam satu jaringan agar dapat saling berkomunikasi.
1. Salin file *environment*:
   `cp .env.example .env`
2. Isi nilai kredensial di file `.env`.
3. Jalankan container di *background*:
   `docker compose up -d --build`

### 3. Cara Menjalankan Deployment
Deployment dikonfigurasi berjalan sepenuhnya secara otomatis menggunakan **GitHub Actions**.
- **Trigger:** Pipeline akan berjalan secara otomatis ketika terdapat `push` ke branch `main`.
- **Alur CI/CD:** Pipeline melakukan checkout, setup Node/Go, menginstal dependensi, menjalankan test, mem-build Docker Image, dan menyimpannya ke GitHub Container Registry (GHCR).
- **Deployment:** Setelah test berhasil, runner masuk ke VM melalui SSH, melakukan `docker login` menggunakan token, menarik image terbaru, dan menjalankan ulang container (`docker compose up -d`).

### 4. Environment Variable yang Digunakan
Sesuai ketentuan keamanan, credential asli tidak dimasukkan ke dalam Git repository. Silakan merujuk pada file `.env.example`. Variabel utama yang digunakan:
- `APP_PORT` / `HTTP_ADDR`: Port yang digunakan aplikasi (contoh: `:8080`).
- `POSTGRES_USER`: Username database (contoh: `candidate_user`).
- `POSTGRES_PASSWORD`: Password database.
- `POSTGRES_DB`: Nama database.
- `DATABASE_URL`: Connection string PostgreSQL (contoh: `postgres://user:pass@postgres:5432/dbname?sslmode=disable`). **Penting:** Host harus diisi dengan nama service `postgres`, bukan `localhost`.

### 5. Cara Melakukan Troubleshooting
Jika aplikasi production mati atau tidak bisa diakses, berikut adalah langkah sistematisnya:
1. **Verifikasi Server & Service (Network/Firewall):** Pastikan port aplikasi terbuka dengan `sudo ufw status` dan aplikasi me-listen port dengan `sudo ss -tuln`.
2. **Cek Status Container:** Jalankan `docker ps -a` untuk melihat apakah container `app` berstatus *Exited* atau *Restarting*.
3. **Analisis Log (Root Cause):** Jalankan `docker compose logs app` untuk melihat pesan error dari dalam aplikasi (contoh: gagal konek database).
4. **Cek Database:** Pastikan koneksi aman dengan `docker compose exec postgres pg_isready -U <user> -d <db>`.
5. **Cek Resource:** Jalankan `./server-check.sh` (jika tersedia) untuk memeriksa lonjakan CPU, sisa Memory, atau Disk space yang penuh.

### 6. Keputusan Teknis Penting yang Dibuat
- **Docker Compose Image Tagging:** File compose menggunakan variabel dinamis `${APP_IMAGE:-candidate-app}` agar bisa digunakan untuk *local development* (build lokal) maupun *production* (pull dari GHCR).
- **Setup SSH Keyscan di CI/CD:** Untuk menghindari gagal verifikasi Host SSH saat deployment pertama kali di Actions, saya menggunakan konfigurasi `StrictHostKeyChecking accept-new` pada SSH runner, bukan sekadar mematikan security checking (`no`).
- **Pengecekan Jaringan (Task 5):** Pada bash script server monitoring, saya mengubah parameter `ping` ICMP menjadi `curl` HTTP request karena protokol ICMP seringkali diblokir oleh *cloud provider* demi alasan keamanan DDoS.