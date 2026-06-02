import { watch, type Ref } from 'vue'

const PREFETCH_COUNT = 5
const MOBILE_PREFETCH_COUNT = 2
const PREFETCH_DEBOUNCE_MS = 300

export function usePagePrefetch(
  mangaId: Ref<number>,
  currentPage: Ref<number>,
  totalPages: Ref<number>,
  getPageUrl: (page: number) => string,
) {
  let prefetchTimeoutId: ReturnType<typeof setTimeout> | undefined
  const loadedPages = new Set<number>()
  let currentMainImageLoaded = false

  function shouldLoadMobilePage(page: number) {
    return Math.abs(page - currentPage.value) <= MOBILE_PREFETCH_COUNT
  }

  function getMobilePageSrc(page: number) {
    return shouldLoadMobilePage(page) ? getPageUrl(page) : undefined
  }

  function prefetchPages() {
    const count = window.matchMedia('(max-width: 767px)').matches
      ? MOBILE_PREFETCH_COUNT
      : PREFETCH_COUNT

    for (let i = 1; i <= count; i++) {
      const ahead = currentPage.value + i
      if (ahead < totalPages.value && !loadedPages.has(ahead)) {
        const img = new Image()
        img.onload = () => loadedPages.add(ahead)
        img.src = getPageUrl(ahead)
      }

      const behind = currentPage.value - i
      if (behind >= 0 && !loadedPages.has(behind)) {
        const img = new Image()
        img.onload = () => loadedPages.add(behind)
        img.src = getPageUrl(behind)
      }
    }
  }

  function prefetchPagesDebounced() {
    clearTimeout(prefetchTimeoutId)
    currentMainImageLoaded = false

    prefetchTimeoutId = setTimeout(() => {
      if (currentMainImageLoaded) {
        prefetchPages()
      }
    }, PREFETCH_DEBOUNCE_MS)
  }

  function onMainImageLoaded() {
    currentMainImageLoaded = true

    if (prefetchTimeoutId !== undefined) {
      clearTimeout(prefetchTimeoutId)
      prefetchPages()
    }
  }

  watch(currentPage, () => {
    loadedPages.delete(currentPage.value)
    prefetchPagesDebounced()
  })

  function cleanup() {
    clearTimeout(prefetchTimeoutId)
  }

  return {
    shouldLoadMobilePage,
    getMobilePageSrc,
    prefetchPages,
    onMainImageLoaded,
    cleanup,
  }
}
