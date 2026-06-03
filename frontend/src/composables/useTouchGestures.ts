import { ref, type Ref } from 'vue'

const EDGE_GUARD = 16
const DOUBLE_TAP_MS = 280
const SWIPE_THRESHOLD_RATIO = 0.18

interface TouchGestureOptions {
  viewportWidth: Ref<number>
  currentVisualPage: Ref<number>
  totalPages: Ref<number>
  scale: Ref<number>
  panX: Ref<number>
  panY: Ref<number>
  onPageSwipe: (direction: 1 | -1) => void
  onTap: () => void
  resetZoom: () => void
  scaleAroundPoint: (scale: number, x: number, y: number) => void
  stopPanMomentum: () => void
  startPanMomentum: () => void
  clampPan: () => { hitX: boolean; hitY: boolean }
  setPanVelocity: (vx: number, vy: number) => void
  updatePan: (dx: number, dy: number) => void
  setPan: (x: number, y: number) => void
  setScale: (scale: number) => void
  clamp: (value: number, min: number, max: number) => number
  MIN_SCALE: number
  MAX_SCALE: number
}

export function useTouchGestures(options: TouchGestureOptions) {
  const dragX = ref(0)
  const isDragging = ref(false)
  const isAnimating = ref(false)

  let startX = 0
  let startY = 0
  let lastX = 0
  let lastY = 0
  let lastMoveAt = 0
  let pinchStartDistance = 0
  let pinchStartScale = 1
  let pinchStartPanX = 0
  let pinchStartPanY = 0
  let pinchStartCenterX = 0
  let pinchStartCenterY = 0
  let lastTapAt = 0
  let tapTimer: number | undefined
  let snapTimer: number | undefined
  const activeTouches = new Map<number, Touch>()

  function snapToCurrentPage() {
    window.clearTimeout(snapTimer)
    dragX.value = 0
    isAnimating.value = true
    snapTimer = window.setTimeout(() => {
      isAnimating.value = false
      snapTimer = undefined
    }, 230)
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

  function isEdgeTouch(touch: Touch) {
    return touch.clientX <= EDGE_GUARD || touch.clientX >= options.viewportWidth.value - EDGE_GUARD
  }

  function handleTouchStart(event: TouchEvent) {
    options.stopPanMomentum()

    if (
      activeTouches.size === 0 &&
      event.changedTouches[0] &&
      isEdgeTouch(event.changedTouches[0])
    ) {
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
      options.setPanVelocity(0, 0)
      dragX.value = 0
      isDragging.value = true
    } else if (activeTouches.size >= 2) {
      const touches = Array.from(activeTouches.values())
      const center = touchCenter(touches[0], touches[1])
      pinchStartDistance = touchDistance(touches[0], touches[1])
      pinchStartScale = options.scale.value
      pinchStartPanX = options.panX.value
      pinchStartPanY = options.panY.value
      pinchStartCenterX = center.x
      pinchStartCenterY = center.y
      options.setPanVelocity(0, 0)
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
      const nextScale = options.clamp(
        pinchStartScale * (distance / pinchStartDistance),
        options.MIN_SCALE,
        options.MAX_SCALE,
      )
      const centerX = options.viewportWidth.value / 2
      const centerY = window.innerHeight / 2
      const ratio = nextScale / pinchStartScale

      const panX = pinchStartPanX + (pinchStartCenterX - centerX - pinchStartPanX) * (1 - ratio)
      const panY = pinchStartPanY + (pinchStartCenterY - centerY - pinchStartPanY) * (1 - ratio)
      options.setScale(nextScale)
      options.setPan(panX, panY)
      return
    }

    const touch = touches[0]
    const dx = touch.clientX - startX
    const dy = touch.clientY - startY
    const moveX = touch.clientX - lastX
    const moveY = touch.clientY - lastY
    const now = performance.now()
    const elapsed = Math.max(now - lastMoveAt, 1)

    if (options.scale.value > 1) {
      options.setPanVelocity(moveX / elapsed, moveY / elapsed)
      options.updatePan(moveX, moveY)
    } else if (Math.abs(dx) > Math.abs(dy)) {
      let nextDragX = dx
      if (
        (options.currentVisualPage.value === 0 && nextDragX > 0) ||
        (options.currentVisualPage.value === options.totalPages.value - 1 && nextDragX < 0)
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
      options.setPanVelocity(0, 0)
      return
    }

    isDragging.value = false

    const touch = event.changedTouches[0]
    const wasTap =
      touch && Math.abs(touch.clientX - startX) < 8 && Math.abs(touch.clientY - startY) < 8

    if (wasTap) {
      handleTap(touch)
      snapToCurrentPage()
      activeTouches.clear()
      return
    }

    const movedX = Math.abs(dragX.value)
    const movedEnough = movedX > options.viewportWidth.value * SWIPE_THRESHOLD_RATIO

    if (options.scale.value <= 1.02) {
      options.setScale(1)
      if (movedEnough && dragX.value < 0) {
        options.onPageSwipe(1)
      } else if (movedEnough && dragX.value > 0) {
        options.onPageSwipe(-1)
      } else {
        snapToCurrentPage()
      }
    } else {
      options.clampPan()
      options.startPanMomentum()
    }

    activeTouches.clear()
  }

  function handleTap(touch: Touch) {
    const now = performance.now()
    if (now - lastTapAt > DOUBLE_TAP_MS) {
      lastTapAt = now
      window.clearTimeout(tapTimer)
      tapTimer = window.setTimeout(() => {
        options.onTap()
        lastTapAt = 0
      }, DOUBLE_TAP_MS)
      return
    }

    window.clearTimeout(tapTimer)
    lastTapAt = 0

    if (options.scale.value > 1) {
      options.resetZoom()
      return
    }

    options.scaleAroundPoint(2.25, touch.clientX, touch.clientY)
  }

  function cleanup() {
    window.clearTimeout(tapTimer)
    window.clearTimeout(snapTimer)
  }

  return {
    dragX,
    isDragging,
    isAnimating,
    snapToCurrentPage,
    handleTouchStart,
    handleTouchMove,
    handleTouchEnd,
    cleanup,
  }
}
