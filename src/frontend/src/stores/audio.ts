import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { AudioRow } from '../types/api'
import { fetchAudioLibrary, moveAudioToCatalog } from '../api/audio'

export const useAudioStore = defineStore('audio', () => {
  const items = ref<AudioRow[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  const totalItems = computed(() => items.value.length)

  async function fetchLibrary() {
    loading.value = true
    error.value = null
    try {
      const response = await fetchAudioLibrary()
      items.value = response.items
      total.value = response.total
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch audio library'
    } finally {
      loading.value = false
    }
  }

  async function moveAudio(id: number, catalogId: number | null) {
    const item = items.value.find((a) => a.id === id)
    const previousCatalogId = item?.catalogId
    if (item) item.catalogId = catalogId

    try {
      await moveAudioToCatalog(id, catalogId)
    } catch (e) {
      if (item) item.catalogId = previousCatalogId ?? null
      error.value = e instanceof Error ? e.message : 'Failed to move audio'
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
    moveAudio,
  }
})
