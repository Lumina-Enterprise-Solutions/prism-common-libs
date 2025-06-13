# 💎 Prism Common Libraries

[![Go CI Pipeline](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml/badge.svg)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Lumina-Enterprise-Solutions/prism-common-libs?style=flat-square&logo=github)](https://github.com/Lumina-Enterprise-Solutions/prism-common-libs/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lumina-Enterprise-Solutions/prism-common-libs)](https://goreportcard.com/report/github.com/Lumina-Enterprise-Solutions/prism-common-libs)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://opensource.org/licenses/MIT)

Repositori ini adalah fondasi bersama untuk semua microservices dalam ekosistem **Prism ERP**. Tujuannya adalah untuk menyediakan pustaka Go yang teruji, konsisten, dan dapat digunakan kembali untuk mempercepat pengembangan, mengurangi duplikasi kode, dan menegakkan praktik terbaik di seluruh platform.

---

## ✨ Prinsip Desain

Pustaka ini dirancang dengan prinsip-prinsip berikut:

-   **🎯 Generik & Dapat Digunakan Kembali**: Modul tidak memiliki pengetahuan spesifik tentang layanan yang menggunakannya. Semua konfigurasi spesifik disuntikkan oleh layanan pemanggil.
-   **🧩 Kohesi Tinggi, Kopling Rendah**: Fungsionalitas yang terkait dikelompokkan bersama (`auth`, `client`), sementara dependensi antar modul diminimalkan.
-   **🛡️ Aman & Teruji**: Semua kode harus melalui pipeline CI yang ketat, termasuk linting dan pengujian, untuk memastikan kualitas dan stabilitas.
-   **🔭 Dapat Diamati (Observable)**: Menyediakan utilitas standar untuk telemetri (tracing & metrics) agar semua layanan dapat dipantau secara konsisten.

---

## 📚 Struktur & Modul yang Tersedia

Setiap direktori tingkat atas adalah sebuah paket Go yang dapat diimpor, dikelompokkan berdasarkan fungsionalitasnya.

<br/>

| Direktori                                                                      | Deskripsi                                                                                                                                                                                            |
| ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 📂 **`auth/`** <br/> `import ".../auth"`                                         | 🔑 Utilitas untuk Otentikasi & Otorisasi. Berisi middleware untuk validasi token JWT dan Role-Based Access Control (RBAC) yang siap digunakan dengan Gin.                                              |
| 📂 **`client/`** <br/> `import ".../client"`                                     | 🔌 Klien untuk berinteraksi dengan layanan infrastruktur eksternal. Menyediakan wrapper yang disederhanakan untuk **HashiCorp Consul** (pendaftaran layanan) dan **HashiCorp Vault** (manajemen rahasia). |
| 📂 **`config/`** <br/> `import ".../config"`                                     | ⚙️ Pemuat konfigurasi terpusat. Membaca nilai dari *environment variables* dengan *fallback* ke **Consul KV store**, memungkinkan konfigurasi yang fleksibel dan dinamis.                           |
| 📂 **`ginutil/`** <br/> `import ".../ginutil"`                                   | 🌐 Utilitas khusus untuk framework Gin. Saat ini berisi *handler* `Health Check` yang dapat diperluas dan konsisten untuk digunakan oleh Consul atau load balancer.                                   |
| 📂 **`model/`** <br/> `import ".../model"`                                       | 📦 Definisi model data (struct) yang menjadi *Single Source of Truth* untuk entitas yang dibagikan antar layanan, seperti `User` dan `FileMetadata`.                                                  |
| 📂 **`telemetry/`** <br/> `import ".../telemetry"`                               | 📡 Utilitas untuk standarisasi observabilitas. Menyediakan *initializer* untuk **OpenTelemetry** (Tracing) yang mengirimkan data ke Jaeger, memastikan semua layanan memiliki jejak terdistribusi. |

---

## 🚀 Cara Menggunakan

Repositori ini dimaksudkan untuk digunakan sebagai modul Go dalam `go.work` atau `go.mod` di setiap microservice.

1.  **Tambahkan sebagai Dependensi**:
    Jalankan perintah berikut di dalam direktori layanan Anda. Pastikan untuk mengganti `vX.Y.Z` dengan versi rilis terbaru.
    ```bash
    go get github.com/Lumina-Enterprise-Solutions/prism-common-libs@vX.Y.Z
    ```

2.  **Impor dan Gunakan**:
    Impor paket yang dibutuhkan dalam kode Go Anda. Contoh:
    ```go
    import (
        "github.com/Lumina-Enterprise-Solutions/prism-common-libs/auth"
        "github.com/Lumina-Enterprise-Solutions/prism-common-libs/client"
    )
    ```

---

## 🤝 Kontribusi

Kontribusi Anda sangat kami hargai! Karena pustaka ini adalah dependensi kritis, kami mengikuti alur kerja yang ketat untuk menjaga kualitas.

1.  **Diskusi Awal**: Buka **GitHub Issue** untuk mendiskusikan perubahan atau fitur baru yang ingin Anda tambahkan. Diskusikan ide Anda dengan tim terlebih dahulu.
2.  **Fork & Branch**: Buat *fork* dari repositori ini dan buat *branch* baru dari `main` dengan nama yang deskriptif (misalnya, `feature/add-kafka-client` atau `bugfix/fix-jwt-parsing`).
3.  **Implementasi & Pengujian**:
    -   Tulis kode Anda sesuai dengan prinsip desain yang ada.
    -   Pastikan kode Anda lolos semua pemeriksaan linter (`golangci-lint run`).
    -   Tambahkan *unit test* yang relevan untuk perubahan Anda.
    -   Jalankan semua tes untuk memastikan tidak ada *regression* (`go test -v ./...`).
4.  **Pull Request (PR)**:
    -   Buka *Pull Request* ke branch `main`.
    -   Pastikan PR Anda memiliki deskripsi yang jelas tentang "mengapa" dan "apa" dari perubahan tersebut.
    -   PR harus mendapatkan persetujuan dari **CODEOWNERS** yang relevan.
    -   Semua *status checks* (CI Pipeline) harus berhasil.
5.  **Merge**: Setelah disetujui, maintainer akan me-*squash and merge* PR Anda.

---

## 🏷️ Versioning & Rilis

Kami mengikuti **Semantic Versioning 2.0.0**. Proses rilis diotomatisasi melalui GitHub Actions.

-   **Rilis Baru**: Untuk membuat rilis baru, seorang maintainer akan membuat dan me-*push* tag Git baru dengan format `vX.Y.Z` (misalnya, `v0.2.0`).
-   **Otomatisasi**: Tindakan ini akan secara otomatis memicu alur kerja `release.yml`, yang akan membuat **GitHub Release** baru berdasarkan tag tersebut.
