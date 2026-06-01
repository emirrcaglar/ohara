<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useMangaStore } from '../stores/manga'
import { usePreferencesStore } from '../stores/preferences'
import { API_BASE } from '../api/client'

const route = useRoute()
const router = useRouter()
const mangaStore = useMangaStore()
const preferencesStore = usePreferencesStore()
const { rightToLeftSwipeForManga } = storeToRefs(preferencesStore)

const mangaId = computed(() => Number(route.query.manga))
const currentPage = ref(Number(route.query.page) || 0)
const totalPages = ref(Number(route.query.total) || 0)

const pageUrl = computed(() => `${API_BASE}/manga/${mangaId.value}/page/${currentPage.value}`)

const currentVisualPage = computed(() => pageToVisualPage(currentPage.value))
const hasVisualLeft = computed(() => currentVisualPage.value > 0)
const hasVisualRight = computed(() => currentVisualPage.value < totalPages.value - 1)
const visualLeftLabel = computed(() => (rightToLeftSwipeForManga.value ? 'NEXT' : 'PREV'))
const visualRightLabel = computed(() => (rightToLeftSwipeForManga.value ? 'PREV' : 'NEXT'))

const readerRef = ref<HTMLElement | null>(null)
const viewportWidth = ref(0)
const dragX = ref(0)
const isDragging = ref(false)
const isAnimating = ref(false)
const isMomentumPanning = ref(false)
const chromeVisible = ref(false)
const scale = ref(1)
const panX = ref(0)
const panY = ref(0)

let startX = 0
let startY = 0
let lastX = 0
let lastY = 0
let lastMoveAt = 0
let panVelocityX = 0
let panVelocityY = 0
let panMomentumFrame: number | undefined
let pinchStartDistance = 0
let pinchStartScale = 1
let pinchStartPanX = 0
let pinchStartPanY = 0
let pinchStartCenterX = 0
let pinchStartCenterY = 0
let twoFingerTapCandidate = false
let lastTapAt = 0
let tapTimer: number | undefined
const activeTouches = new Map<number, Touch>()

const MIN_SCALE = 1
const MAX_SCALE = 4
const EDGE_GUARD = 16
const DOUBLE_TAP_MS = 280
const SWIPE_THRESHOLD_RATIO = 0.18
const PAN_MOMENTUM_MIN_VELOCITY = 0.08
const PAN_MOMENTUM_FRICTION = 0.94

const visualDragX = computed(() => dragX.value)

const mobileTrackStyle = computed(() => ({
  transform: `translate3d(${-currentVisualPage.value * viewportWidth.value + visualDragX.value}px, 0, 0)`,
  transition: isAnimating.value && !isDragging.value ? 'transform 220ms ease-out' : 'none',
}))

const currentImageStyle = computed(() => ({
  transform: `translate3d(${panX.value}px, ${panY.value}px, 0) scale(${scale.value})`,
}))

const mobileReaderClass = computed(() => (chromeVisible.value ? 'z-20' : 'z-40'))

async function saveProgress() {
  await mangaStore.updateProgress(mangaId.value, currentPage.value)
}

function pageToVisualPage(page: number) {
  if (totalPages.value <= 0) return page
  return rightToLeftSwipeForManga.value ? totalPages.value - 1 - page : page
}

function visualPageToPage(visualPage: number) {
  if (totalPages.value <= 0) return visualPage
  return rightToLeftSwipeForManga.value ? totalPages.value - 1 - visualPage : visualPage
}

function getVisualPageIndex(visualPageNumber: number) {
  return visualPageToPage(visualPageNumber - 1)
}

function goVisualLeft() {
  if (hasVisualLeft.value) {
    commitPage(visualPageToPage(currentVisualPage.value - 1))
  }
}

function goVisualRight() {
  if (hasVisualRight.value) {
    commitPage(visualPageToPage(currentVisualPage.value + 1))
  }
}

