# Prism Common Libraries

[![Go CI Pipeline](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml/badge.svg)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Lumina-Enterprise-Solutions/prism-common-libs)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Repositori ini berisi kumpulan pustaka (libraries) dan utilitas bersama yang digunakan oleh semua microservices di dalam ekosistem **Prism ERP** oleh **Lumina Enterprise Solutions**. Tujuan utama dari repositori ini adalah untuk standarisasi kode, mengurangi duplikasi, dan menyederhanakan dependensi antar layanan.

---

## Modul yang Tersedia

-   `config`: Loader konfigurasi terpusat yang membaca dari *environment variables* dan *Consul KV*.
-   `consul`: Utilitas untuk pendaftaran layanan (service discovery) ke Consul.
-   `handlers`: Handler HTTP umum seperti `HealthCheckHandler`.
-   `jwt`: Middleware dan helper untuk validasi token JWT.
-   `model`: Definisi model data bersama seperti `User` dan `FileMetadata`.
-   `rbac`: Middleware untuk Role-Based Access Control (RBAC).
-   `telemetry`: Konfigurasi terpusat untuk OpenTelemetry (tracing).
-   `vault`: Klien sederhana untuk membaca *secrets* dari HashiCorp Vault.

---

## Cara Penggunaan

Repositori ini dimaksudkan untuk digunakan sebagai Go Module. Untuk menambahkannya sebagai dependensi pada layanan Anda:

1.  **Tambahkan ke dependensi Anda**:
    Jalankan perintah berikut di dalam direktori layanan Anda:
    ```bash
    go get github.com/Lumina-Enterprise-Solutions/prism-common-libs@v0.1.0
    ```
    *(Ganti `v0.1.0` dengan versi rilis terbaru yang ingin Anda gunakan)*

2.  **Impor dan gunakan modul**:
    Impor paket yang dibutuhkan di dalam kode Go Anda.

    ```go
    import (
        "github.com/gin-gonic/gin"
        "github.com/Lumina-Enterprise-Solutions/prism-common-libs/handlers"
        commonjwt "github.com/Lumina-Enterprise-Solutions/prism-common-libs/jwt"
    )

    func main() {
        router := gin.Default()

        // Contoh penggunaan health check handler
        handlers.SetupHealthRoutes(router, "my-service", "1.0.0")

        // Contoh penggunaan middleware JWT
        protected := router.Group("/api")
        protected.Use(commonjwt.JWTMiddleware())
        {
            // ... rute yang dilindungi ...
        }

        router.Run()
    }
    ```

---

## Kontribusi

Perubahan pada pustaka ini memiliki dampak luas ke semua layanan. Oleh karena itu, proses kontribusi diatur dengan ketat untuk menjaga kualitas dan stabilitas.

1.  **Buat Issue**: Diskusikan perubahan yang ingin Anda buat melalui GitHub Issues terlebih dahulu.
2.  **Fork dan Branch**: Buat *fork* dari repositori ini dan buat *branch* baru untuk perubahan Anda.
3.  **Implementasi & Test**: Implementasikan perubahan Anda dan pastikan semua tes lolos (`go test -v ./...`).
4.  **Buat Pull Request**: Buka Pull Request (PR) ke branch `main`.
5.  **Review**: PR Anda harus disetujui oleh **Code Owners** yang ditunjuk.
6.  **Merge**: Setelah disetujui dan semua *status checks* berhasil, PR akan di-*squash and merge* oleh maintainer.

## Rilis (Versioning)

Repositori ini mengikuti **Semantic Versioning**. Versi baru akan dirilis melalui GitHub Actions setiap kali sebuah tag baru dengan format `vX.Y.Z` (contoh: `v0.2.0`) di-*push* ke repositori.
