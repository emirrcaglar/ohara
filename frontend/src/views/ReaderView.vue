<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMangaStore } from '../stores/manga'

const route = useRoute()
const router = useRouter()
const mangaStore = useMangaStore()

const mangaId = computed(() => Number(route.query.manga))
const currentPage = ref(Number(route.query.page) || 0)
const totalPages = ref(Number(route.query.total) || 0)

const pageUrl = computed(() => `/manga/${mangaId.value}/page/${currentPage.value}`)

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
  <div class="flex-1 flex flex-col bg-black overflow-hidden">
    <div class="flex-1 flex items-center justify-center overflow-auto p-4">
      <img
        :src="pageUrl"
        :alt="`Page ${currentPage + 1}`"
        class="max-w-none max-h-full w-auto h-auto object-contain"
      />
    </div>

    <div class="bg-surface-container-low border-t border-white/10 px-8 py-4">
      <div class="flex items-center justify-between max-w-4xl mx-auto">
        <button
          class="px-6 py-3 bg-surface-container-high text-on-surface font-bold uppercase tracking-wider disabled:opacity-30 disabled:cursor-not-allowed hover:bg-primary-container hover:text-on-primary-container transition-colors"
          :disabled="!hasPrev"
          @click="prevPage"
        >
          <span class="flex items-center gap-2">
            <span class="material-symbols-outlined">arrow_back</span>
            Prev
          </span>
        </button>

        <div class="text-center">
          <p class="text-sm font-mono text-secondary uppercase tracking-widest mb-1">Page</p>
          <p class="text-2xl font-black text-on-surface">{{ progressText }}</p>
        </div>

        <button
          class="px-6 py-3 bg-surface-container-high text-on-surface font-bold uppercase tracking-wider disabled:opacity-30 disabled:cursor-not-allowed hover:bg-primary-container hover:text-on-primary-container transition-colors"
          :disabled="!hasNext"
          @click="nextPage"
        >
          <span class="flex items-center gap-2">
            Next
            <span class="material-symbols-outlined">arrow_forward</span>
          </span>
        </button>
      </div>

      <div class="text-center mt-4">
        <div class="h-1 bg-surface-container-highest rounded-full overflow-hidden">
          <div
            class="h-full bg-primary-container transition-all duration-200"
            :style="{ width: `${((currentPage + 1) / totalPages) * 100}%` }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>