function navigate() {
  router.replace({
    query: {
      manga: mangaId.value,
      page: currentPage.value,
      total: totalPages.value,
    },
  })
  void saveProgress()
}

function commitPage(page: number) {
  const next = Math.min(Math.max(page, 0), totalPages.value - 1)
  if (next === currentPage.value) {
    snapToCurrentPage()
    return
  }

  resetZoom()
  currentPage.value = next
  snapToCurrentPage()
  navigate()
}

function snapToCurrentPage() {
  dragX.value = 0
  isAnimating.value = true
  window.setTimeout(() => {
    isAnimating.value = false
  }, 230)
}

const PREFETCH_COUNT = 5
const MOBILE_PREFETCH_COUNT = 2

function shouldLoadMobilePage(page: number) {
  return Math.abs(page - currentPage.value) <= MOBILE_PREFETCH_COUNT
}

function getPageUrl(page: number) {
  return `${API_BASE}/manga/${mangaId.value}/page/${page}`
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
    if (ahead < totalPages.value) {
      const img = new Image()
      img.src = getPageUrl(ahead)
    }

    const behind = currentPage.value - i
    if (behind >= 0) {
      const img = new Image()
      img.src = getPageUrl(behind)
    }
  }
}

watch(currentPage, prefetchPages)

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowLeft' || e.key === 'a') {
    goVisualLeft()
  } else if (e.key === 'ArrowRight' || e.key === 'd') {
    goVisualRight()
  }
}

