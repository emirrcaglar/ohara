import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { AudioRow } from '../types/api'
import { fetchAudioLibrary } from '../api/audio'

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

  return {
    items,
    loading,
    error,
    total,
    totalItems,
    fetchLibrary
  }
})
