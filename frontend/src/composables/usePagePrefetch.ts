import { ref, watch, type Ref } from 'vue'

const PREFETCH_COUNT = 5
const MOBILE_PREFETCH_COUNT = 2
const PREFETCH_DEBOUNCE_MS = 300
const PAGE_LOADING_SKELETON_DELAY_MS = 50

export function usePagePrefetch(
  mangaId: Ref<number>,
  currentPage: Ref<number>,
  totalPages: Ref<number>,
  getPageUrl: (page: number) => string,
) {
  let prefetchTimeoutId: ReturnType<typeof setTimeout> | undefined
  let skeletonTimeoutId: ReturnType<typeof setTimeout> | undefined
  const loadedPages = new Set<number>()
  const showPageSkeleton = ref(false)
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
    clearTimeout(skeletonTimeoutId)
    currentMainImageLoaded = loadedPages.has(currentPage.value)
    showPageSkeleton.value = false

    if (!currentMainImageLoaded) {
      skeletonTimeoutId = setTimeout(() => {
        if (!currentMainImageLoaded) {
          showPageSkeleton.value = true
        }
      }, PAGE_LOADING_SKELETON_DELAY_MS)
    }

    prefetchTimeoutId = setTimeout(() => {
      if (currentMainImageLoaded) {
        prefetchPages()
      }
    }, PREFETCH_DEBOUNCE_MS)
  }

  function onMainImageLoaded(page = currentPage.value) {
    loadedPages.add(page)

    if (page !== currentPage.value) {
      return
    }

    currentMainImageLoaded = true
    clearTimeout(skeletonTimeoutId)
    showPageSkeleton.value = false

    if (prefetchTimeoutId !== undefined) {
      clearTimeout(prefetchTimeoutId)
      prefetchPages()
    }
  }

  watch(currentPage, () => {
    prefetchPagesDebounced()
  })

  function cleanup() {
    clearTimeout(prefetchTimeoutId)
    clearTimeout(skeletonTimeoutId)
  }

  return {
    shouldLoadMobilePage,
    getMobilePageSrc,
    prefetchPages,
    showPageSkeleton,
    onMainImageLoaded,
    cleanup,
  }
}
