import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { MangaRow, MangaInfo } from '../types/api'
import { fetchMangaLibrary, fetchMangaInfo, saveMangaProgress } from '../api/manga'

export const useMangaStore = defineStore('manga', () => {
  const items = ref<MangaRow[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  const totalItems = computed(() => items.value.length)

  async function fetchLibrary() {
    loading.value = true
    error.value = null
    try {
      const response = await fetchMangaLibrary()
      items.value = response.items
      total.value = response.total
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch manga library'
    } finally {
      loading.value = false
    }
  }

  async function getMangaInfo(id: number): Promise<MangaInfo | null> {
    try {
      return await fetchMangaInfo(id)
    } catch {
      return null
    }
  }

  async function updateProgress(id: number, page: number) {
    try {
      await saveMangaProgress(id, page)
      const item = items.value.find(m => m.id === id)
      if (item) {
        item.currentPage = page
      }
    } catch (e) {
      console.error('Failed to save progress:', e)
    }
  }

  return {
    items,
    loading,
    error,
    total,
    totalItems,
    fetchLibrary,
    getMangaInfo,
    updateProgress
  }
})
