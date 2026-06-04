import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { MangaRow, MangaInfo } from '../types/api'
import {
  deleteManga,
  fetchMangaInfo,
  fetchMangaLibrary,
  moveMangaToCatalog,
  saveMangaProgress,
} from '../api/manga'

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
      const item = items.value.find((m) => m.id === id)
      if (item) {
        item.currentPage = page
      }
    } catch (e) {
      console.error('Failed to save progress:', e)
    }
  }

  async function moveManga(id: number, catalogId: number | null) {
    const item = items.value.find((m) => m.id === id)
    const previousCatalogId = item?.catalogId
    if (item) item.catalogId = catalogId

    try {
      await moveMangaToCatalog(id, catalogId)
    } catch (e) {
      if (item) item.catalogId = previousCatalogId ?? null
      error.value = e instanceof Error ? e.message : 'Failed to move manga'
      throw e
    }
  }

  async function removeManga(id: number) {
    const previousItems = items.value
    const previousTotal = total.value

    items.value = items.value.filter((m) => m.id !== id)
    total.value = items.value.length

    try {
      await deleteManga(id)
    } catch (e) {
      items.value = previousItems
      total.value = previousTotal
      error.value = e instanceof Error ? e.message : 'Failed to delete manga'
      throw e
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
    updateProgress,
    moveManga,
    removeManga,
  }
})
