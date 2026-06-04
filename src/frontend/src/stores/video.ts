import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { VideoInfo, VideoRow, VideoStateUpdate } from '../types/api'
import {
  deleteVideo,
  fetchVideoInfo,
  fetchVideoLibrary,
  moveVideoToCatalog,
  saveVideoState,
} from '../api/video'

export const useVideoStore = defineStore('video', () => {
  const items = ref<VideoRow[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  const totalItems = computed(() => items.value.length)

  async function fetchLibrary() {
    loading.value = true
    error.value = null
    try {
      const response = await fetchVideoLibrary()
      items.value = response.items
      total.value = response.total
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch video library'
    } finally {
      loading.value = false
    }
  }

  async function getVideoInfo(id: number): Promise<VideoInfo | null> {
    try {
      return await fetchVideoInfo(id)
    } catch {
      return null
    }
  }

  async function updateVideoState(id: number, state: VideoStateUpdate) {
    await saveVideoState(id, state)

    const item = items.value.find((v) => v.id === id)
    if (!item) return

    item.duration = state.duration || item.duration
    item.width = state.width || item.width
    item.height = state.height || item.height
    item.position = state.position
    item.completed = state.completed
    item.lastError = state.lastError
  }

  async function moveVideo(id: number, catalogId: number | null) {
    const item = items.value.find((v) => v.id === id)
    const previousCatalogId = item?.catalogId
    if (item) item.catalogId = catalogId

    try {
      await moveVideoToCatalog(id, catalogId)
    } catch (e) {
      if (item) item.catalogId = previousCatalogId ?? null
      error.value = e instanceof Error ? e.message : 'Failed to move video'
      throw e
    }
  }

  async function removeVideo(id: number) {
    const previousItems = items.value
    const previousTotal = total.value

    items.value = items.value.filter((v) => v.id !== id)
    total.value = items.value.length

    try {
      await deleteVideo(id)
    } catch (e) {
      items.value = previousItems
      total.value = previousTotal
      error.value = e instanceof Error ? e.message : 'Failed to delete video'
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
    getVideoInfo,
    updateVideoState,
    moveVideo,
    removeVideo,
  }
})
