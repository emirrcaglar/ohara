# Backend Structure

## Overview

Go backend serving manga and audio content via REST API. Uses SQLite for metadata, serves compressed images from CBZ archives, and streams audio files.

## Directory Layout

```
backend/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── cache/
│   │   └── disk_cache.go    # LRU cache utilities
│   ├── db/
│   │   ├── db.go            # SQLite init + migrations
│   │   ├── manga.go         # Manga queries
│   │   └── audio.go         # Audio queries
│   ├── handler/
│   │   ├── manga.go         # Manga HTTP handlers
│   │   ├── audio.go         # Audio HTTP handlers
│   │   └── cache.go         # PageCache + Inflight deduplication
│   ├── media/
│   │   ├── audio.go         # Audio metadata parsing
│   │   └── cbz.go           # CBZ (comic book zip) parser
│   ├── router/
│   │   └── router.go        # Route definitions
│   ├── scanner/
│   │   └── scanner.go        # Media file indexer
│   ├── server/
│   │   └── server.go        # HTTP/HTTPS server
│   ├── ui/
│   │   ├── file.go          # embed.FS directive
│   │   ├── home.html        # Home page template
│   │   ├── index.html       # Reader page template
│   │   └── static/
│   │       └── style.css    # Shared styles
│   ├── utils/
│   │   └── imgutil/
│   │       └── imgutil.go   # Image compression
│   └── worker/
│       └── cache_worker.go  # Background cache cleanup
└── app-data/                # Runtime data (DB, cache, certs)
```

## Request Flow

```
HTTP Request
    ↓
http.ServeMux (router/router.go)
    ↓
Handler (manga.go / audio.go)
    ↓
DB (db.go) ←→ SQLite
    ↓
Media File (CBZ / audio file)
```

## Key Components

### main.go

Entry point handling:
- CLI flags: `--domain`, `--port`, `--data`, `--scan`
- Database initialization
- Optional media scanning
- Cache cleaner worker startup
- Server startup (HTTP or HTTPS)

### router.go

`SetupRoutes()` creates `http.ServeMux` and registers all handlers:

```go
mux.Handle("GET /static/", http.FileServer(http.FS(ui.Files)))

mux.HandleFunc("GET /", homePageHandler)
mux.HandleFunc("GET /reader", readerHandler)

mux.HandleFunc("GET /manga/library", mangaHandler.HandleMangaList)
mux.HandleFunc("GET /manga/{id}/resume", mangaHandler.HandleMangaResume)
mux.HandleFunc("GET /manga/{id}/page/{page}", mangaHandler.HandleMangaPage)
mux.HandleFunc("POST /manga/{id}/progress/{page}", mangaHandler.HandleMangaProgress)
mux.HandleFunc("GET /manga/{id}/info", mangaHandler.HandleMangaInfo)

mux.HandleFunc("GET /audio/library", audioHandler.HandleAudioList)
mux.HandleFunc("GET /audio/{id}/stream", audioHandler.HandleAudioStream)
```

### Manga Handler

Handles manga page delivery with caching and prefetching:

**Endpoints:**
| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| GET | `/manga/library` | HandleMangaList | Renders HTML grid |
| GET | `/manga/{id}/resume` | HandleMangaResume | Redirects to reader at last position |
| GET | `/manga/{id}/page/{page}` | HandleMangaPage | Returns compressed JPEG |
| POST | `/manga/{id}/progress/{page}` | HandleMangaProgress | Saves reading position |
| GET | `/manga/{id}/info` | HandleMangaInfo | Returns JSON metadata |

**Page Delivery Flow:**
1. Check `PageCache` for cached compressed JPEG
2. If miss, check `Inflight` map (prevent duplicate work)
3. Open CBZ file with `cbz.Open(path)`
4. Extract page image
5. Compress with `imgutil.Compress()`
6. Store in cache
7. Return JPEG response

**Prefetching:** After returning a page, spawns goroutine to compress next 15 pages in background.

### Audio Handler

Handles audio streaming and metadata:

**Endpoints:**
| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| GET | `/audio/library` | HandleAudioList | Renders HTML grid with player |
| GET | `/audio/{id}/stream` | HandleAudioStream | Streams audio file |

### Database Schema

```sql
CREATE TABLE user (
    id INTEGER PRIMARY KEY,
    username TEXT,
    role TEXT,
    created_at TIMESTAMP
);

CREATE TABLE manga (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE,
    title TEXT,
    page_count INTEGER,
    indexed_at TIMESTAMP
);

CREATE TABLE manga_progress (
    user_id INTEGER,
    manga_id INTEGER,
    page INTEGER,
    updated_at TIMESTAMP,
    PRIMARY KEY (user_id, manga_id)
);

CREATE TABLE audio (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE,
    title TEXT,
    artist TEXT,
    album TEXT,
    duration INTEGER,
    indexed_at TIMESTAMP
);
```

### Caching

**PageCache** (`cache.go`):
- Disk-based cache at `{dataDir}/cache/{mangaID}_{page}.jpg`
- LRU eviction when cache exceeds size limit
- Default max size: 1GB

**Inflight** (`cache.go`):
- `sync.Map` tracking in-flight requests
- Prevents thundering herd problem (multiple goroutines compressing same page)

### Media Parsers

**CBZ** (`media/cbz.go`):
- Opens ZIP archive
- Extracts image files (jpg/png/webp)
- Sorts by filename
- Returns page readers

**Audio** (`media/audio.go`):
- Uses `github.com/dhowden/tag` for ID3/metadata
- Uses ffprobe for duration

### UI Embedding

`ui/file.go`:
```go
//go:embed *
var Files embed.FS
```

All files in `ui/` directory embedded at compile time.

## Dependencies

| Package | Purpose |
|---------|---------|
| `modernc.org/sqlite` | Pure Go SQLite driver |
| `github.com/dhowden/tag` | ID3/metadata tag reading |
| `golang.org/x/crypto` | ACME/autocert for HTTPS |
| `golang.org/x/image` | Image resizing |

## Server Modes

### HTTP Mode
Plain HTTP server on specified port.

### HTTPS Mode (Auto-cert)
When `--domain` flag provided:
1. Uses `golang.org/x/crypto/acme/autocert`
2. Requests certificate from Let's Encrypt
3. Caches certificates to `{dataDir}/certs/`
4. Redirects HTTP → HTTPS
