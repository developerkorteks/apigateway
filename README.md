Tentu, saya akan integrasikan lampiran skema JSON dan penyesuaian *timeout* ke dalam dokumen persyaratan produk yang sudah ada. Versi ini menjadi lebih lengkap dan siap untuk diimplementasikan.
Performa & Skalabilitas

Gunakan goroutines untuk paralel request API.

Cache hasil (Redis/in-memory) untuk mengurangi beban scraping.

Health check otomatis setiap X menit, update status di dashboard.

Siap melayani ribuan user secara bersamaan.

Tambah/hapus kategori.

Tambah/hapus API utama & fallback.

Lihat log request & error.

Lihat health check API (status OK/Timeout/Error).

Jalankan tes manual pada API tertentu.

Performa & Skalabilitas

Gunakan goroutines untuk paralel request API.

Cache hasil (valkey/in-memory) untuk mengurangi beban scraping.

Health check otomatis setiap X menit, update status di dashboard.

Siap melayani ribuan user secara bersamaan.
-----

### **Dokumen Persyaratan Produk (PRD): Sistem Agregator API Dinamis dengan Fallback**

**Versi 1.1**

#### **1. Visi & Tujuan Utama**

Membangun sebuah sistem API Gateway berbasis Golang yang tangguh, scalable, dan dinamis dengan nama proyek `apicategorywithfallback`. Sistem ini berfungsi sebagai *single source of truth* yang mengagregasi data dari berbagai API eksternal (terutama berbasis *web scraping*) berdasarkan kategori. Fitur utamanya adalah mekanisme *fallback* otomatis, validasi data yang ketat, caching cerdas, dan sebuah dashboard manajemen untuk mengelola sumber API tanpa perlu *re-deploy*.

**Tujuan Bisnis:**

  * Menyediakan data yang konsisten dan andal kepada *end-user* meskipun salah satu sumber API eksternal mengalami gangguan atau lambat.
  * Mempermudah penambahan kategori dan sumber API baru di masa depan secara cepat.
  * Mengurangi latensi dan beban pada API eksternal melalui *caching*.
  * Memberikan visibilitas penuh terhadap kesehatan dan performa setiap API melalui dashboard.

-----

#### **2. Arsitektur & Logika Inti**

##### **2.1. Manajemen Kategori & API Dinamis**

  * **Sumber Konfigurasi**: Sistem harus memuat konfigurasi kategori dan daftar API dari sebuah database (misal: PostgreSQL, MongoDB) untuk memungkinkan perubahan *real-time* melalui dashboard.

  * **Hierarki API**:

      * Sistem memiliki beberapa **Kategori** (contoh: `anime`, `drakor`).
      * Setiap **Kategori** memiliki satu atau lebih **API Utama (Primary API)**.
      * Setiap **API Utama** memiliki urutan prioritas **API Fallback (Fallback APIs)**.

    **Contoh Struktur Konfigurasi (JSON):**

    ```json
    {
      "categories": [
        {
          "name": "anime",
          "is_active": true,
          "endpoints": [
            {
              "path": "/api/v1/home",
              "primary_apis": [
                { "source_name": "samehadaku_v1", "base_url": "http://api_source_1/v1", "priority": 1 },
                { "source_name": "otakudesu_v3", "base_url": "http://api_source_2/v1", "priority": 2 }
              ],
              "fallback_apis": {
                "samehadaku_v1": ["http://fallback1/v1", "http://fallback2/v1"],
                "otakudesu_v3": ["http://fallback3/v1"]
              }
            }
          ]
        }
      ]
    }
    ```

##### **2.2. Alur Request & Mekanisme Fallback**

1.  *Client* mengirim *request* ke *endpoint* sistem, misal: `GET /api/v1/home?category=anime`.
2.  Sistem mengidentifikasi kategori (`anime`) dan *endpoint* (`/home`).
3.  Sistem secara *concurrent* (menggunakan *goroutines*) mengirim *request* ke semua **API Utama** untuk kategori tersebut.
4.  Untuk setiap *response* yang diterima:
      * **Validasi Confidence Score**: Jika `confidence_score < 0.5`, *response* dianggap tidak valid. Segera proses *fallback* untuk API tersebut.
      * **Validasi Integritas Data (Schema Validation)**: *Response* harus divalidasi berdasarkan skema JSON yang telah didefinisikan (lihat **Lampiran A**). Periksa apakah *field* wajib (seperti `url`, `judul`, `cover`, `streaming_url`) tidak *null*, tidak kosong, dan bukan *placeholder* error. Jika validasi gagal, anggap *response* tidak valid dan proses *fallback*.
