<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMangaStore } from '../stores/manga'
import { API_BASE } from '../api/client'

const route = useRoute()
const router = useRouter()
const mangaStore = useMangaStore()

const mangaId = computed(() => Number(route.query.manga))
const currentPage = ref(Number(route.query.page) || 0)
const totalPages = ref(Number(route.query.total) || 0)

const pageUrl = computed(() => `${API_BASE}/manga/${mangaId.value}/page/${currentPage.value}`)

const hasPrev = computed(() => currentPage.value > 0)
const hasNext = computed(() => currentPage.value < totalPages.value - 1)

async function saveProgress() {
  await mangaStore.updateProgress(mangaId.value, currentPage.value)
}

function prevPage() {
  if (hasPrev.value) {
    currentPage.value--
    navigate()
  }
}

function nextPage() {
  if (hasNext.value) {
    currentPage.value++
    navigate()
  }
}

function navigate() {
  router.replace({
    query: {
      manga: mangaId.value,
      page: currentPage.value,
      total: totalPages.value,
    },
  })
  saveProgress()
}

const PREFETCH_COUNT = 5

function prefetchPages() {
  for (let i = 1; i <= PREFETCH_COUNT; i++) {
    const ahead = currentPage.value + i
    if (ahead < totalPages.value) {
      const img = new Image()
      img.src = `${API_BASE}/manga/${mangaId.value}/page/${ahead}`
    }

    const behind = currentPage.value - i
    if (behind >= 0) {
      const img = new Image()
      img.src = `${API_BASE}/manga/${mangaId.value}/page/${behind}`
    }
  }
}

watch(currentPage, prefetchPages)

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowLeft' || e.key === 'a') {
    prevPage()
  } else if (e.key === 'ArrowRight' || e.key === 'd') {
    nextPage()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeydown)

  if (totalPages.value === 0 && mangaId.value) {
    const info = await mangaStore.getMangaInfo(mangaId.value)
    if (info) {
      totalPages.value = info.pageCount
    }
  }

  prefetchPages()
})
</script>

<template>
  <div class="relative flex-1 flex flex-col bg-black overflow-clip overscroll-none">
    <div class="flex-1 min-h-0 p-2 md:p-4">
      <div class="relative w-full h-full">
        <img
          :src="pageUrl"
          :alt="`Page ${currentPage + 1}`"
          class="absolute inset-0 w-full h-full object-contain"
        />
      </div>
    </div>

    <div
      class="fixed right-2 top-16 z-60 flex items-center bg-surface-container-lowest border-2 border-primary-container p-0.5 shadow-[0_0_40px_rgba(0,0,0,0.8)] sm:right-3 md:right-6 md:top-20"
      :style="{ marginTop: 'env(safe-area-inset-top)' }"
    >
      <div class="flex items-center gap-0.5 sm:gap-1">
        <button
          class="flex min-h-9 items-center bg-primary-container px-2 py-1.5 text-[9px] font-black uppercase tracking-[0.12em] text-on-primary-container transition-all hover:bg-primary disabled:opacity-30 disabled:cursor-not-allowed sm:min-h-10 sm:px-3 sm:text-[10px] md:px-4 md:text-xs md:tracking-[0.16em]"
          :disabled="!hasPrev"
          @click="prevPage"
        >
          <span class="material-symbols-outlined mr-0.5 text-[13px] sm:mr-1 sm:text-sm"
            >chevron_left</span
          >
          PREV
        </button>

        <div
          class="flex min-h-9 items-center gap-1.5 bg-surface px-2 py-1.5 font-mono sm:min-h-10 sm:gap-2 sm:px-3 md:gap-3 md:px-4"
        >
          <span
            class="text-[8px] font-bold uppercase tracking-widest text-outline sm:text-[9px] md:text-[10px]"
          >
            PAGE
          </span>
          <span class="text-[11px] font-bold tracking-tighter text-primary sm:text-xs">
            {{ currentPage + 1 }}
          </span>
          <span class="text-outline-variant">/</span>
          <span class="text-[11px] font-bold tracking-tighter text-on-surface sm:text-xs">{{
            totalPages
          }}</span>
        </div>

        <button
          class="flex min-h-9 items-center bg-primary-container px-2 py-1.5 text-[9px] font-black uppercase tracking-[0.12em] text-on-primary-container transition-all hover:bg-primary disabled:opacity-30 disabled:cursor-not-allowed sm:min-h-10 sm:px-3 sm:text-[10px] md:px-4 md:text-xs md:tracking-[0.16em]"
          :disabled="!hasNext"
          @click="nextPage"
        >
          NEXT
          <span class="material-symbols-outlined ml-0.5 text-[13px] sm:ml-1 sm:text-sm"
            >chevron_right</span
          >
        </button>
      </div>
    </div>
  </div>
</template>
