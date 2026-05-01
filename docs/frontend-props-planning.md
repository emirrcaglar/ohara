# Frontend Props Planning

## Overview

Connect Vue 3 frontend to Go backend via REST API. The backend serves manga/audio content; frontend needs to consume these APIs and render dynamic data instead of hardcoded placeholders.

---

## Backend API Summary

### Manga Endpoints

| Method | Endpoint | Response |
|--------|----------|----------|
| `GET` | `/manga/library` | HTML grid (server-rendered) |
| `GET` | `/manga/{id}/info` | JSON: `{id, path, title, page_count, pages: [{index, name, width, height}]}` |
| `GET` | `/manga/{id}/page/{page}` | JPEG binary |
| `GET` | `/manga/{id}/resume` | Redirects to `/reader?manga={id}&page={progress}&total={page_count}` |
| `POST` | `/manga/{id}/progress/{page}` | 204 No Content |

### Audio Endpoints

| Method | Endpoint | Response |
|--------|----------|----------|
| `GET` | `/audio/library` | HTML (server-rendered) |
| `GET` | `/audio/{id}/stream` | Audio file binary |

### Data Models

**MangaRow** (from DB):
```go
type MangaRow struct {
    ID        int64
    Path      string
    Title     string
    PageCount int
    Progress  int  // user's reading progress
}
```

**AudioRow** (from DB):
```go
type AudioRow struct {
    ID       int64
    Path     string
    Title    string
    Artist   string
    Album    string
    Duration int  // seconds
}
```

---

## Frontend Components - Required Props

### 1. VaultCard.vue

**Current props:**
```ts
title: string
category: string
image?: string
badge: string
stats: string
isAudio?: boolean
```

**Backend mapping:**
- `title` → `MangaRow.Title` or `AudioRow.Title`
- `category` → hardcoded "MANGA" or "AUDIO"
- `image` → `/manga/{id}/page/0` (manga cover), or generated waveform (audio)
- `badge` → file extension (`.cbz`, `.mp3`, etc.)
- `stats` → format from `PageCount` (manga) or `Duration` (audio)
- `isAudio` → boolean based on type

**New props:**
```ts
interface MangaProps {
  id: number
  title: string
  pageCount: number
  currentPage: number
  coverUrl: string
  fileExtension: string
}

interface AudioProps {
  id: number
  title: string
  artist: string
  album: string
  duration: number
  fileExtension: string
}
```

### 2. MediaCard.vue

**Current props:**
```ts
title: string
subtitle: string
image: string
badge?: string
badgeType?: 'primary' | 'secondary'
```

**Backend mapping:**
- `title` → `AudioRow.Title`
- `subtitle` → `AudioRow.Artist` or `AudioRow.Album`
- `image` → placeholder or album art
- `badge` → optional (e.g., "4K_HDR" for featured)
- `badgeType` → 'primary' | 'secondary'

**Proposed new interface:**
```ts
interface MediaCardProps {
  id: number
  title: string
  subtitle: string
  imageUrl: string
  badge?: string
  badgeType?: 'primary' | 'secondary'
  type: 'manga' | 'audio'
}
```

### 3. MediaDisplay.vue

**Current props:**
```ts
title: string
subLabel: string
image: string
bitrate: string
```

**Backend mapping:**
- `title` → current playing `AudioRow.Title`
- `subLabel` → current playing `AudioRow.Album` or "Live Link"
- `image` → album art or manga cover
- `bitrate` → calculated from audio file metadata

**Proposed new interface:**
```ts
interface MediaDisplayProps {
  title: string
  subLabel: string
  imageUrl: string
  bitrate: string
}
```

### 4. MediaControls.vue

**Current props:**
```ts
currentTime: string
totalTime: string
progress: number   // 0-100
volume: number     // 0-100
```

**New interface:**
```ts
interface MediaControlsProps {
  currentTime: number  // seconds
  totalTime: number    // seconds
  progress: number      // 0-100
  volume: number        // 0-100
  isPlaying: boolean
  onPlay: () => void
  onPause: () => void
  onSeek: (time: number) => void
  onVolumeChange: (volume: number) => void
  onNext: () => void
  onPrevious: () => void
}
```

### 5. QueueItem.vue

**Current props:**
```ts
title: string
subtitle: string
image: string
isActive?: boolean
```

**Proposed new interface:**
```ts
interface QueueItemProps {
  id: number
  title: string
  subtitle: string
  imageUrl: string
  isActive?: boolean
  duration?: number
}
```

### 6. VaultHeader.vue

**Current:** Static content

**New interface for dynamic stats:**
```ts
interface VaultHeaderProps {
  totalManga: number
  totalAudio: number
  activeFilter: 'all' | 'manga' | 'audio'
  onFilterChange: (filter: 'all' | 'manga' | 'audio') => void
}
```

### 7. ImportCard.vue

**Current:** Static, placeholder for upload flow

**No props changes needed - upload is out of scope for initial backend connection.**

---

## Required Pinia Stores

### mangaStore

