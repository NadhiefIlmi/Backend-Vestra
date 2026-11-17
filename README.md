# Vestra Backend — AR Virtual Try-On API

Vestra Backend adalah REST API berperforma tinggi yang dibangun menggunakan **Golang (Standard Library)**. Backend ini bertanggung jawab untuk mengelola **metadata katalog pakaian, konfigurasi anchor point AR, autentikasi token Decart AI, dan pengiriman URL aset busana** ke frontend.

> 🔗 **Repositori Terkait:** Antarmuka web pengguna (*frontend*) untuk sistem ini tersedia di [**Frontend-Vestra (React 18 + Vite + Decart AI)**](https://github.com/NadhiefIlmi/Frontend-Vestra).

> 🔒 **JAMINAN PRIVASI (CLIENT-SIDE ONLY):**
> Sesuai arsitektur privasi PRD, backend ini **tidak pernah** menerima, memproses, atau menyimpan video stream kamera, foto snapshot, maupun data landmark koordinat tubuh pengguna. Seluruh pemrosesan video dan inferensi AR sepenuhnya diisolasi di sisi browser pengguna.

---

## 🚀 Fitur Utama

- **Clean Architecture:** Pemisahan tanggung jawab yang jelas antara `domain`, `service`, `repository`, dan `handler`.
- **Decart AI Token Gateway:** Endpoint server-side proxy untuk membuat *client token* Decart AI secara aman tanpa mengekspos API key di frontend.
- **Catalog Management:** Penyajian katalog busana dengan filter kategori (*top, bottom, outerwear*), pagination, dan dukungan `MemoryRepository` tanpa kebutuhan database saat pengembangan.
- **Privacy by Design:** Backend tidak menyentuh data kamera, landmark, atau frame pengguna sama sekali — seluruh inferensi AR berjalan lokal di browser.
- **Security Hardening:** CORS whitelist ketat, Panic Recovery, SecurityHeaders OWASP (`X-Frame-Options`, `X-XSS-Protection`, `Content-Security-Policy`), Body Limit 1MB, dan Timeout per-request.
- **Structured Logging:** Logging berbasis `log/slog` (bawaan Go 1.21+) dengan format JSON / Text dan redaksi otomatis terhadap secret.
- **100% Test Coverage:** Unit test, HTTP integration test, dan middleware test menggunakan `net/http/httptest`.

---

## 🛠️ Tech Stack

| Komponen | Pilihan Teknologi | Versi / Keterangan |
|---|---|---|
| **Bahasa Pemrograman** | Go (Golang) | `1.22+` (Standard Library first, tanpa framework berat) |
| **HTTP Router** | `net/http` Standard Library | Router idiomatic Go murni tanpa dependensi eksternal |
| **Structured Logger** | `log/slog` | Logger bawaan Go 1.21+, format JSON / Text, redaksi secret |
| **Database Driver** | `database/sql` Standard Library | Interface standar Go untuk PostgreSQL dengan connection pool |
| **Object Storage** | MinIO / AWS S3 Compatible | Hosting aset pakaian PNG transparan dan WebP |
| **Generative AI** | Decart AI API (`api.decart.ai`) | Token gateway untuk Lucy V-TON WebRTC video stream |
| **Spesifikasi Kontrak** | OpenAPI 3.0 (`api/openapi.yaml`) | Dokumentasi REST API formal |
| **Testing Suite** | Go `testing` + `net/http/httptest` | 100% test coverage pada handler, service, repository, middleware |

---

## 📁 Struktur Direktori & Penjelasan File

Struktur proyek mengadopsi prinsip **Clean Architecture** (decoupling ketat antara domain bisnis, persistensi data, dan protokol pengiriman):

```text
backend/
├── cmd/
│   └── api/
│       └── main.go               # Entry point: bootstrapping konfigurasi, database, routes, middleware, graceful shutdown
├── api/
│   └── openapi.yaml              # Spesifikasi formal OpenAPI 3.0 REST API kontrak
├── migrations/
│   └── 20261107120000_create_clothing_items.sql # Skema DDL tabel, indeks, dan data seed PostgreSQL
├── internal/
│   ├── config/                   # Pengelolaan konfigurasi environment
│   │   ├── config.go             # Loader, validasi wajib, parser durasi, dan masking secret logging
│   │   └── config_test.go        # Unit test konfigurasi development dan production
│   ├── handler/                  # Fondasi HTTP response standard
│   │   ├── response.go           # Standard JSON envelope: WriteSuccess dan WriteError
│   │   └── response_test.go      # Unit test HTTP response format dan headers
│   ├── middleware/               # HTTP Middlewares pipeline
│   │   ├── middleware.go         # RequestID, Slog Logging, Panic Recovery, CORS, Timeout, SecurityHeaders, BodyLimit
│   │   └── middleware_test.go    # Unit test seluruh middleware pipeline
│   ├── platform/
│   │   └── database/             # Platform adapters
│   │       ├── db.go             # Inisialisasi koneksi pool PostgreSQL dengan ping timeout
│   │       └── db_test.go        # Unit test kegagalan koneksi database
│   ├── decart/                   # Domain Decart AI Token Gateway
│   │   └── handler.go            # HTTP handler POST /api/v1/decart/token — server-side proxy Decart API
│   └── catalog/                  # Domain Katalog Busana (Feature Module)
│       ├── model/
│       │   └── clothing.go       # Struct entitas ClothingItem, AssetMetadata, AnchorConfig, dan OffsetConfig
│       ├── repository/
│       │   ├── repository.go     # Interface Repository, MemoryRepository (In-Memory + Seeding), dan PostgresRepository
│       │   └── repository_test.go# Unit test: query, category filter, pagination, dan context cancellation
│       ├── service/
│       │   ├── service.go        # Business logic: validasi kategori, aturan publikasi, dan pagination bounds
│       │   └── service_test.go   # Table-driven unit test service
│       └── handler/
│           ├── handler.go        # HTTP handlers: GET /api/v1/clothing dan GET /api/v1/clothing/{id}
│           └── handler_test.go   # HTTP integration test menggunakan net/http/httptest
├── bin/
│   └── api                       # Binary hasil kompilasi (di-ignore oleh Git)
├── .env.example                  # Template variabel environment
├── .env                          # Konfigurasi lokal lingkungan kerja (di-ignore oleh Git)
├── .gitignore                    # Daftar file yang diabaikan version control
└── go.mod                        # Definisi Go module & dependensi pustaka
```

### 📑 Rincian Fungsi Per-Modul & File

#### 1. `cmd/api/main.go` (Application Entrypoint)
Titik masuk utama aplikasi server HTTP. Bertugas membaca konfigurasi environment via `internal/config`, menginisialisasi `log/slog` structured logger, menyambungkan *connection pool* database PostgreSQL (atau fallback `MemoryRepository`), merangkai dependency injection (`Repository` ➔ `Service` ➔ `Handler`), mendaftarkan rute, memasang middleware pipeline, serta mengelola *Graceful Shutdown* saat menerima sinyal OS (`SIGINT`, `SIGTERM`).

#### 2. `internal/config/` (Configuration Management)
* **`config.go`**: Loader konfigurasi dari environment variables / file `.env`. Memvalidasi variabel wajib saat startup, mem-parse durasi (`REQUEST_TIMEOUT`), dan melakukan *masking redaksi* otomatis terhadap `DB_PASSWORD` dan `STORAGE_SECRET_KEY` sebelum dicatat ke log.
* **`config_test.go`**: Unit test validasi parsing konfigurasi untuk mode `development` dan `production`.

#### 3. `internal/handler/` (HTTP Response Foundation)
* **`response.go`**: Helper utilitas `WriteSuccess` dan `WriteError` untuk membungkus semua respons HTTP dalam format standar JSON envelope yang konsisten.
* **`response_test.go`**: Unit test format respons, status code, dan Content-Type header.

#### 4. `internal/middleware/` (Security & Request Pipeline)
* **`middleware.go`**: Pipeline middleware berurutan: `RequestID` (UUID tracking `X-Request-ID`) → `SecurityHeaders` (proteksi OWASP) → `Logging` (slog per-request) → `Recovery` (panic handler) → `CORS` (whitelist origin) → `Timeout` (context deadline) → `BodyLimit` (maks 1MB).
* **`middleware_test.go`**: Unit test seluruh komponen middleware pipeline secara terisolasi.

#### 5. `internal/decart/` (Decart AI Token Gateway)
* **`handler.go`**: Handler HTTP `POST /api/v1/decart/token`. Bertindak sebagai *server-side proxy* yang menerima request dari frontend, membuat client token ke `api.decart.ai/v1/client/tokens` menggunakan `DECART_API_KEY`, lalu meneruskan response tokennya ke client. Memastikan API key tidak pernah terekspos di bundle JavaScript frontend.

#### 6. `internal/catalog/` (Clothing Catalog Domain)
* **`model/clothing.go`**: Mendefinisikan struct entitas `ClothingItem` (ID, nama, deskripsi, kategori, URL aset), `AssetMetadata` (anchor X/Y, referenceWidth, scaleMultiplier, rotationOffset), dan `OffsetConfig`.
* **`repository/repository.go`**: Interface `Repository` dengan dua implementasi: `MemoryRepository` (in-memory + seed data busana default untuk dev tanpa DB) dan `PostgresRepository` (query SQL teroptimasi dengan filter kategori dan pagination).
* **`service/service.go`**: Logika bisnis validasi kategori, aturan publikasi (`published` only), dan sanitasi pagination bounds (page minimum 1, limit 1–100).
* **`handler/handler.go`**: HTTP handler `GET /api/v1/clothing` (list dengan filter `?category=&page=&limit=`) dan `GET /api/v1/clothing/{id}` (detail per ID).

#### 7. `internal/platform/database/` (Database Platform Adapter)
* **`db.go`**: Menginisialisasi connection pool PostgreSQL menggunakan `database/sql` dengan konfigurasi `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, dan verifikasi ping timeout.
* **`db_test.go`**: Unit test skenario kegagalan koneksi database.

#### 8. `migrations/` (Database Schema)
* **`20261107120000_create_clothing_items.sql`**: DDL lengkap tabel `clothing_items` dengan kolom anchor, metadata aset, indeks performa (`idx_clothing_status`, `idx_clothing_category`), dan data seed busana default.

---

## ⚙️ Persyaratan Sistem

- **Go:** Versi 1.22 atau lebih baru
- **PostgreSQL:** Versi 15+ *(Opsional — backend otomatis beralih ke `MemoryRepository` jika DB tidak tersedia)*
- **Git:** Untuk manajemen version control

---

## 🚀 Panduan Menjalankan

### 1. Kloning & Masuk ke Direktori Backend
```bash
cd backend
```

### 2. Pasang Dependensi Go Module
```bash
go mod download
```

### 3. Konfigurasi Lingkungan (`.env`)
Salin file template `.env.example` menjadi `.env` dan sesuaikan nilainya:
```bash
cp .env.example .env
```

### 4. Menjalankan Server API
```bash
go run ./cmd/api/
```
Server akan aktif di: **`http://localhost:8080`**

### 5. Membangun Binary Produksi
```bash
# Kompilasi binary ke folder bin/
go build -o ./bin/api ./cmd/api/

# Menjalankan binary hasil kompilasi
./bin/api
```

---

## ⚙️ Konfigurasi Environment (`.env`)

| Variabel | Contoh Nilai | Wajib? | Keterangan |
|---|---|---|---|
| `APP_ENV` | `development` | Ya | Pilihan: `development`, `test`, `production` |
| `HTTP_PORT` | `8080` | Ya | Port listen HTTP server |
| `REQUEST_TIMEOUT` | `15s` | Tidak | Batas timeout per-request (default: `15s`) |
| `DATABASE_URL` | `postgres://user:pass@localhost:5432/vestra_dev?sslmode=disable` | Tidak | Connection string PostgreSQL (opsional saat dev) |
| `DB_MAX_OPEN_CONNS` | `25` | Tidak | Maks koneksi terbuka di connection pool |
| `DB_MAX_IDLE_CONNS` | `10` | Tidak | Maks koneksi idle di connection pool |
| `DB_CONN_MAX_LIFETIME` | `5m` | Tidak | Masa aktif maksimum koneksi database |
| `STORAGE_ENDPOINT` | `http://localhost:9000` | Tidak | Endpoint MinIO / S3 compatible storage |
| `STORAGE_BUCKET` | `catalog` | Tidak | Nama bucket penyimpanan aset busana |
| `STORAGE_ACCESS_KEY` | `minioadmin` | Tidak | Access key storage |
| `STORAGE_SECRET_KEY` | `minioadmin` | Tidak | Secret key storage *(di-redact saat logging)* |
| `DECART_API_KEY` | `sk-decart-xxxx` | Ya | API key Decart AI untuk token gateway |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:3000` | Ya | Origin frontend yang diizinkan (dipisah koma) |

> 🔒 **Keamanan:** `internal/config` secara otomatis me-redact `STORAGE_SECRET_KEY` dan `DATABASE_URL` password sebelum dicatat ke log.

---

## 🧪 Pengujian & Kualitas Kode

Jalankan perintah berikut di dalam direktori `backend/`:

```bash
# Menjalankan seluruh test suite unit & HTTP handler
go test -v ./...

# Menjalankan test secara segar tanpa cache
go test -count=1 ./...

# Memeriksa static analysis dan code correctness
go vet ./...

# Memeriksa code formatting standar Go
go fmt ./...

# Kompilasi binary executable
go build -o ./bin/api ./cmd/api/
```

---

## 📡 Daftar Endpoint API

Semua endpoint API disajikan dengan basis path `/api/v1`:

| Modul | Method | Endpoint | Deskripsi | Auth |
|---|:---:|---|---|---|
| **Health** | `GET` | `/health` | Liveness probe — status server aktif | Public |
| **Readiness** | `GET` | `/ready` | Readiness probe — status DB & Storage | Public |
| **Catalog** | `GET` | `/api/v1/clothing` | List busana `published` dengan filter `?category=&page=&limit=` | Public |
| **Catalog** | `GET` | `/api/v1/clothing/{id}` | Detail metadata 1 item busana berdasarkan ID | Public |
| **Decart AI** | `POST` | `/api/v1/decart/token` | Generate client token Decart AI (server-side proxy) | Public |

### Contoh Request & Response

#### List Busana (`GET /api/v1/clothing?category=top&page=1&limit=20`)
**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "shirt-basic-black",
      "name": "Classic Black T-Shirt",
      "category": "top",
      "thumbnailUrl": "/assets/clothing/shirt-black-thumb.webp",
      "assetUrl": "/assets/clothing/shirt-black.svg",
      "assetMetadata": {
        "anchor": { "x": 0.5, "y": 0.16 },
        "referenceWidth": 1000,
        "scaleMultiplier": 1.0,
        "rotationOffset": 0,
        "offset": { "x": 0, "y": 0 }
      },
      "status": "published"
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 3 }
}
```

#### Decart Token (`POST /api/v1/decart/token`)
**Response `200 OK`:**
```json
{
  "token": "eyJhbGc...",
  "expiresAt": "2026-11-17T16:51:00Z"
}
```

---

## 🔐 Keamanan & Middleware

Pipeline middleware dieksekusi secara berurutan pada setiap *incoming request*:

1. **`RequestID`:** Menyematkan UUID unik (`X-Request-ID`) pada context & header response untuk audit log terdistribusi.
2. **`SecurityHeaders`:** Menginjeksi proteksi browser OWASP (`X-Content-Type-Options`, `X-Frame-Options: DENY`, `X-XSS-Protection`, `Referrer-Policy`, `Content-Security-Policy`).
3. **`Logging`:** Mencatat HTTP method, path, status, remote IP, dan latency setiap request menggunakan `log/slog`.
4. **`Recovery`:** Menangkap *panic runtime* secara elegan dan mengembalikan `500 INTERNAL_ERROR` aman tanpa membocorkan stack trace.
5. **`CORS`:** Mengatur whitelist origin frontend (`CORS_ALLOWED_ORIGINS`) dan menangani preflight `OPTIONS`.
6. **`Timeout`:** Membatasi waktu eksekusi handler via `context.WithTimeout` mencegah *hanging request*.
7. **`BodyLimit`:** Membatasi payload request maksimal **1 MB** mencegah serangan DoS.

---

## 🛠️ Troubleshooting & Solusi

### 1. Error: `DECART_API_KEY is not configured`
- **Penyebab:** Variabel `DECART_API_KEY` belum diisi di file `.env`.
- **Solusi:** Tambahkan API key Decart AI Anda ke `.env`: `DECART_API_KEY=sk-decart-xxxxxx`, lalu restart server.

### 2. Error: `CORS policy: No 'Access-Control-Allow-Origin' header`
- **Penyebab:** Origin frontend belum terdaftar di `CORS_ALLOWED_ORIGINS`.
- **Solusi:** Tambahkan origin frontend ke `.env`: `CORS_ALLOWED_ORIGINS=http://localhost:5173`, lalu restart server.

### 3. Error: `failed to open database: unknown driver "postgres"`
- **Penyebab:** `DATABASE_URL` tidak valid atau PostgreSQL belum menyala.
- **Solusi:** Backend otomatis beralih ke `MemoryRepository` jika DB tidak tersedia saat development. Jika menggunakan DB asli, pastikan connection string valid dan service PostgreSQL aktif.

### 4. Error saat kompilasi `go build`
- **Solusi:** Jalankan `go mod tidy` dan `go mod download` untuk memastikan seluruh dependencies sinkron dengan `go.mod`.

---

## 🔗 Ekosistem Repositori Vestra

| Komponen | Repositori GitHub | Deskripsi & Stack |
|---|---|---|
| **Backend REST API** | [Backend-Vestra](https://github.com/NadhiefIlmi/Backend-Vestra) | REST API (Golang, Standard Library, PostgreSQL, Decart AI Token Gateway) |
| **Frontend Web App** | [Frontend-Vestra](https://github.com/NadhiefIlmi/Frontend-Vestra) | AR Virtual Try-On (React 18, TypeScript, Vite, Decart AI SDK, LiveKit) |
