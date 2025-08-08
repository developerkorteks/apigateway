# Scrape Package - Winbu.TV Web Scraper

Package ini berisi kumpulan test functions untuk melakukan web scraping pada situs Winbu.TV menggunakan library Colly. Setiap test function menghasilkan output dalam format JSON yang terstruktur.

## Daftar Test Functions

### 1. TestHome (`homepage_test.go`)
**Fungsi**: Scraping halaman utama Winbu.TV  
**URL Target**: `https://winbu.tv/`  
**Command**: `go test -v ./scrape -run TestHome`

#### Struktur JSON Output:
```json
{
  "pageInfo": {
    "title": "string",
    "description": "string", 
    "canonicalUrl": "string",
    "ogTitle": "string",
    "ogDescription": "string",
    "ogUrl": "string",
    "ogImage": "string",
    "twitterCard": "string",
    "twitterSite": "string"
  },
  "notice": "string",
  "navigationMenu": [
    {
      "text": "string",
      "url": "string"
    }
  ],
  "top10Series": [
    {
      "rank": "string",
      "title": "string", 
      "url": "string",
      "imageUrl": "string",
      "rating": "string"
    }
  ],
  "latestDonghuaAnime": {
    "sectionTitle": "string",
    "moreLink": "string",
    "items": [
      {
        "title": "string",
        "url": "string", 
        "imageUrl": "string",
        "episode": "string",
        "time": "string",
        "views": "string",
        "quality": "string"
      }
    ]
  },
  "genres": [
    {
      "name": "string",
      "url": "string",
      "count": "string"
    }
  ],
  "top10Films": [...],
  "latestFilms": {...},
  "otherSeries": {...},
  "tvShows": {...}
}
```

### 2. TestScrapeAnimeDetail (`detail_page_test.go`)
**Fungsi**: Scraping detail halaman anime/series  
**URL Target**: `https://winbu.tv/series/legend-of-the-female-general/`  
**Command**: `go test -v ./scrape -run TestScrapeAnimeDetail`

#### Struktur JSON Output:
```json
{
  "url": "string",
  "title": "string",
  "posterImageUrl": "string", 
  "trailerUrl": "string",
  "rating": "string",
  "releaseDate": "string",
  "genres": ["string"],
  "synopsis": "string",
  "episodes": [
    {
      "title": "string",
      "url": "string"
    }
  ],
  "recommendations": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string", 
      "rating": "string"
    }
  ]
}
```

### 3. TestScrapeEpisodePageComplete (`detail_episode_test.go`)
**Fungsi**: Scraping halaman episode lengkap dengan stream dan download links  
**URL Target**: `https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/`  
**Command**: `go test -v ./scrape -run TestScrapeEpisodePageComplete`

#### Struktur JSON Output:
```json
{
  "url": "string",
  "episodeTitle": "string",
  "seriesTitle": "string",
  "episodeNav": {
    "previousEpisodeUrl": "string",
    "allEpisodesUrl": "string", 
    "nextEpisodeUrl": "string"
  },
  "seriesInfo": {
    "posterImageUrl": "string",
    "rating": "string",
    "genres": ["string"],
    "synopsis": "string"
  },
  "streamGroups": [
    {
      "quality": "string",
      "servers": [
        {
          "name": "string",
          "streamUrl": "string"
        }
      ]
    }
  ],
  "downloadGroups": [
    {
      "quality": "string", 
      "downloadLinks": [
        {
          "provider": "string",
          "url": "string"
        }
      ]
    }
  ],
  "recommendations": [...]
}
```

### 4. TestScrapeAvailableFilters (`Daftar_Anime_test.go`)
**Fungsi**: Scraping opsi filter yang tersedia  
**URL Target**: `https://winbu.tv/daftar-anime-2/`  
**Command**: `go test -v ./scrape -run TestScrapeAvailableFilters`

#### Struktur JSON Output:
```json
{
  "statusOptions": [
    {
      "displayName": "string",
      "queryValue": "string"
    }
  ],
  "typeOptions": [...],
  "orderOptions": [...], 
  "genreOptions": [...]
}
```

