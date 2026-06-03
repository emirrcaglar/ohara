import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchPreferences, savePreference } from '../api/preferences'

const RIGHT_TO_LEFT_SWIPE_FOR_MANGA_KEY = 'reader.rightToLeftSwipeForManga'
const SCROLL_READING_FOR_MANGA_KEY = 'reader.scrollReadingForManga'

export const usePreferencesStore = defineStore('preferences', () => {
  const preferences = ref<Record<string, string>>({})
  const isLoading = ref(false)
  const hasLoaded = ref(false)
  const error = ref<string | null>(null)

  const rightToLeftSwipeForManga = computed(
    () => preferences.value[RIGHT_TO_LEFT_SWIPE_FOR_MANGA_KEY] === 'true',
  )
  const scrollReadingForManga = computed(
    () => preferences.value[SCROLL_READING_FOR_MANGA_KEY] === 'true',
  )

  async function loadPreferences() {
    if (isLoading.value) return

    isLoading.value = true
    error.value = null

    try {
      const response = await fetchPreferences()
      preferences.value = response.preferences
      hasLoaded.value = true
    } catch (e) {
      console.error('Failed to load preferences:', e)
      error.value = 'FAILED_TO_LOAD_PREFERENCES'
    } finally {
      isLoading.value = false
    }
  }

  async function setBooleanPreference(key: string, value: boolean) {
    const previousValue = preferences.value[key]
    preferences.value = {
      ...preferences.value,
      [key]: value ? 'true' : 'false',
    }
    error.value = null

    try {
      await savePreference(key, value ? 'true' : 'false')
    } catch (e) {
      console.error('Failed to save preference:', e)

      const nextPreferences = { ...preferences.value }
      if (previousValue === undefined) {
        delete nextPreferences[key]
      } else {
        nextPreferences[key] = previousValue
      }
      preferences.value = nextPreferences
      error.value = 'FAILED_TO_SAVE_PREFERENCE'
      throw e
    }
  }

  async function setRightToLeftSwipeForManga(value: boolean) {
    await setBooleanPreference(RIGHT_TO_LEFT_SWIPE_FOR_MANGA_KEY, value)
  }

  async function setScrollReadingForManga(value: boolean) {
    await setBooleanPreference(SCROLL_READING_FOR_MANGA_KEY, value)
  }

  return {
    preferences,
    isLoading,
    hasLoaded,
    error,
    rightToLeftSwipeForManga,
    scrollReadingForManga,
    loadPreferences,
    setRightToLeftSwipeForManga,
    setScrollReadingForManga,
  }
})
