import { ref, onBeforeUnmount, type Ref } from 'vue'

interface ReaderScrollPage {
  page: number
  src: string
}

interface ReaderScrollPagesRef {
  value: ReaderScrollPage[]
}

interface DesktopReaderZoomOptions {
  mobileScrollRef: Ref<HTMLElement | null>
  mobileScrollPageRefs: Ref<HTMLElement[]>
  mobileScrollPages: ReaderScrollPagesRef
}

const ZOOM_HYSTERESIS_PX = 150
const AUTO_SCROLL_ZONE = 0.2 // fraction of viewport height per zone
const AUTO_SCROLL_MAX_SPEED = 15 // px per frame at max depth (≈900 px/s @ 60 fps)

interface PendingUnzoom {
  page: number
  startY: number
}

export function useDesktopReaderZoom(options: DesktopReaderZoomOptions) {
  const desktopZoomedPages = ref<number[]>([])
  const desktopZoomOrigins = ref<Record<number, string>>({})
  const desktopZoomEnabled = ref(false)
  const activeDesktopZoomPage = ref<number | null>(null)

  let pendingUnzoom: PendingUnzoom | null = null
  let pageLeaveTimer: ReturnType<typeof setTimeout> | undefined
  let lastDesktopCursor: { clientX: number; clientY: number } | null = null
  let autoScrollSpeed = 0
  let autoScrollFrame: number | undefined

  function isDesktopPointer() {
    return window.matchMedia('(min-width: 768px)').matches
  }

  function computeOrigin(
    target: HTMLElement,
    clientX: number,
    clientY: number,
    clamp = true,
  ): string {
    const rect = target.getBoundingClientRect()
    const rawX = ((clientX - rect.left) / rect.width) * 100
    const rawY = ((clientY - rect.top) / rect.height) * 100
    const x = clamp ? Math.min(100, Math.max(0, rawX)) : rawX
    const y = clamp ? Math.min(100, Math.max(0, rawY)) : rawY
    return `${x}% ${y}%`
  }

  function getDesktopImageStyle(page: number) {
    if (!desktopZoomedPages.value.includes(page)) return undefined
    return {
      transform: 'scale(2.25)',
      transformOrigin: desktopZoomOrigins.value[page] ?? '50% 50%',
    }
  }

  function getScrollPageClass(page: number) {
    if (activeDesktopZoomPage.value === page) return 'md:relative md:z-50 md:overflow-visible'
    if (desktopZoomedPages.value.includes(page)) return 'md:relative md:z-40 md:overflow-visible'
    return 'md:relative md:z-0 md:overflow-hidden'
  }

  function getScrollPageImage(page: number) {
    const visualIndex = options.mobileScrollPages.value.findIndex((p) => p.page === page)
    return visualIndex >= 0
      ? options.mobileScrollPageRefs.value[visualIndex]?.querySelector<HTMLElement>('img')
      : null
  }

  function applyUnifiedZoomOrigin(clientX: number, clientY: number) {
    desktopZoomOrigins.value = Object.fromEntries(
      desktopZoomedPages.value.flatMap((page) => {
        const img = getScrollPageImage(page)
        return img ? [[page, computeOrigin(img, clientX, clientY, false)]] : []
      }),
    )
  }

  function clearZoomState() {
    desktopZoomedPages.value = []
    desktopZoomOrigins.value = {}
    activeDesktopZoomPage.value = null
    pendingUnzoom = null
  }

  function toggleDesktopZoom(page: number, event: MouseEvent) {
    if (!isDesktopPointer()) return
    if (desktopZoomEnabled.value) {
      clearZoomState()
      desktopZoomEnabled.value = false
    } else {
      const target = event.currentTarget instanceof HTMLElement ? event.currentTarget : null
      const origin = target ? computeOrigin(target, event.clientX, event.clientY) : '50% 50%'
      desktopZoomedPages.value = [page]
      desktopZoomOrigins.value = { [page]: origin }
      desktopZoomEnabled.value = true
      activeDesktopZoomPage.value = page
    }
  }

  function resetDesktopZoom() {
    clearZoomState()
    desktopZoomEnabled.value = false
  }

  function updateZoomOriginForPage(page: number, event: MouseEvent) {
    if (!desktopZoomedPages.value.includes(page)) return

    applyUnifiedZoomOrigin(event.clientX, event.clientY)
  }

  function cancelPageLeaveTimer() {
    if (pageLeaveTimer !== undefined) {
      clearTimeout(pageLeaveTimer)
      pageLeaveTimer = undefined
    }
  }

  function removePage(page: number) {
    desktopZoomedPages.value = desktopZoomedPages.value.filter((p) => p !== page)
    if (activeDesktopZoomPage.value === page) {
      activeDesktopZoomPage.value = desktopZoomedPages.value.at(-1) ?? null
    }
    const next = { ...desktopZoomOrigins.value }
    delete next[page]
    desktopZoomOrigins.value = next
  }

  function onScrollPageMouseLeave(page: number) {
    cancelPageLeaveTimer()
    pageLeaveTimer = setTimeout(() => {
      pageLeaveTimer = undefined
      removePage(page)
      if (pendingUnzoom?.page === page) pendingUnzoom = null
    }, 100)
  }

  function getScrollPageAtPoint(clientY: number) {
    const visualIndex = options.mobileScrollPageRefs.value.findIndex((el) => {
      const rect = el.getBoundingClientRect()
      return clientY >= rect.top && clientY <= rect.bottom
    })

    return visualIndex >= 0 ? options.mobileScrollPages.value[visualIndex]?.page : undefined
  }

  function activateScrollZoomPage(page: number, clientX: number, clientY: number) {
    if (!desktopZoomEnabled.value) return
    cancelPageLeaveTimer()

    const prevPage = desktopZoomedPages.value.find((p) => p !== page)
    if (prevPage !== undefined) {
      pendingUnzoom = {
        page: prevPage,
        startY: clientY,
      }
    }

    const existing = desktopZoomedPages.value.filter((p) => p !== page)
    desktopZoomedPages.value = [...existing, page].slice(-2)
    activeDesktopZoomPage.value = page
    applyUnifiedZoomOrigin(clientX, clientY)
  }

  function syncScrollZoomPageAtCursor() {
    if (!lastDesktopCursor || !desktopZoomEnabled.value) return

    const page = getScrollPageAtPoint(lastDesktopCursor.clientY)
    if (page === undefined) return

    activateScrollZoomPage(page, lastDesktopCursor.clientX, lastDesktopCursor.clientY)
  }

  function handleScrollPageImageEnter(page: number, event: MouseEvent) {
    if (!isDesktopPointer()) return
    lastDesktopCursor = { clientX: event.clientX, clientY: event.clientY }
    activateScrollZoomPage(page, event.clientX, event.clientY)
  }

  function stopAutoScroll() {
    autoScrollSpeed = 0
    if (autoScrollFrame !== undefined) {
      cancelAnimationFrame(autoScrollFrame)
      autoScrollFrame = undefined
    }
  }

  function doAutoScroll() {
    if (!options.mobileScrollRef.value || autoScrollSpeed === 0) {
      autoScrollFrame = undefined
      return
    }
    options.mobileScrollRef.value.scrollTop += autoScrollSpeed
    syncScrollZoomPageAtCursor()
    autoScrollFrame = requestAnimationFrame(doAutoScroll)
  }

  function handleScrollAreaMouseMove(event: MouseEvent) {
    if (!isDesktopPointer()) return
    lastDesktopCursor = { clientX: event.clientX, clientY: event.clientY }

    if (pendingUnzoom !== null) {
      const { page, startY } = pendingUnzoom
      if (Math.abs(event.clientY - startY) >= ZOOM_HYSTERESIS_PX) {
        removePage(page)
        pendingUnzoom = null
      }
    }

    const zoneHeight = window.innerHeight * AUTO_SCROLL_ZONE
    const bottomThreshold = window.innerHeight - zoneHeight
    const topThreshold = zoneHeight

    if (event.clientY > bottomThreshold) {
      const t = (event.clientY - bottomThreshold) / zoneHeight
      autoScrollSpeed = t * t * AUTO_SCROLL_MAX_SPEED
    } else if (event.clientY < topThreshold) {
      const t = (topThreshold - event.clientY) / zoneHeight
      autoScrollSpeed = -(t * t * AUTO_SCROLL_MAX_SPEED)
    } else {
      autoScrollSpeed = 0
    }

    if (autoScrollSpeed !== 0 && autoScrollFrame === undefined) {
      autoScrollFrame = requestAnimationFrame(doAutoScroll)
    }
  }

  function handleScrollAreaMouseLeave() {
    stopAutoScroll()
    cancelPageLeaveTimer()
    lastDesktopCursor = null
    clearZoomState()
  }

  onBeforeUnmount(() => {
    stopAutoScroll()
    cancelPageLeaveTimer()
  })

  return {
    desktopZoomedPages,
    desktopZoomEnabled,
    activeDesktopZoomPage,
    getDesktopImageStyle,
    getScrollPageClass,
    toggleDesktopZoom,
    resetDesktopZoom,
    updateZoomOriginForPage,
    onScrollPageMouseLeave,
    handleScrollPageImageEnter,
    handleScrollAreaMouseMove,
    handleScrollAreaMouseLeave,
  }
}
