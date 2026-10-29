import { onBeforeUnmount, onMounted, ref, shallowReactive, watch, type Ref } from 'vue'

const PREFETCH_AHEAD = 5
const PREFETCH_BEHIND = 5

interface BufferedPage {
  url: string
  controller: AbortController
  src?: string
  objectUrl?: string
}

export function useReaderPageLoading(
  currentPage: Ref<number>,
  totalPages: Ref<number>,
  scrollReading: Ref<boolean>,
  getPageUrl: (page: number) => string,
) {
  const showPageSkeleton = ref(false)
  const isMobile = ref(window.innerWidth < 768)
  const bufferedPages = shallowReactive(new Map<number, BufferedPage>())
  const loadedSources = new Set<string>()
  let skeletonTimer: ReturnType<typeof setTimeout> | undefined

  function isPageInPrefetchWindow(page: number) {
    return (
      page >= Math.max(0, currentPage.value - PREFETCH_BEHIND) &&
      page <= Math.min(totalPages.value - 1, currentPage.value + PREFETCH_AHEAD)
    )
  }

  function releasePage(page: number, buffered: BufferedPage) {
    bufferedPages.delete(page)
    buffered.controller.abort()
    if (buffered.src) loadedSources.delete(buffered.src)
    if (buffered.objectUrl) URL.revokeObjectURL(buffered.objectUrl)
  }

  async function loadPage(page: number) {
    if (bufferedPages.has(page)) return

    const buffered: BufferedPage = { url: getPageUrl(page), controller: new AbortController() }
    bufferedPages.set(page, buffered)
    try {
      // Keep nearby originals ready for display while allowing the browser's
      // private HTTP cache to serve pages evicted from this bounded buffer.
      const response = await fetch(buffered.url, {
        credentials: 'include',
        cache: 'default',
        signal: buffered.controller.signal,
        priority: page === currentPage.value ? 'high' : 'low',
      })
      if (!response.ok) throw new Error('Failed to load page')
      const blob = await response.blob()
      if (bufferedPages.get(page) !== buffered) return

      const objectUrl = URL.createObjectURL(blob)
      bufferedPages.set(page, { ...buffered, src: objectUrl, objectUrl })
    } catch {
      if (bufferedPages.get(page) !== buffered) return
      // Preserve normal image loading/error behavior if buffering fails.
      bufferedPages.set(page, { ...buffered, src: buffered.url })
    }
  }

  function updatePrefetchWindow() {
    for (const [page, buffered] of bufferedPages) {
      if (
        scrollReading.value ||
        !isPageInPrefetchWindow(page) ||
        buffered.url !== getPageUrl(page)
      ) {
        releasePage(page, buffered)
      }
    }
    // Scroll mode keeps its image elements mounted and eagerly loads this same
    // window, so it can reuse those elements without a second set of requests.
    if (scrollReading.value || !isPageInPrefetchWindow(currentPage.value)) return

    void loadPage(currentPage.value)
    for (let offset = 1; offset <= Math.max(PREFETCH_AHEAD, PREFETCH_BEHIND); offset++) {
      const ahead = currentPage.value + offset
      const behind = currentPage.value - offset
      if (isPageInPrefetchWindow(ahead)) void loadPage(ahead)
      if (isPageInPrefetchWindow(behind)) void loadPage(behind)
    }
  }

  function getPageSrc(page: number) {
    return bufferedPages.get(page)?.src
  }

  function updateDisplay() {
    isMobile.value = window.innerWidth < 768
  }

  function getMobilePageSrc(page: number) {
    return isMobile.value && isPageInPrefetchWindow(page) ? getPageSrc(page) : undefined
  }

  function onMainImageLoaded(page = currentPage.value) {
    const src = getPageSrc(page)
    if (src) loadedSources.add(src)
    if (page !== currentPage.value) return
    clearTimeout(skeletonTimer)
    showPageSkeleton.value = false
  }

  function cleanup() {
    clearTimeout(skeletonTimer)
    for (const [page, buffered] of bufferedPages) releasePage(page, buffered)
  }

  watch(
    [() => getPageUrl(currentPage.value), scrollReading],
    () => {
      clearTimeout(skeletonTimer)
      showPageSkeleton.value = false
      const buffered = bufferedPages.get(currentPage.value)
      const alreadyLoaded =
        buffered?.url === getPageUrl(currentPage.value) &&
        buffered.src &&
        loadedSources.has(buffered.src)
      if (!scrollReading.value && !alreadyLoaded) {
        skeletonTimer = setTimeout(() => {
          showPageSkeleton.value = true
        }, 50)
      }
    },
    { immediate: true },
  )

  watch(
    [currentPage, totalPages, scrollReading, () => getPageUrl(currentPage.value)],
    updatePrefetchWindow,
    { immediate: true },
  )

  onMounted(() => window.addEventListener('resize', updateDisplay))
  onBeforeUnmount(() => {
    cleanup()
    window.removeEventListener('resize', updateDisplay)
  })

  return {
    getPageSrc,
    getMobilePageSrc,
    isPageInPrefetchWindow,
    onMainImageLoaded,
    showPageSkeleton,
    cleanup,
  }
}
