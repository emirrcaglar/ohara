import { ref, computed, onMounted, onBeforeUnmount, type Ref } from 'vue'
import type { Router } from 'vue-router'
import { useMangaStore } from '../stores/manga'
import { usePreferencesStore } from '../stores/preferences'
import { usePageNavigation } from './usePageNavigation'
import { usePagePrefetch } from './usePagePrefetch'
import { useZoomAndPan } from './useZoomAndPan'
import { useTouchGestures } from './useTouchGestures'

interface ReaderSetupOptions {
  mangaId: Ref<number>
  totalPages: Ref<number>
  initialPage: number
  rightToLeftSwipe: Ref<boolean>
  router: Router
  getPageUrl: (page: number) => string
}

export function useReaderSetup(options: ReaderSetupOptions) {
  const mangaStore = useMangaStore()
  const preferencesStore = usePreferencesStore()

  const readerRef = ref<HTMLElement | null>(null)
  const viewportWidth = ref(0)
  const chromeVisible = ref(false)

  const navigation = usePageNavigation(
    options.mangaId,
    options.initialPage,
    options.totalPages,
    options.rightToLeftSwipe,
    options.router,
  )

  const {
    currentPage,
    currentVisualPage,
    hasVisualLeft,
    hasVisualRight,
    visualLeftLabel,
    visualRightLabel,
    visualPageToPage,
    getVisualPageIndex,
    commitPage,
    goVisualLeft,
    goVisualRight,
  } = navigation

  const prefetch = usePagePrefetch(
    options.mangaId,
    currentPage,
    options.totalPages,
    options.getPageUrl,
  )

  const { getMobilePageSrc, prefetchPages, onMainImageLoaded } = prefetch

  const zoom = useZoomAndPan(viewportWidth)

  const {
    scale,
    panX,
    panY,
    currentImageStyle,
    resetZoom,
    scaleAroundPoint,
    stopPanMomentum,
    startPanMomentum,
    clampPan,
    setPanVelocity,
    updatePan,
    setPan,
    setScale,
    clamp,
    MIN_SCALE,
    MAX_SCALE,
  } = zoom

  function snapAfterPageChange() {
    resetZoom()
    gestures.snapToCurrentPage()
  }

  const gestures = useTouchGestures({
    viewportWidth,
    currentVisualPage,
    totalPages: options.totalPages,
    scale,
    panX,
    panY,
    onPageSwipe: (direction) => {
      commitPage(visualPageToPage(currentVisualPage.value + direction), snapAfterPageChange)
    },
    onTap: () => {
      chromeVisible.value = !chromeVisible.value
    },
    resetZoom,
    scaleAroundPoint,
    stopPanMomentum,
    startPanMomentum,
    clampPan,
    setPanVelocity,
    updatePan,
    setPan,
    setScale,
    clamp,
    MIN_SCALE,
    MAX_SCALE,
  })

  const { dragX, isDragging, isAnimating, handleTouchStart, handleTouchMove, handleTouchEnd } =
    gestures

  const mobileTrackStyle = computed(() => ({
    transform: `translate3d(${-currentVisualPage.value * viewportWidth.value + dragX.value}px, 0, 0)`,
    transition: isAnimating.value && !isDragging.value ? 'transform 220ms ease-out' : 'none',
  }))

  const mobileReaderClass = computed(() => (chromeVisible.value ? 'z-20' : 'z-40'))

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowLeft' || e.key === 'a') {
      goVisualLeft(snapAfterPageChange)
    } else if (e.key === 'ArrowRight' || e.key === 'd') {
      goVisualRight(snapAfterPageChange)
    }
  }

  function updateViewportWidth() {
    viewportWidth.value = readerRef.value?.clientWidth || window.innerWidth
    zoom.clampPan()
  }

  onMounted(async () => {
    if (!preferencesStore.hasLoaded) {
      void preferencesStore.loadPreferences()
    }

    window.addEventListener('keydown', handleKeydown)
    window.addEventListener('resize', updateViewportWidth)
    window.addEventListener('orientationchange', updateViewportWidth)
    updateViewportWidth()

    if (options.totalPages.value === 0 && options.mangaId.value) {
      const info = await mangaStore.getMangaInfo(options.mangaId.value)
      if (info) {
        options.totalPages.value = info.pageCount
      }
    }

    prefetchPages()
  })

  onBeforeUnmount(() => {
    navigation.cleanup()
    prefetch.cleanup()
    gestures.cleanup()
    window.removeEventListener('keydown', handleKeydown)
    window.removeEventListener('resize', updateViewportWidth)
    window.removeEventListener('orientationchange', updateViewportWidth)
  })

  return {
    readerRef,
    viewportWidth,
    chromeVisible,
    currentPage,
    currentVisualPage,
    hasVisualLeft,
    hasVisualRight,
    visualLeftLabel,
    visualRightLabel,
    visualPageToPage,
    getVisualPageIndex,
    commitPage,
    goVisualLeft,
    goVisualRight,
    getMobilePageSrc,
    onMainImageLoaded,
    scale,
    currentImageStyle,
    resetZoom,
    snapAfterPageChange,
    dragX,
    isDragging,
    isAnimating,
    handleTouchStart,
    handleTouchMove,
    handleTouchEnd,
    mobileTrackStyle,
    mobileReaderClass,
    gestures,
  }
}