function updateViewportWidth() {
  viewportWidth.value = readerRef.value?.clientWidth || window.innerWidth
  clampPan()
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function clampPan() {
  if (scale.value <= 1) {
    panX.value = 0
    panY.value = 0
    return { hitX: true, hitY: true }
  }

  const maxPanX = (viewportWidth.value * (scale.value - 1)) / 2
  const maxPanY = (window.innerHeight * (scale.value - 1)) / 2
  const nextPanX = clamp(panX.value, -maxPanX, maxPanX)
  const nextPanY = clamp(panY.value, -maxPanY, maxPanY)
  const hitX = nextPanX !== panX.value
  const hitY = nextPanY !== panY.value
  panX.value = nextPanX
  panY.value = nextPanY
  return { hitX, hitY }
}

function stopPanMomentum() {
  if (panMomentumFrame !== undefined) {
    window.cancelAnimationFrame(panMomentumFrame)
    panMomentumFrame = undefined
  }
  isMomentumPanning.value = false
}

function startPanMomentum() {
  stopPanMomentum()

  if (
    Math.abs(panVelocityX) < PAN_MOMENTUM_MIN_VELOCITY &&
    Math.abs(panVelocityY) < PAN_MOMENTUM_MIN_VELOCITY
  ) {
    return
  }

  let previousTime = performance.now()
  isMomentumPanning.value = true

  function step(now: number) {
    const delta = Math.min(now - previousTime, 32)
    previousTime = now

    panX.value += panVelocityX * delta
    panY.value += panVelocityY * delta
    const { hitX, hitY } = clampPan()

    if (hitX) panVelocityX = 0
    if (hitY) panVelocityY = 0

    panVelocityX *= PAN_MOMENTUM_FRICTION
    panVelocityY *= PAN_MOMENTUM_FRICTION

    if (
      Math.abs(panVelocityX) >= PAN_MOMENTUM_MIN_VELOCITY ||
      Math.abs(panVelocityY) >= PAN_MOMENTUM_MIN_VELOCITY
    ) {
      panMomentumFrame = window.requestAnimationFrame(step)
    } else {
      panMomentumFrame = undefined
      isMomentumPanning.value = false
    }
  }

  panMomentumFrame = window.requestAnimationFrame(step)
}

function resetZoom() {
  stopPanMomentum()
  scale.value = 1
  panX.value = 0
  panY.value = 0
}

function touchDistance(a: Touch, b: Touch) {
  return Math.hypot(a.clientX - b.clientX, a.clientY - b.clientY)
}

function touchCenter(a: Touch, b: Touch) {
  return {
    x: (a.clientX + b.clientX) / 2,
    y: (a.clientY + b.clientY) / 2,
  }
}

function scaleAroundPoint(nextScale: number, pointX: number, pointY: number) {
  const previousScale = scale.value
  const centerX = viewportWidth.value / 2
  const centerY = window.innerHeight / 2
  const ratio = nextScale / previousScale

  panX.value += (pointX - centerX - panX.value) * (1 - ratio)
  panY.value += (pointY - centerY - panY.value) * (1 - ratio)
  scale.value = nextScale
  clampPan()
}

function isEdgeTouch(touch: Touch) {
  return touch.clientX <= EDGE_GUARD || touch.clientX >= viewportWidth.value - EDGE_GUARD
}

function handleTouchStart(event: TouchEvent) {
  updateViewportWidth()
  stopPanMomentum()

  if (activeTouches.size === 0 && event.changedTouches[0] && isEdgeTouch(event.changedTouches[0])) {
    return
  }

  event.preventDefault()

  for (const touch of Array.from(event.changedTouches)) {
    activeTouches.set(touch.identifier, touch)
  }

  isAnimating.value = false

  if (activeTouches.size === 1) {
    const touch = Array.from(activeTouches.values())[0]
    startX = touch.clientX
    startY = touch.clientY
    lastX = touch.clientX
    lastY = touch.clientY
    lastMoveAt = performance.now()
    panVelocityX = 0
    panVelocityY = 0
    dragX.value = 0
    isDragging.value = true
  } else if (activeTouches.size >= 2) {
    const touches = Array.from(activeTouches.values())
    const center = touchCenter(touches[0], touches[1])
    pinchStartDistance = touchDistance(touches[0], touches[1])
    pinchStartScale = scale.value
    pinchStartPanX = panX.value
    pinchStartPanY = panY.value
    pinchStartCenterX = center.x
    pinchStartCenterY = center.y
    twoFingerTapCandidate = true
    panVelocityX = 0
    panVelocityY = 0
  }
}

function handleTouchMove(event: TouchEvent) {
  if (activeTouches.size === 0) return

  event.preventDefault()

  for (const touch of Array.from(event.changedTouches)) {
    if (activeTouches.has(touch.identifier)) {
      activeTouches.set(touch.identifier, touch)
    }
  }

  const touches = Array.from(activeTouches.values())

  if (touches.length >= 2) {
    const distance = touchDistance(touches[0], touches[1])
    const center = touchCenter(touches[0], touches[1])
    const nextScale = clamp(pinchStartScale * (distance / pinchStartDistance), MIN_SCALE, MAX_SCALE)
    const centerX = viewportWidth.value / 2
    const centerY = window.innerHeight / 2
    const ratio = nextScale / pinchStartScale

    if (
      Math.abs(distance - pinchStartDistance) > 8 ||
      Math.abs(center.x - pinchStartCenterX) > 8 ||
      Math.abs(center.y - pinchStartCenterY) > 8
    ) {
      twoFingerTapCandidate = false
    }

    panX.value = pinchStartPanX + (pinchStartCenterX - centerX - pinchStartPanX) * (1 - ratio)
    panY.value = pinchStartPanY + (pinchStartCenterY - centerY - pinchStartPanY) * (1 - ratio)
    scale.value = nextScale
    clampPan()
    return
  }

  const touch = touches[0]
  const dx = touch.clientX - startX
  const dy = touch.clientY - startY
  const moveX = touch.clientX - lastX
  const moveY = touch.clientY - lastY
  const now = performance.now()
  const elapsed = Math.max(now - lastMoveAt, 1)

  if (scale.value > 1) {
    panVelocityX = moveX / elapsed
    panVelocityY = moveY / elapsed
    panX.value += moveX
    panY.value += moveY
    clampPan()
  } else if (Math.abs(dx) > Math.abs(dy)) {
    let nextDragX = dx
    if (
      (currentVisualPage.value === 0 && nextDragX > 0) ||
      (currentVisualPage.value === totalPages.value - 1 && nextDragX < 0)
    ) {
      nextDragX *= 0.28
    }
    dragX.value = nextDragX
  }

  lastX = touch.clientX
  lastY = touch.clientY
  lastMoveAt = now
}

function handleTouchEnd(event: TouchEvent) {
  if (activeTouches.size === 0) return

  for (const touch of Array.from(event.changedTouches)) {
    activeTouches.delete(touch.identifier)
  }

  if (activeTouches.size > 0) {
    const touch = Array.from(activeTouches.values())[0]
    startX = touch.clientX
    startY = touch.clientY
    lastX = touch.clientX
    lastY = touch.clientY
    lastMoveAt = performance.now()
    panVelocityX = 0
    panVelocityY = 0
    return
  }

  isDragging.value = false

  if (twoFingerTapCandidate && scale.value > 1) {
    twoFingerTapCandidate = false
    resetZoom()
    activeTouches.clear()
    return
  }

  twoFingerTapCandidate = false

  const movedX = Math.abs(dragX.value)
  const movedEnough = movedX > viewportWidth.value * SWIPE_THRESHOLD_RATIO
  const committedVisualDragX = visualDragX.value

  if (scale.value <= 1.02) {
    scale.value = 1
    if (movedEnough && committedVisualDragX < 0) {
      commitPage(visualPageToPage(currentVisualPage.value + 1))
    } else if (movedEnough && committedVisualDragX > 0) {
      commitPage(visualPageToPage(currentVisualPage.value - 1))
    } else {
      const touch = event.changedTouches[0]
      const wasTap =
        touch && Math.abs(touch.clientX - startX) < 8 && Math.abs(touch.clientY - startY) < 8
      if (wasTap) handleTap(touch)
      snapToCurrentPage()
    }
  } else {
    clampPan()
    startPanMomentum()
  }

  activeTouches.clear()
}

function handleTap(touch: Touch) {
  const now = performance.now()
  if (now - lastTapAt > DOUBLE_TAP_MS) {
    lastTapAt = now
    window.clearTimeout(tapTimer)
    tapTimer = window.setTimeout(() => {
      chromeVisible.value = !chromeVisible.value
      lastTapAt = 0
    }, DOUBLE_TAP_MS)
    return
  }

  window.clearTimeout(tapTimer)
  lastTapAt = 0

  if (scale.value > 1) {
    resetZoom()
    return
  }

  scaleAroundPoint(2.25, touch.clientX, touch.clientY)
}

onMounted(async () => {
  if (!preferencesStore.hasLoaded) {
    void preferencesStore.loadPreferences()
  }

  window.addEventListener('keydown', handleKeydown)
  window.addEventListener('resize', updateViewportWidth)
  window.addEventListener('orientationchange', updateViewportWidth)
  updateViewportWidth()

  if (totalPages.value === 0 && mangaId.value) {
    const info = await mangaStore.getMangaInfo(mangaId.value)
    if (info) {
      totalPages.value = info.pageCount
    }
  }

  prefetchPages()
})

onBeforeUnmount(() => {
  window.clearTimeout(tapTimer)
  stopPanMomentum()
  window.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('resize', updateViewportWidth)
  window.removeEventListener('orientationchange', updateViewportWidth)
})
</script>

<template>
  <div
    ref="readerRef"
    class="relative flex-1 flex flex-col bg-black overflow-hidden overscroll-none"
  >
    <div
      class="md:hidden fixed inset-0 overflow-hidden bg-black touch-none select-none"
      :class="mobileReaderClass"
      @touchstart="handleTouchStart"
      @touchmove="handleTouchMove"
      @touchend="handleTouchEnd"
      @touchcancel="handleTouchEnd"
    >
      <div class="flex h-dvh will-change-transform" :style="mobileTrackStyle">
        <div
          v-for="visualPage in totalPages"
          :key="getVisualPageIndex(visualPage)"
          class="flex h-dvh w-screen shrink-0 items-center justify-center overflow-hidden bg-black"
        >
          <img
            v-if="getMobilePageSrc(getVisualPageIndex(visualPage))"
            :src="getMobilePageSrc(getVisualPageIndex(visualPage))"
            :alt="`Page ${getVisualPageIndex(visualPage) + 1}`"
            class="h-dvh w-screen object-contain will-change-transform select-none"
            :class="{ 'transition-transform duration-150': !isDragging && !isMomentumPanning }"
            :style="getVisualPageIndex(visualPage) === currentPage ? currentImageStyle : undefined"
            draggable="false"
          />
        </div>
      </div>

      <div
        class="fixed bottom-4 right-4 z-50 rounded bg-black/70 px-2.5 py-1 font-mono text-xs font-bold text-white"
        :style="{ marginBottom: 'env(safe-area-inset-bottom)' }"
      >
        {{ currentPage + 1 }} / {{ totalPages }}
      </div>
    </div>

    <div class="hidden md:block flex-1 min-h-0 p-2 md:p-4">
      <div class="relative w-full h-full">
        <img
          :src="pageUrl"
          :alt="`Page ${currentPage + 1}`"
          class="absolute inset-0 w-full h-full object-contain"
        />
      </div>
    </div>

    <div
      class="fixed right-2 top-16 z-60 hidden items-center bg-surface-container-lowest border-2 border-primary-container p-0.5 shadow-[0_0_40px_rgba(0,0,0,0.8)] sm:right-3 md:right-6 md:top-20 md:flex"
      :style="{ marginTop: 'env(safe-area-inset-top)' }"
    >
      <div class="flex items-center gap-0.5 sm:gap-1">
        <button
          class="flex min-h-9 items-center bg-primary-container px-2 py-1.5 text-[9px] font-black uppercase tracking-[0.12em] text-on-primary-container transition-all hover:bg-primary disabled:opacity-30 disabled:cursor-not-allowed sm:min-h-10 sm:px-3 sm:text-[10px] md:px-4 md:text-xs md:tracking-[0.16em]"
          :disabled="!hasVisualLeft"
          @click="goVisualLeft"
        >
          <span class="material-symbols-outlined mr-0.5 text-[13px] sm:mr-1 sm:text-sm"
            >chevron_left</span
          >
          {{ visualLeftLabel }}
        </button>

        <div
          class="flex min-h-9 items-center gap-1.5 bg-surface px-2 py-1.5 font-mono sm:min-h-10 sm:gap-2 sm:px-3 md:gap-3 md:px-4"
        >
          <span
            class="text-[8px] font-bold uppercase tracking-widest text-outline sm:text-[9px] md:text-[10px]"
          >
            PAGE
          </span>
          <span class="text-[11px] font-bold tracking-tighter text-primary sm:text-xs">
            {{ currentPage + 1 }}
          </span>
          <span class="text-outline-variant">/</span>
          <span class="text-[11px] font-bold tracking-tighter text-on-surface sm:text-xs">{{
            totalPages
          }}</span>
        </div>

        <button
          class="flex min-h-9 items-center bg-primary-container px-2 py-1.5 text-[9px] font-black uppercase tracking-[0.12em] text-on-primary-container transition-all hover:bg-primary disabled:opacity-30 disabled:cursor-not-allowed sm:min-h-10 sm:px-3 sm:text-[10px] md:px-4 md:text-xs md:tracking-[0.16em]"
          :disabled="!hasVisualRight"
          @click="goVisualRight"
        >
          {{ visualRightLabel }}
          <span class="material-symbols-outlined ml-0.5 text-[13px] sm:ml-1 sm:text-sm"
            >chevron_right</span
          >
        </button>
      </div>
    </div>
  </div>
</template>
