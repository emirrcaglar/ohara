import { ref, computed, type Ref } from 'vue'

const MIN_SCALE = 1
const MAX_SCALE = 4
const PAN_MOMENTUM_MIN_VELOCITY = 0.08
const PAN_MOMENTUM_FRICTION = 0.94

export function useZoomAndPan(viewportWidth: Ref<number>) {
  const scale = ref(1)
  const panX = ref(0)
  const panY = ref(0)
  const isMomentumPanning = ref(false)

  let panVelocityX = 0
  let panVelocityY = 0
  let panMomentumFrame: number | undefined

  const currentImageStyle = computed(() => ({
    transform: `translate3d(${panX.value}px, ${panY.value}px, 0) scale(${scale.value})`,
  }))

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

  function setPanVelocity(vx: number, vy: number) {
    panVelocityX = vx
    panVelocityY = vy
  }

  function updatePan(dx: number, dy: number) {
    panX.value += dx
    panY.value += dy
    clampPan()
  }

  function setPan(nextPanX: number, nextPanY: number) {
    panX.value = nextPanX
    panY.value = nextPanY
    clampPan()
  }

  function setScale(newScale: number) {
    scale.value = clamp(newScale, MIN_SCALE, MAX_SCALE)
  }

  return {
    scale,
    panX,
    panY,
    isMomentumPanning,
    currentImageStyle,
    MIN_SCALE,
    MAX_SCALE,
    clamp,
    clampPan,
    stopPanMomentum,
    startPanMomentum,
    resetZoom,
    scaleAroundPoint,
    setPanVelocity,
    updatePan,
    setPan,
    setScale,
  }
}