### 5. TestScrapeDaftarAnimeFiltered (`Daftar_Anime_test.go`)
**Fungsi**: Scraping daftar anime dengan filter  
**URL Target**: `https://winbu.tv/daftar-anime-2/` (dengan parameter filter)  
**Command**: `go test -v ./scrape -run TestScrapeDaftarAnimeFiltered`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string",
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string",
      "time": "string",
      "views": "string"
    }
  ]
}
```

### 6. TestScrapeFilmLimited (`film_test.go`)
**Fungsi**: Scraping daftar film (terbatas 3 halaman)  
**URL Target**: `https://winbu.tv/film/`  
**Command**: `go test -v ./scrape -run TestScrapeFilmLimited`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string", 
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string",
      "rating": "string",
      "quality": "string",
      "time": "string",
      "views": "string"
    }
  ]
}
```

### 7. TestScrapeAnimeTerbaruAnimasuLimited (`Anime_terbaru_animasu_test.go`)
**Fungsi**: Scraping anime terbaru Animasu (terbatas 3 halaman)  
**URL Target**: `https://winbu.tv/anime-terbaru-animasu/`  
**Command**: `go test -v ./scrape -run TestScrapeAnimeTerbaruAnimasuLimited`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string",
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string", 
      "imageUrl": "string",
      "episode": "string",
      "time": "string",
      "views": "string"
    }
  ]
}
```

### 8. TestScrapeAnimeDonghuaLimited (`animedonghua_test.go`)
**Fungsi**: Scraping anime donghua (terbatas 3 halaman)  
**URL Target**: `https://winbu.tv/animedonghua/`  
**Command**: `go test -v ./scrape -run TestScrapeAnimeDonghuaLimited`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string",
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string", 
      "episode": "string",
      "time": "string",
      "views": "string"
    }
  ]
}
```

### 9. TestScrapeOthersLimited (`jepangkoreachinabarat_test.go`)
**Fungsi**: Scraping series lainnya (Jepang, Korea, China, Barat)  
**URL Target**: `https://winbu.tv/others/`  
**Command**: `go test -v ./scrape -run TestScrapeOthersLimited`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string",
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string",
      "rating": "string",
      "episode": "string", 
      "time": "string",
      "views": "string"
    }
  ]
}
```

### 10. TestScrapeTVShowLimited (`tv_show_test.go`)
**Fungsi**: Scraping TV Show (terbatas 3 halaman)  
**URL Target**: `https://winbu.tv/tvshow/`  
**Command**: `go test -v ./scrape -run TestScrapeTVShowLimited`

#### Struktur JSON Output:
```json
{
  "sourceUrl": "string",
  "pagesScraped": 0,
  "paginationInfo": {
    "currentPage": "string",
    "lastVisiblePage": "string",
    "nextPageUrl": "string"
  },
  "totalItems": 0,
  "items": [
    {
      "title": "string",
      "url": "string",
      "imageUrl": "string",
      "rating": "string",
      "episode": "string",
      "time": "string", 
      "views": "string"
    }
  ]
}
```

## Cara Menjalankan

### Menjalankan Semua Test
```bash
cd /home/korteks/Documents/project/winbutv
go test -v ./scrape
```

### Menjalankan Test Spesifik
```bash
cd /home/korteks/Documents/project/winbutv
go test -v ./scrape -run TestHome
go test -v ./scrape -run TestScrapeAnimeDetail
go test -v ./scrape -run TestScrapeEpisodePageComplete
# dst...
```

## Dependencies

- `github.com/gocolly/colly/v2` - Web scraping framework
- `github.com/PuerkitoBio/goquery` - HTML parsing (untuk beberapa test)
- `encoding/json` - JSON marshaling
- `regexp` - Regular expressions untuk text cleaning
- `strings` - String manipulation
- `net/http` - HTTP client (untuk stream URL fetching)
- `sync` - Goroutine synchronization

## Fitur Utama

1. **Comprehensive Scraping**: Mencakup semua jenis konten di Winbu.TV
2. **Structured JSON Output**: Semua output dalam format JSON yang konsisten
3. **Pagination Support**: Mendukung scraping multi-halaman dengan batasan
4. **Stream URL Extraction**: Mengambil URL stream aktual dari server
5. **Download Links**: Mengekstrak link download dari berbagai provider
6. **Error Handling**: Penanganan error yang baik dengan logging
7. **Rate Limiting**: Menggunakan batasan halaman untuk menghindari overload
8. **Concurrent Processing**: Menggunakan goroutine untuk fetch stream URLs

## Catatan Penting

- Semua test yang melakukan pagination dibatasi maksimal 3 halaman untuk menghindari overload server
- Test `TestScrapeEpisodePageComplete` menggunakan goroutine untuk mengambil stream URLs secara concurrent
- Beberapa field mungkin kosong tergantung pada konten yang tersedia di halaman
- User-Agent disetel untuk menghindari blocking dari server
- Semua URL yang di-scrape adalah URL aktual dari situs Winbu.TV

## Output Format

Semua output menggunakan format JSON dengan indentasi 2 spasi untuk keterbacaan. Field yang tidak tersedia akan berupa string kosong atau array kosong, bukan `null`.