5.  *Response* valid pertama yang diterima (berdasarkan prioritas API Utama) akan langsung di-*cache* dan dikirim kembali ke *client*.
6.  Jika semua API Utama dan *fallback*-nya gagal, kembalikan *response* error yang sesuai (misal: `503 Service Unavailable`).

-----

#### **3. Spesifikasi Teknis**

##### **3.1. Validasi Skema & Data**

  * Buat modul `validator` khusus yang berisi definisi skema (struct Golang) untuk setiap *endpoint* seperti yang terlampir.
  * Validator ini harus memeriksa keberadaan *field*, tipe data, dan format (misal: URL valid).
  * **Field Wajib Universal**: `url`, `judul` atau `title`, `anime_slug` atau `slug`, `cover` atau `cover_url`. Untuk *endpoint* episode, tambahkan `streaming_url`.
  * Kegagalan validasi adalah pemicu *fallback*.

##### **3.2. Caching (Redis)**

  * Gunakan Redis untuk *caching*.
  * Hanya *response* yang lolos semua validasi (`confidence_score >= 0.5` dan validasi skema) yang boleh di-*cache*.
  * **Struktur Cache Key**: `category:endpoint:parameter_hash`. Contoh: `anime:/api/v1/search:query=naruto`.
  * Setel TTL (Time-To-Live) yang sesuai untuk setiap *endpoint* (misal: 15 menit untuk `/home`, 1 jam untuk `/anime-detail`).

##### **3.3. Performa & Skalabilitas**

  * **Concurrency**: Manfaatkan *goroutines* dan *channels* untuk menangani *request* ke API eksternal secara paralel.
  * **Timeout**: Mengingat sumber API berasal dari *scraping* yang bisa lambat, terapkan *timeout* yang lebih longgar namun tetap terkontrol. **Rekomendasi awal: 15-20 detik** untuk setiap *request* ke API eksternal. Nilai ini harus bisa dikonfigurasi.
  * **Rate Limiting**: Implementasikan *rate limiting* per IP untuk mencegah penyalahgunaan.
  * **Background Worker (Health Check)**: Sebuah *worker* yang berjalan di latar belakang harus secara periodik (misal: setiap 10 menit) melakukan *health check* ke semua API yang terdaftar. Hasilnya (OK, Timeout, Error) harus diperbarui di dashboard.

##### **3.4. Dashboard Manajemen (Admin)**

  * Buat antarmuka web sederhana menggunakan *template* HTML Golang atau *framework* frontend.
  * **Fitur Wajib:**
      * CRUD Kategori & API.
      * Log Viewer.
      * Health Check Status.
      * Statistik (jumlah *request*, tingkat *fallback*, tingkat *error*) per API.

-----

#### **4. Definisi Endpoint**

Sistem harus mengimplementasikan *endpoint-endpoint* berikut. Skema *response* yang diharapkan untuk setiap *endpoint* terlampir dalam **Lampiran A**.
Sebaiknya pisah secara logis di backend, tapi fleksibel di frontend.
Jadi, API endpoint bisa mengakomodasi query per kategori (?category=anime) atau semua kategori (?category=all).

Dengan begini, kamu bisa menambah kategori tanpa ngoding ulang karena sistem sudah dinamis, cukup tambah konfigurasi kategori di database / config file.
Mendukung query:

GET ?category=anime → Semua API utama kategori anime

GET ?category=all → Semua kategori
  * `GET /api/v1/home`
  * `GET /api/v1/jadwal-rilis`
  * `GET /api/v1/jadwal-rilis/{day}`
  * `GET /api/v1/anime-terbaru`
  * `GET /api/v1/movie`
  * `GET /api/v1/anime-detail`
  * `GET /api/v1/episode-detail`
  * `GET /api/v1/search`

-----

#### **5. Struktur Proyek (Rekomendasi)**

```
apicategorywithfallback/
├── cmd/main.go
├── pkg/
│   ├── config/
│   ├── database/
│   ├── cache/
│   ├── logger/
│   └── validator/
├── internal/
│   ├── api/
│   ├── domain/
│   ├── service/
│   ├── repository/
│   └── dashboard/
├── web/
├── README.md
└── go.mod
```

-----

### **Lampiran A: Skema Struktur JSON per Endpoint**

Berikut adalah detail struktur JSON yang harus divalidasi untuk setiap *response* dari API eksternal.

