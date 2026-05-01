# Pinia Stores

## What is a Store?

A store is a state management pattern for Vue applications. In Pinia (the official Vue 5 state management library), a store is essentially a reactive container that holds state and business logic.

Think of it as a global reactive object - accessible from anywhere in your app, without prop drilling or event buses.

## Why Do We Need Stores?

### The Problem

Without stores, state lives in individual components:

```
HomeView (has items[])
    ↓ props
LibraryView (has items[])
    ↓ props
VaultCard (displays items)
```

Problems:
- Same data fetched multiple times
- Hard to share state between unrelated components
- Components become bloated with data fetching logic

### The Store Solution

```
mangaStore (has items[])
    ↓
Any component can access it directly
```

Benefits:
- Single source of truth
- Reactive updates propagate everywhere
- Business logic encapsulated in one place

## Our Stores

### mangaStore

**Location:** `src/stores/manga.ts`

**State:**
```ts
items: MangaRow[]      // List of all manga
loading: boolean       // Currently fetching?
error: string | null   // Last error message
total: number          // Total count from API
```

**Actions:**
```ts
fetchLibrary()         // GET /api/manga
getMangaInfo(id)       // GET /api/manga/:id/info
updateProgress(id, page)  // POST /api/manga/:id/progress/:page
```

**Usage:**
```ts
import { useMangaStore } from '../stores/manga'

const mangaStore = useMangaStore()

// Access state directly (reactive)
console.log(mangaStore.items)

// Call actions
mangaStore.fetchLibrary()
```

### audioStore

**Location:** `src/stores/audio.ts`

**State:**
```ts
items: AudioRow[]      // List of all audio tracks
loading: boolean
error: string | null
total: number
```

**Actions:**
```ts
fetchLibrary()         // GET /api/audio
```

### playerStore

**Location:** `src/stores/player.ts`

**State:**
```ts
currentTrack: AudioRow | null   // What's playing
isPlaying: boolean               // Playback state
currentTime: number             // Seconds into track
duration: number                // Total track length
volume: number                  // 0-100
queue: AudioRow[]               // Up next
```

**Computed Properties:**
```ts
currentTrackUrl     // Full stream URL
progress             // 0-100 percentage
formattedCurrentTime // "1:23"
formattedDuration    // "4:56"
```

**Actions:**
```ts
play(track?)         // Start playback
pause()              // Pause
togglePlay()         // Flip state
seek(time)           // Jump to position
setVolume(v)         // Change volume
next()               // Skip to next
previous()           // Go back
addToQueue(track)    // Add to queue
```

## How Pinia Works Under the Hood

### Setup Syntax (What We Use)

Pinia offers two syntax styles. We use the **setup store** syntax:

```ts
export const useMangaStore = defineStore('manga', () => {
  // State as refs
  const items = ref<MangaRow[]>([])
  const loading = ref(false)

  // Computed
  const totalItems = computed(() => items.value.length)

  // Actions (plain functions)
  async function fetchLibrary() {
    loading.value = true
    // ...
    loading.value = false
  }

  // Return everything you want to expose
  return { items, loading, totalItems, fetchLibrary }
})
```

### What `defineStore` Does

1. **Creates a unique store instance** per Pinia installation
2. **Wraps your setup function** and manages its lifecycle
3. **Makes it reactive** - changes trigger Vue reactivity system

### How State Becomes Reactive

Inside the setup function:

```ts
const count = ref(0)  // Creates a reactive Ref
```

`ref()` wraps a value in a reactive proxy. When you access `store.count`, Pinia unwraps it automatically.

### Action Execution Flow

```ts
async function fetchLibrary() {
  // 1. Set loading state
  loading.value = true

  // 2. Await API call
  const response = await fetchMangaLibrary()

  // 3. Update state (triggers reactivity)
  items.value = response.items

  // 4. Clear loading
  loading.value = false
}
```

When `items.value` changes, Vue automatically re-renders any component that uses `mangaStore.items`.

### Singleton Pattern

Pinia stores are **singletons** - there's only one instance per store:

```ts
const store1 = useMangaStore()
const store2 = useMangaStore()

store1 === store2  // true - same object
```

This means calling `useMangaStore()` anywhere returns the same reactive instance.

## Store Access in Components

### In `<script setup>`

```ts
<script setup lang="ts">
import { useMangaStore } from '../stores/manga'

const mangaStore = useMangaStore()

// Template can access directly:
// {{ mangaStore.items }}
// {{ mangaStore.loading }}
</script>
```

### Reactivity

Changes to store state are automatically reflected in components:

```ts
// This is reactive - component re-renders when items change
const displayItems = computed(() => mangaStore.items.filter(...))
```

## When to Use Stores

### Use Stores For

- Global app state (user, theme, auth)
- Data fetched from APIs (shared across components)
- Complex state logic (player controls, filters)
- Caching API responses

### Don't Use Stores For

- Local component UI state (modal open/closed, form inputs)
- Props that flow down to children

## Pinia vs Vuex

Pinia is Vue 3's officially recommended state management (replacing Vuex):

| Vuex | Pinia |
|------|-------|
| Mutations + Actions | Actions only |
| One global store | Multiple stores |
| Complex modules | Simpler setup syntax |
| 4.x | 3.x (Vue 3) |

## API Client Pattern

Stores don't make API calls directly. They import from `src/api/`:

```
stores/manga.ts  →  api/manga.ts  →  api/client.ts  →  Backend
```

`src/api/client.ts` is our fetch wrapper with error handling:

```ts
async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: { 'Content-Type': 'application/json' }
  })

  if (!response.ok) {
    throw { message: response.statusText, status: response.status }
  }

  return response.json()
}
```

This keeps stores clean and testable.

## Debugging Stores

### Vue DevTools

Pinia is integrated into Vue DevTools. You can:
- Inspect state values
- See action calls
- Time travel debug (with plugins)

### Manual

```ts
// Watch for changes
watch(() => mangaStore.items, (newVal) => {
  console.log('Items changed:', newVal)
})
```

## Summary

| Concept | Description |
|---------|-------------|
| Store | Reactive singleton holding state + logic |
| `ref()` | Makes a value reactive |
| `computed()` | Derives value from state |
| Action | Function that modifies state |
| Setup syntax | Function-based store definition |
| Singleton | One instance shared app-wide |
