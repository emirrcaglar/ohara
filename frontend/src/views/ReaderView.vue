<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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

const progressText = computed(() => `${currentPage.value + 1} / ${totalPages.value}`)

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
      total: totalPages.value
    }
  })
  saveProgress()
}

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
})
</script>

<template>
  <div class="relative flex-1 flex flex-col bg-black overflow-hidden">
    <div class="flex-1 flex items-center justify-center overflow-auto p-4">
      <img
        :src="pageUrl"
        :alt="`Page ${currentPage + 1}`"
        class="max-w-none max-h-full w-auto h-auto object-contain"
      />
    </div>

    <div class="absolute top-6 right-6 z-30 w-[340px] bg-surface-container-low/95 border border-white/10 p-4 backdrop-blur-sm">
      <div class="flex items-center justify-between gap-3">
        <button
          class="px-4 py-2 bg-surface-container-high text-on-surface font-bold uppercase tracking-wider disabled:opacity-30 disabled:cursor-not-allowed hover:bg-primary-container hover:text-on-primary-container transition-colors"
          :disabled="!hasPrev"
          @click="prevPage"
        >
          <span class="flex items-center gap-2">
            <span class="material-symbols-outlined">arrow_back</span>
            Prev
          </span>
        </button>

        <div class="text-center">
          <p class="text-[10px] font-mono text-secondary uppercase tracking-widest mb-1">Page</p>
          <p class="text-lg font-black text-on-surface">{{ progressText }}</p>
        </div>

        <button
          class="px-4 py-2 bg-surface-container-high text-on-surface font-bold uppercase tracking-wider disabled:opacity-30 disabled:cursor-not-allowed hover:bg-primary-container hover:text-on-primary-container transition-colors"
          :disabled="!hasNext"
          @click="nextPage"
        >
          <span class="flex items-center gap-2">
            Next
            <span class="material-symbols-outlined">arrow_forward</span>
          </span>
        </button>
      </div>

      <div class="text-center mt-3">
        <div class="h-1 bg-surface-container-highest rounded-full overflow-hidden">
          <div
            class="h-full bg-primary-container transition-all duration-200"
            :style="{ width: `${totalPages > 0 ? ((currentPage + 1) / totalPages) * 100 : 0}%` }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>
