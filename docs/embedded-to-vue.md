# Transitioning from Embedded HTML to Vue

## The Problem

Currently the backend generates HTML for the library pages server-side:

```go
// backend/internal/handler/manga.go
func (h *MangaHandler) HandleMangaList(w http.ResponseWriter, r *http.Request) {
    // Builds HTML string manually
    html := fmt.Sprintf(`<!DOCTYPE html>
        <html>...
        <div class="grid">%s</div>
        ...`, cards.String())
    w.Write([]byte(html))
}
```

This approach has significant limitations:

| Issue | Impact |
|-------|--------|
| No component reuse | Copy-paste HTML everywhere |
| No state management | Can't react to user interactions |
| No reactivity | Page reload needed for updates |
| Hard to maintain | HTML strings in Go code |
| No type safety | JavaScript operates on guessed types |
| No hot reload | Backend restart needed for changes |

## Current Architecture

```
┌─────────────┐     ┌─────────────┐
│   Go HTML   │     │  Vue SPA    │
│  Templates  │     │  Frontend   │
└──────┬──────┘     └──────┬──────┘
       │                   │
       ▼                   ▼
   /manga/library      /library
   /audio/library     /media
```

Two separate UIs that don't share state or code.

## Target Architecture

```
┌─────────────────────────────────┐
│          Vue SPA                │
│  ┌──────────────────────────┐   │
│  │      Pinia Stores        │   │
│  │   mangaStore (state)     │   │
│  │   audioStore (state)     │   │
│  │   playerStore (state)     │   │
│  └──────────────────────────┘   │
│              │                   │
│              ▼                   │
│  ┌──────────────────────────┐   │
│  │   Vue Components          │   │
│  │   VaultCard, MediaCard,   │   │
│  │   MediaControls, etc.     │   │
│  └──────────────────────────┘   │
└──────────────┬──────────────────┘
               │
               ▼ (fetch JSON)
        ┌─────────────┐
        │  Go Backend │
        │   REST API  │
        └─────────────┘
```

## What Changes

### 1. New JSON API Endpoints

Instead of returning HTML, endpoints return JSON:

**Before (HTML):**
```
GET /manga/library → <html><body>...</body></html>
```

**After (JSON):**
```
GET /api/manga → { "items": [...], "total": 42 }
```

### 2. Frontend Fetches Data

Vue components use `fetch()` to get data from stores:

```ts
// stores/manga.ts
async function fetchLibrary() {
  const response = await fetch('/api/manga')
  items.value = await response.json()
}
```

### 3. Components Render Dynamically

**Before (static HTML string):**
```html
<a class="manga-card" href="/manga/1/resume">
  <img src="/manga/1/page/0">
  <span class="title">Neon Genesis</span>
</a>
```

**After (Vue template):**
```vue
<VaultCard
  v-for="manga in mangaStore.items"
  :key="manga.id"
  :manga="manga"
  :coverUrl="`/api/manga/${manga.id}/page/0`"
/>
```

## Implementation Steps

### Phase 1: API Endpoints (Backend)

Add new JSON endpoints alongside existing HTML endpoints:

```go
// router/router.go
mux.HandleFunc("GET /api/manga", mangaHandler.HandleMangaListJSON)
mux.HandleFunc("GET /api/manga/{id}", mangaHandler.HandleMangaInfoJSON)
mux.HandleFunc("GET /api/audio", audioHandler.HandleAudioListJSON)
```

Response format:
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

### Phase 2: Pinia Stores (Frontend)

Already implemented:
- `stores/manga.ts` - manga state and actions
- `stores/audio.ts` - audio state and actions
- `stores/player.ts` - playback state and controls

### Phase 3: Component Props (Frontend)

Update components to accept typed props instead of hardcoded values:

**VaultCard before:**
```ts
props: {
  title: String,
  category: String,
  badge: String,
  stats: String,
  image: String
}
```

**VaultCard after:**
```ts
props: {
  manga: Object as () => MangaRow,
  audio: Object as () => AudioRow,
  coverUrl: String,
  stats: String
}
```

### Phase 4: View Integration

Connect views to stores:

```ts
// LibraryView.vue
onMounted(() => {
  mangaStore.fetchLibrary()
})
```

## Breaking Changes

### URL Changes

| Old | New |
|-----|-----|
| `/manga/library` | `/library` |
| `/audio/library` | `/media` |
| `/manga/{id}/resume` | `/reader?manga={id}` |
| `/audio/{id}/stream` | `/api/audio/{id}/stream` |

### Frontend Depends on Backend

The Vue SPA now **requires** the backend API to function. No more:
- Static HTML fallback
- Independent operation
- Debugging with mock data

## Benefits

| Benefit | Description |
|---------|-------------|
| Type safety | TypeScript interfaces for API responses |
| Reactivity | UI updates automatically on state changes |
| Component reuse | DRY code with Vue components |
| Hot reload | Vue dev server updates instantly |
| State sharing | Same data accessible everywhere |
| Single responsibility | Backend is API, Frontend is UI |

## Migration Order

1. **Add `/api/*` JSON endpoints** to backend
2. **Keep existing HTML endpoints** temporarily
3. **Update Vue components** to use stores
4. **Wire stores** to API endpoints
5. **Update routes** in Vue Router
6. **Remove HTML endpoints** when Vue is verified working

## Debugging Tips

### Backend API
```bash
curl http://localhost:8080/api/manga
curl http://localhost:8080/api/audio
```

### Frontend State
```ts
// In Vue DevTools or console
import { useMangaStore } from '../stores/manga'
const store = useMangaStore()
store.items  // See current data
store.loading  // Check fetch state
```

### Network
Check browser DevTools → Network tab for API responses and timing.
