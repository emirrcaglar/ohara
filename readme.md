# Ohara

Media server for manga and audio content.

## Architecture

- **Backend**: Go API server (port 3000) - serves JSON API and media files
- **Frontend**: Vue SPA (port 5173 dev, or serve via nginx/cdn) - UI client

## Quick Start

### Backend

```bash
cd backend

# Build
go build ./...

# Start server (default port 3000)
go run ./cmd --port 3000 --data ./app-data
```

### Frontend (Development)

```bash
cd frontend

# Install dependencies
npm install

# Start dev server (proxies /api, /manga, /audio to localhost:3000)
npm run dev
```

Then open http://localhost:5173

### Production Build

```bash
# 1. Build frontend into the backend embed directory
cd frontend
npm install
npm run build:embed

# 2. Build the single deployable backend binary
cd ../backend
go build -o ohara ./cmd
```

The resulting binary serves both the JSON API and the embedded Vue frontend.

### Linux VPS Build

Backend only (API server):
```powershell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-w -s" -o ohara ./cmd
```

## Adding Media

### Using the Scanner

```bash
# Scan a directory for all media types
go run ./cmd --scan all /path/to/media

# Scan only manga (CBZ files)
go run ./cmd --scan manga /path/to/media

# Scan only audio (mp3, flac, ogg, m4a, wav, aac)
go run ./cmd --scan audio /path/to/music
```

### Media Requirements

**Manga:**
- Must be `.cbz` (ZIP archive containing images)
- Images inside should be jpg/png/webp
- Title is extracted from the CBZ filename

**Audio:**
- Supported formats: `*.mp3`, `*.flac`, `*.ogg`, `*.m4a`, `*.wav`, `*.aac`
- Metadata (title, artist, album) is read from file tags
- Duration is extracted via ffprobe

### Database Location

SQLite DB is stored at: `{dataDir}/ohara.db` (default: `./app-data/ohara.db`)

## API Endpoints

### Manga

| Method | Path | Returns |
|--------|------|---------|
| GET | `/api/manga` | JSON list of manga |
| GET | `/api/manga/{id}` | JSON manga details |
| GET | `/manga/{id}/page/{page}` | JPEG page image |
| GET | `/manga/{id}/resume` | Redirect to reader |
| POST | `/manga/{id}/progress/{page}` | Save reading progress |

### Audio

| Method | Path | Returns |
|--------|------|---------|
| GET | `/api/audio` | JSON list of tracks |
| GET | `/audio/{id}/stream` | Audio file stream |

### Frontend Routes

| Path | View |
|------|------|
| `/` | Redirects to `/library` |
| `/library` | Manga library grid |
| `/media` | Audio player with queue |
| `/reader` | Manga reader (prev/next navigation) |
| `/uploads` | Upload management |

## Project Structure

```
ohara/
├── backend/
│   ├── cmd/main.go           # Entry point
│   └── internal/
│       ├── db/               # SQLite database
│       ├── handler/          # HTTP handlers
│       ├── media/            # Media parsers (CBZ, audio)
│       ├── router/           # Route definitions
│       ├── scanner/          # Media file indexer
│       ├── server/           # HTTP/HTTPS server
│       └── worker/           # Background tasks
├── frontend/
│   ├── src/
│   │   ├── api/              # API client functions
│   │   ├── components/       # Vue components
│   │   ├── stores/           # Pinia stores
│   │   ├── types/            # TypeScript interfaces
│   │   └── views/            # Page components
│   └── vite.config.ts        # Vite config with API proxy
└── docs/                     # Documentation
```
