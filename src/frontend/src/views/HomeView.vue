<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, BookOpen, History, RefreshCw } from 'lucide-vue-next'
import { fetchMangaLibrary, getMangaCoverUrl } from '../api/manga'
import type { MangaRow } from '../types/api'
import { mangaTitle, mangaSubtitle } from '../utils/mangaMetadata'

const library = ref<MangaRow[]>([])
const chapters = computed(() =>
  library.value
    .filter(
      (chapter): chapter is MangaRow & { updatedAt: string } =>
        chapter.pageCount > 0 &&
        typeof chapter.updatedAt === 'string' &&
        Number.isFinite(Date.parse(chapter.updatedAt)),
    )
    .sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt) || b.id - a.id)
    .slice(0, 24),
)
const loading = ref(true)
const error = ref(false)
const failedCovers = ref(new Set<number>())
const latest = computed(() => chapters.value[0])
const earlier = computed(() => chapters.value.slice(1))
const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

async function loadHistory() {
  loading.value = true
  error.value = false
  try {
    library.value = (await fetchMangaLibrary()).items
    failedCovers.value.clear()
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function pageNumber(chapter: MangaRow) {
  return Math.min(chapter.pageCount, Math.max(1, chapter.currentPage + 1))
}

function progress(chapter: MangaRow) {
  return chapter.pageCount > 0 ? Math.round((pageNumber(chapter) / chapter.pageCount) * 100) : 0
}

function readerLink(chapter: MangaRow) {
  return {
    path: '/reader',
    query: { manga: chapter.id, page: pageNumber(chapter) - 1, total: chapter.pageCount },
  }
}

function readDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? 'Recently read' : dateFormatter.format(date)
}

onMounted(loadHistory)
</script>

<template>
  <section class="flex-1 overflow-y-auto p-4 md:p-8 lg:p-12" aria-labelledby="home-title">
    <div class="mx-auto max-w-7xl">
      <header class="mb-8 flex flex-wrap items-end justify-between gap-6">
        <div>
          <p class="mb-3 text-xs font-bold uppercase tracking-[0.2em] text-secondary">
            Your reading history
          </p>
          <h1 id="home-title" class="text-4xl font-bold uppercase tracking-tighter md:text-5xl">
            Latest read chapters
          </h1>
          <p class="mt-3 text-sm text-on-surface-variant">Pick up where you left off.</p>
        </div>
        <RouterLink
          to="/library"
          class="home-link flex items-center gap-3 bg-surface-container-high px-5 py-3 text-xs font-bold uppercase tracking-wider hover:bg-surface-container-highest"
        >
          Browse library <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </RouterLink>
      </header>

      <div
        v-if="loading"
        role="status"
        class="bg-surface-container-low p-8 text-sm text-on-surface-variant"
      >
        Loading reading history…
      </div>

      <div v-else-if="error" role="alert" class="bg-surface-container-low p-8">
        <h2 class="text-lg font-bold uppercase">Reading history is unavailable</h2>
        <p class="mt-2 text-sm text-on-surface-variant">
          We couldn’t load your latest chapters. Please try again.
        </p>
        <button
          type="button"
          class="home-link mt-6 inline-flex items-center gap-2 bg-primary-container px-5 py-3 text-sm font-bold text-on-primary-container hover:bg-primary"
          @click="loadHistory"
        >
          <RefreshCw class="h-4 w-4" aria-hidden="true" /> Try again
        </button>
      </div>

      <div v-else-if="!latest" class="bg-surface-container-low px-6 py-16 text-center md:py-24">
        <BookOpen class="mx-auto mb-6 h-12 w-12 text-primary-container" aria-hidden="true" />
        <h2 class="text-2xl font-bold uppercase tracking-tight">Your next chapter starts here</h2>
        <p class="mx-auto mt-3 max-w-md text-sm leading-relaxed text-on-surface-variant">
          Open a chapter from your library. The chapters you read will appear here, with your place
          saved.
        </p>
        <RouterLink
          to="/library"
          class="home-link mt-8 inline-flex items-center gap-3 bg-primary-container px-6 py-4 text-sm font-bold uppercase text-on-primary-container hover:bg-primary"
        >
          Open library <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </RouterLink>
      </div>

      <template v-else>
        <RouterLink
          :to="readerLink(latest)"
          class="home-link group mb-12 grid bg-surface-container-low hover:bg-surface-container-high sm:grid-cols-[minmax(0,1fr)_2fr] lg:grid-cols-[minmax(0,1fr)_3fr]"
        >
          <div
            class="relative flex aspect-3/4 max-h-80 items-center justify-center overflow-hidden bg-surface-container-lowest sm:max-h-none"
          >
            <BookOpen class="absolute h-12 w-12 text-outline" aria-hidden="true" />
            <img
              v-if="!failedCovers.has(latest.id)"
              :src="getMangaCoverUrl(latest.id)"
              alt=""
              class="relative h-full w-full object-cover"
              decoding="async"
              @error="failedCovers.add(latest.id)"
            />
          </div>
          <div class="flex min-w-0 flex-col items-start justify-between gap-8 p-6 md:p-8 lg:p-10">
            <div class="w-full">
              <p
                class="mb-5 flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-secondary"
              >
                <History class="h-4 w-4" aria-hidden="true" /> Last read
              </p>
              <h2
                :title="latest.fileName || latest.title"
                class="break-words text-2xl font-bold uppercase leading-tight tracking-tighter group-hover:text-primary md:text-4xl"
              >
                {{ mangaTitle(latest) }}
              </h2>
              <p v-if="mangaSubtitle(latest)" class="mt-3 text-sm text-on-surface-variant">
                {{ mangaSubtitle(latest) }}
              </p>
              <time
                :datetime="latest.updatedAt"
                class="mt-4 block text-xs text-on-surface-variant"
                >{{ readDate(latest.updatedAt) }}</time
              >
            </div>
            <div class="w-full">
              <div
                class="mb-3 flex flex-wrap justify-between gap-2 text-xs text-on-surface-variant"
              >
                <span>Page {{ pageNumber(latest) }} of {{ latest.pageCount }}</span>
                <span class="font-bold text-secondary">{{
                  pageNumber(latest) === latest.pageCount
                    ? 'Last page reached'
                    : `${progress(latest)}% read`
                }}</span>
              </div>
              <div class="mb-6 h-1 bg-surface-container-highest" aria-hidden="true">
                <div
                  class="h-full bg-primary-container"
                  :style="{ width: `${progress(latest)}%` }"
                ></div>
              </div>
              <span
                class="inline-flex items-center gap-4 bg-primary-container px-6 py-4 text-sm font-bold uppercase text-on-primary-container group-hover:bg-primary"
                >Resume reading <ArrowRight class="h-4 w-4" aria-hidden="true"
              /></span>
            </div>
          </div>
        </RouterLink>

        <section v-if="earlier.length" aria-labelledby="earlier-title">
          <div class="mb-6 flex flex-wrap items-center justify-between gap-2">
            <h2 id="earlier-title" class="text-xl font-bold uppercase tracking-tight">
              Previously read
            </h2>
            <span class="text-xs text-on-surface-variant"
              >Most recent first · Latest {{ chapters.length }} chapters</span
            >
          </div>
          <div class="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-6 xl:grid-cols-4">
            <RouterLink
              v-for="chapter in earlier"
              :key="chapter.id"
              :to="readerLink(chapter)"
              class="home-link group flex min-w-0 flex-col bg-surface-container-low hover:bg-surface-container-high"
            >
              <div
                class="relative flex aspect-3/4 items-center justify-center overflow-hidden bg-surface-container-lowest"
              >
                <BookOpen class="absolute h-10 w-10 text-outline" aria-hidden="true" />
                <img
                  v-if="!failedCovers.has(chapter.id)"
                  :src="getMangaCoverUrl(chapter.id)"
                  alt=""
                  class="relative h-full w-full object-cover"
                  loading="lazy"
                  decoding="async"
                  @error="failedCovers.add(chapter.id)"
                />
                <span
                  v-if="pageNumber(chapter) === chapter.pageCount"
                  class="absolute right-2 top-2 bg-secondary-container px-2 py-1 text-[10px] font-bold uppercase text-on-secondary-container"
                  >Last page reached</span
                >
              </div>
              <div class="flex flex-1 flex-col gap-3 p-3 md:p-5">
                <h3
                  :title="chapter.fileName || chapter.title"
                  class="break-words text-sm font-bold uppercase leading-snug tracking-tight group-hover:text-primary md:text-base"
                >
                  {{ mangaTitle(chapter) }}
                </h3>
                <p v-if="mangaSubtitle(chapter)" class="text-xs text-on-surface-variant">
                  {{ mangaSubtitle(chapter) }}
                </p>
                <time :datetime="chapter.updatedAt" class="text-xs text-on-surface-variant">{{
                  readDate(chapter.updatedAt)
                }}</time>
                <div class="mt-auto pt-2">
                  <div class="mb-3 h-1 bg-surface-container-highest" aria-hidden="true">
                    <div
                      class="h-full bg-primary-container"
                      :style="{ width: `${progress(chapter)}%` }"
                    ></div>
                  </div>
                  <div
                    class="flex items-center justify-between gap-2 text-xs text-on-surface-variant"
                  >
                    <span>Page {{ pageNumber(chapter) }} / {{ chapter.pageCount }}</span
                    ><ArrowRight class="h-4 w-4 shrink-0 text-primary" aria-hidden="true" />
                  </div>
                </div>
              </div>
            </RouterLink>
          </div>
        </section>
      </template>
    </div>
  </section>
</template>

<style scoped>
.home-link:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 4px;
}
</style>
