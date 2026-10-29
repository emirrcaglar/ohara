import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

export function useReaderPageLoading(currentPage: Ref<number>, getPageUrl: (page: number) => string) {
  const showPageSkeleton = ref(false)
  const isMobile = ref(window.innerWidth < 768)
  let skeletonTimer: ReturnType<typeof setTimeout> | undefined

  function updateDisplay() {
    isMobile.value = window.innerWidth < 768
  }

  function getMobilePageSrc(page: number) {
    return isMobile.value && Math.abs(page - currentPage.value) <= 1 ? getPageUrl(page) : undefined
  }

  function onMainImageLoaded(page = currentPage.value) {
    if (page !== currentPage.value) return
    clearTimeout(skeletonTimer)
    showPageSkeleton.value = false
  }

  function cleanup() {
    clearTimeout(skeletonTimer)
  }

  watch(() => getPageUrl(currentPage.value), () => {
    cleanup()
    showPageSkeleton.value = false
    skeletonTimer = setTimeout(() => { showPageSkeleton.value = true }, 50)
  })

  onMounted(() => window.addEventListener('resize', updateDisplay))
  onBeforeUnmount(() => {
    cleanup()
    window.removeEventListener('resize', updateDisplay)
  })

  return { getMobilePageSrc, onMainImageLoaded, showPageSkeleton, cleanup }
}