\<details\>
\<summary\>\<strong\>1. GET /api/v1/home/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "top10": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "rating": "string",
      "cover": "string",
      "genres": ["string"]
    }
  ],
  "new_eps": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "episode": "string",
      "rilis": "string",
      "cover": "string"
    }
  ],
  "movies": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "tanggal": "string",
      "cover": "string",
      "genres": ["string"]
    }
  ],
  "jadwal_rilis": {
    "DayName": [
      {
        "title": "string",
        "url": "string",
        "anime_slug": "string",
        "cover_url": "string",
        "type": "string",
        "score": "string",
        "genres": ["string"],
        "release_time": "string"
      }
    ]
  }
}
```

\</details\>

\<details\>
\<summary\>\<strong\>2. GET /api/v1/jadwal-rilis/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "data": {
    "DayName": [
      {
        "title": "string",
        "url": "string",
        "anime_slug": "string",
        "cover_url": "string",
        "type": "string",
        "score": "string",
        "genres": ["string"],
        "release_time": "string"
      }
    ]
  }
}
```

\</details\>

\<details\>
\<summary\>\<strong\>3. GET /api/v1/jadwal-rilis/{day}\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "data": [
    {
      "title": "string",
      "url": "string",
      "anime_slug": "string",
      "cover_url": "string",
      "type": "string",
      "score": "string",
      "genres": ["string"],
      "release_time": "string"
    }
  ]
}
```

\</details\>

\<details\>
\<summary\>\<strong\>4. GET /api/v1/anime-terbaru/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "data": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "episode": "string",
      "uploader": "string",
      "rilis": "string",
      "cover": "string"
    }
  ]
}
```

\</details\>

\<details\>
\<summary\>\<strong\>5. GET /api/v1/movie/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "data": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "status": "string",
      "skor": "string",
      "sinopsis": "string",
      "views": "string",
      "cover": "string",
      "genres": ["string"],
      "tanggal": "string"
    }
  ]
}
```

\</details\>

\<details\>
\<summary\>\<strong\>6. GET /api/v1/anime-detail/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "judul": "string",
  "url": "string",
  "anime_slug": "string",
  "cover": "string",
  "episode_list": [
    {
      "episode": "string",
      "title": "string",
      "url": "string",
      "episode_slug": "string",
      "release_date": "string"
    }
  ],
  "recommendations": [
    {
      "title": "string",
      "url": "string",
      "anime_slug": "string",
      "cover_url": "string",
      "rating": "string",
      "episode": "string"
    }
  ],
  "status": "string",
  "tipe": "string",
  "skor": "string",
  "penonton": "string",
  "sinopsis": "string",
  "genre": ["string"],
  "details": {
    "Japanese": "string",
    "Synonyms": "string",
    "English": "string",
    "Status": "string",
    "Type": "string",
    "Source": "string",
    "Duration": "string",
    "Total Episode": "string",
    "Studio": "string",
    "Producers": "string",
    "Released:": "string"
  },
  "rating": {
    "score": "string",
    "users": "string"
  }
}
```

\</details\>

\<details\>
\<summary\>\<strong\>7. GET /api/v1/episode-detail/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "title": "string",
  "thumbnail_url": "string",
  "streaming_servers": [
    {
      "server_name": "string",
      "streaming_url": "string"
    }
  ],
  "release_info": "string",
  "download_links": {
    "FormatName": {
      "QualityName": [
        {
          "provider": "string",
          "url": "string"
        }
      ]
    }
  },
  "navigation": {
    "previous_episode_url": "string | null",
    "all_episodes_url": "string",
    "next_episode_url": "string | null"
  },
  "anime_info": {
    "title": "string",
    "thumbnail_url": "string",
    "synopsis": "string",
    "genres": ["string"]
  },
  "other_episodes": [
    {
      "title": "string",
      "url": "string",
      "thumbnail_url": "string",
      "release_date": "string"
    }
  ]
}
```

\</details\>

\<details\>
\<summary\>\<strong\>8. GET /api/v1/search/\</strong\>\</summary\>

```json
{
  "confidence_score": "number",
  "message": "string",
  "source": "string",
  "data": [
    {
      "judul": "string",
      "url": "string",
      "anime_slug": "string",
      "status": "string",
      "tipe": "string",
      "skor": "string",
      "penonton": "string",
      "sinopsis": "string",
      "genre": ["string"],
      "cover": "string"
    }
  ]
}
```

\</details\>