```ts
interface MangaState {
  items: MangaRow[]
  loading: boolean
  error: string | null
}

interface MangaRow {
  id: number
  title: string
  path: string
  pageCount: number
  currentPage: number
}

actions:
  - fetchLibrary(): Promise<void>
  - fetchMangaInfo(id: number): Promise<MangaInfo>
  - saveProgress(id: number, page: number): Promise<void>
```

### audioStore

```ts
interface AudioState {
  items: AudioRow[]
  currentTrack: AudioRow | null
  queue: AudioRow[]
  loading: boolean
  error: string | null
}

interface AudioRow {
  id: number
  title: string
  artist: string
  album: string
  duration: number
  path: string
}

actions:
  - fetchLibrary(): Promise<void>
  - playTrack(id: number): void
  - streamUrl(id: number): string  // returns /audio/{id}/stream
```

### playerStore

```ts
interface PlayerState {
  currentTrack: AudioRow | null
  isPlaying: boolean
  currentTime: number
  duration: number
  volume: number
  queue: AudioRow[]
}

actions:
  - play()
  - pause()
  - seek(time: number)
  - setVolume(volume: number)
  - next()
  - previous()
  - addToQueue(track: AudioRow)
```

---

## View Updates Required

### LibraryView.vue

**Current:** Renders hardcoded VaultCards + ImportCard

**Changes:**
1. Fetch manga list from `GET /manga/library` (or better: create new `GET /api/manga` returning JSON)
2. Map `MangaRow[]` to `VaultCard` props
3. Add filter tabs for All/Manga/Audio
4. Show loading/error states

**Proposed interface:**
```vue
<VaultCard
  v-for="manga in mangaStore.items"
  :key="manga.id"
  :id="manga.id"
  :title="manga.title"
  :pageCount="manga.pageCount"
  :currentPage="manga.currentPage"
  :coverUrl="`/manga/${manga.id}/page/0`"
  badge=".cbz"
  :stats="`SIZE: ${estimatedSize} MB`"
/>
```

### MediaView.vue

**Current:** Hardcoded audio player with placeholder data

**Changes:**
1. Use `audioStore.currentTrack` for MediaDisplay
2. Use `playerStore` for MediaControls
3. Use `audioStore.queue` for QueueItem list
4. Stream audio via `<audio src="/audio/{id}/stream">`

### HomeView.vue

**Current:** Static MediaCards with hardcoded data

**Changes:**
1. Fetch featured/recent items from stores
2. Show real data from backend

---

## API Enhancements (Backend Changes)

To support the frontend properly, consider adding these JSON endpoints:

### New: `GET /api/manga` → JSON list
```json
{
  "items": [
    {
      "id": 1,
      "title": "Neon Genesis Vol 01",
      "pageCount": 256,
      "currentPage": 42,
      "fileExtension": ".cbz"
    }
  ],
  "total": 1248
}
```

### New: `GET /api/manga/{id}` → JSON
```json
{
  "id": 1,
  "title": "Neon Genesis Vol 01",
  "path": "/mnt/media/manga/...",
  "pageCount": 256,
  "currentPage": 42,
  "pages": [
    {"index": 0, "name": "001.jpg", "width": 1200, "height": 1800}
  ]
}
```

### New: `GET /api/audio` → JSON list
```json
{
  "items": [
    {
      "id": 1,
      "title": "Track Name",
      "artist": "Artist Name",
      "album": "Album Name",
      "duration": 245,
      "fileExtension": ".mp3"
    }
  ],
  "total": 512
}
```

### New: `POST /api/audio/{id}/progress` (if needed)
For audio playback progress tracking.

---

## Implementation Order

1. **Create TypeScript types** (`src/types/api.ts`)
   - Define all API response interfaces

2. **Create API client** (`src/api/client.ts`)
   - Fetch wrapper with error handling
   - Typed response parsing

3. **Create Pinia stores**
   - `mangaStore` with fetch/actions
   - `audioStore` with fetch/actions
   - `playerStore` for playback state

4. **Update VaultCard** - accept dynamic props

5. **Update LibraryView** - use mangaStore

6. **Update MediaView** - use audioStore + playerStore

7. **Update MediaControls** - add event handlers

8. **Update Sidebar** - update stats from stores

9. **Add loading/error states** to all views

---

## File Structure (New Files)

```
frontend/src/
├── api/
│   ├── client.ts      # Base fetch wrapper
│   ├── manga.ts       # Manga API calls
│   └── audio.ts       # Audio API calls
├── types/
│   └── api.ts         # TypeScript interfaces
└── stores/
    ├── manga.ts       # Manga Pinia store
    ├── audio.ts       # Audio Pinia store
    └── player.ts      # Player Pinia store
```

---

## Notes

- Backend serves images as JPEG (`/manga/{id}/page/{page}`)
- Audio streaming via `/audio/{id}/stream`
- Progress tracking via `POST /manga/{id}/progress/{page}`
- Currently no auth/user system (hardcoded user_id=1 for progress)
- Backend uses embedded HTML for library pages - frontend can either:
  a) Render those HTML strings in iframes
  b) Backend team adds JSON endpoints (recommended)
