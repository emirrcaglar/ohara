import { ref, computed, type Ref } from 'vue'
import type { Router } from 'vue-router'
import { useMangaStore } from '../stores/manga'

export function usePageNavigation(
  mangaId: Ref<number>,
  initialPage: number,
  totalPages: Ref<number>,
  rightToLeftSwipe: Ref<boolean>,
  router: Router,
) {
  const mangaStore = useMangaStore()
  const currentPage = ref(initialPage)

  const currentVisualPage = computed(() => pageToVisualPage(currentPage.value))
  const hasVisualLeft = computed(() => currentVisualPage.value > 0)
  const hasVisualRight = computed(() => currentVisualPage.value < totalPages.value - 1)
  const visualLeftLabel = computed(() => (rightToLeftSwipe.value ? 'NEXT' : 'PREV'))
  const visualRightLabel = computed(() => (rightToLeftSwipe.value ? 'PREV' : 'NEXT'))

  let saveProgressTimeoutId: ReturnType<typeof setTimeout> | undefined

  function pageToVisualPage(page: number) {
    if (totalPages.value <= 0) return page
    return rightToLeftSwipe.value ? totalPages.value - 1 - page : page
  }

  function visualPageToPage(visualPage: number) {
    if (totalPages.value <= 0) return visualPage
    return rightToLeftSwipe.value ? totalPages.value - 1 - visualPage : visualPage
  }

  function getVisualPageIndex(visualPageNumber: number) {
    return visualPageToPage(visualPageNumber - 1)
  }

  function saveProgress() {
    clearTimeout(saveProgressTimeoutId)
    saveProgressTimeoutId = setTimeout(async () => {
      await mangaStore.updateProgress(mangaId.value, currentPage.value)
    }, 500)
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

  function commitPage(page: number, onCommit?: () => void) {
    const maxPage = Math.max(totalPages.value - 1, 0)
    const next = Math.min(Math.max(page, 0), maxPage)
    if (next === currentPage.value) {
      onCommit?.()
      return
    }

    currentPage.value = next
    onCommit?.()
    navigate()
  }

  function goVisualLeft(onCommit?: () => void) {
    if (hasVisualLeft.value) {
      commitPage(visualPageToPage(currentVisualPage.value - 1), onCommit)
    }
  }

  function goVisualRight(onCommit?: () => void) {
    if (hasVisualRight.value) {
      commitPage(visualPageToPage(currentVisualPage.value + 1), onCommit)
    }
  }

  function cleanup() {
    clearTimeout(saveProgressTimeoutId)
  }

  return {
    currentPage,
    currentVisualPage,
    hasVisualLeft,
    hasVisualRight,
    visualLeftLabel,
    visualRightLabel,
    pageToVisualPage,
    visualPageToPage,
    getVisualPageIndex,
    commitPage,
    goVisualLeft,
    goVisualRight,
    cleanup,
  }
}
