import { computed, nextTick, onBeforeUnmount, ref, watch, type Ref } from 'vue'

interface ReaderScrollPagesOptions {
  totalPages: Ref<number>
  currentPage: Ref<number>
  scrollReadingForManga: Ref<boolean>
  getVisualPageIndex: (visualPage: number) => number
  getMobilePageSrc: (page: number) => string | undefined
  getPageUrl: (page: number) => string
  commitPage: (page: number) => void
}

export function useReaderScrollPages(options: ReaderScrollPagesOptions) {
  const mobileScrollRef = ref<HTMLElement | null>(null)
  const mobileScrollPageRefs = ref<HTMLElement[]>([])
  let mobileScrollObserver: IntersectionObserver | undefined

  const mobilePages = computed(() =>
    Array.from({ length: options.totalPages.value }, (_, visualIndex) => {
      const page = options.getVisualPageIndex(visualIndex + 1)
      return {
        page,
        src: options.getMobilePageSrc(page),
      }
    }),
  )

  const mobileScrollPages = computed(() =>
    Array.from({ length: options.totalPages.value }, (_, visualIndex) => {
      const page = options.getVisualPageIndex(visualIndex + 1)
      return {
        page,
        src: options.getPageUrl(page),
      }
    }),
  )

  function setMobileScrollPageRef(el: unknown, index: number) {
    if (el instanceof HTMLElement) {
      mobileScrollPageRefs.value[index] = el
    }
  }

  function scrollToCurrentPage() {
    const visualIndex = mobileScrollPages.value.findIndex(
      (mobilePage) => mobilePage.page === options.currentPage.value,
    )
    const el = visualIndex >= 0 ? mobileScrollPageRefs.value[visualIndex] : undefined
    el?.scrollIntoView({ block: 'start' })
  }

  function setupMobileScrollObserver() {
    mobileScrollObserver?.disconnect()
    mobileScrollObserver = undefined

    if (!options.scrollReadingForManga.value || !mobileScrollRef.value) return

    mobileScrollObserver = new IntersectionObserver(
      (entries) => {
        const visibleEntry = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0]
        const page = Number((visibleEntry?.target as HTMLElement | undefined)?.dataset.page)

        if (Number.isFinite(page) && page !== options.currentPage.value) {
          options.commitPage(page)
        }
      },
      {
        root: mobileScrollRef.value,
        threshold: [0.55],
      },
    )

    mobileScrollPageRefs.value.forEach((el) => mobileScrollObserver?.observe(el))
    scrollToCurrentPage()
  }

  watch(
    [options.scrollReadingForManga, mobileScrollPages],
    async () => {
      mobileScrollPageRefs.value = []
      await nextTick()
      setupMobileScrollObserver()
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    mobileScrollObserver?.disconnect()
  })

  return {
    mobileScrollRef,
    mobileScrollPageRefs,
    mobilePages,
    mobileScrollPages,
    setMobileScrollPageRef,
  }
}
