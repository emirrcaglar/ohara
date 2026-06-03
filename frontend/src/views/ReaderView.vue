<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '../stores/preferences'
import { API_BASE } from '../api/client'
import { useReaderSetup } from '../composables/useReaderSetup'
import { useDesktopReaderZoom } from '../composables/useDesktopReaderZoom'
import { useReaderScrollPages } from '../composables/useReaderScrollPages'

const route = useRoute()
const router = useRouter()
const preferencesStore = usePreferencesStore()
const { rightToLeftSwipeForManga, scrollReadingForManga } = storeToRefs(preferencesStore)

const mangaId = computed(() => Number(route.query.manga))
const totalPages = ref(Number(route.query.total) || 0)

const pageUrl = computed(() => `${API_BASE}/manga/${mangaId.value}/page/${currentPage.value}`)

function getPageUrl(page: number) {
  return `${API_BASE}/manga/${mangaId.value}/page/${page}`
}

const {
  readerRef,
  currentPage,
  hasVisualLeft,
  hasVisualRight,
  visualLeftLabel,
  visualRightLabel,
  getVisualPageIndex,
  goVisualLeft,
  goVisualRight,
  commitPage,
  getMobilePageSrc,
  showPageSkeleton,
  onMainImageLoaded,
  currentImageStyle,
  snapAfterPageChange,
  handleTouchStart,
  handleTouchMove,
  handleTouchEnd,
  mobileTrackStyle,
  mobileReaderClass,
  chromeVisible,
} = useReaderSetup({
  mangaId,
  totalPages,
  initialPage: Number(route.query.page) || 0,
  rightToLeftSwipe: rightToLeftSwipeForManga,
  router,
  getPageUrl,
})

const {
  mobileScrollRef,
  mobileScrollPageRefs,
  mobilePages,
  mobileScrollPages,
  setMobileScrollPageRef,
} = useReaderScrollPages({
  totalPages,
  currentPage,
  scrollReadingForManga,
  getVisualPageIndex,
  getMobilePageSrc,
  getPageUrl,
  commitPage,
})

function isDesktopPointer() {
  return window.matchMedia('(min-width: 768px)').matches
}

function toggleMobileChrome() {
  if (isDesktopPointer()) return
  chromeVisible.value = !chromeVisible.value
}

const {
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
} = useDesktopReaderZoom({
  mobileScrollRef,
  mobileScrollPageRefs,
  mobileScrollPages,
})
</script>

<template>
  <div
    ref="readerRef"
    class="relative flex-1 flex flex-col bg-black overflow-hidden overscroll-none"
  >
    <div
      v-if="!scrollReadingForManga"
      class="md:hidden fixed inset-0 overflow-hidden bg-black touch-none select-none"
      :class="mobileReaderClass"
      @touchstart="handleTouchStart"
      @touchmove="handleTouchMove"
      @touchend="handleTouchEnd"
      @touchcancel="handleTouchEnd"
    >
      <div class="flex h-dvh will-change-transform" :style="mobileTrackStyle">
        <div
          v-for="mobilePage in mobilePages"
          :key="mobilePage.page"
          class="flex h-dvh w-screen shrink-0 items-center justify-center overflow-hidden bg-black"
        >
          <img
            v-if="mobilePage.src"
            :src="mobilePage.src"
            :alt="`Page ${mobilePage.page + 1}`"
            class="h-dvh w-screen object-contain will-change-transform select-none"
            :class="mobilePage.page === currentPage ? 'transition-transform duration-150' : ''"
            :style="mobilePage.page === currentPage ? currentImageStyle : undefined"
            draggable="false"
            @load="onMainImageLoaded(mobilePage.page)"
          />
        </div>
      </div>

      <div
        v-if="showPageSkeleton"
        class="pointer-events-none fixed inset-0 z-40 flex items-center justify-center bg-black"
      >
        <div class="h-dvh w-screen animate-pulse bg-surface-container-lowest/40" />
      </div>

      <div
        class="fixed bottom-4 right-4 z-50 rounded bg-black/70 px-2.5 py-1 font-mono text-xs font-bold text-white"
        :style="{ marginBottom: 'env(safe-area-inset-bottom)' }"
      >
        {{ currentPage + 1 }} / {{ totalPages }}
      </div>
    </div>

    <!-- Scroll read mode (mobile + desktop) -->
    <div
      v-else
      ref="mobileScrollRef"
      class="fixed inset-0 overflow-y-auto bg-black overscroll-contain select-none md:static md:flex-1 md:min-h-0"
      :class="mobileReaderClass"
      @click="toggleMobileChrome"
      @mousemove="handleScrollAreaMouseMove"
      @mouseleave="handleScrollAreaMouseLeave"
    >
      <div
        v-for="(mobilePage, index) in mobileScrollPages"
        :key="mobilePage.page"
        :ref="(el) => setMobileScrollPageRef(el, index)"
        :data-page="mobilePage.page"
        class="flex min-h-dvh w-full items-center justify-center bg-black"
        :class="getScrollPageClass(mobilePage.page)"
        @mousemove="updateZoomOriginForPage(mobilePage.page, $event)"
        @mouseleave="onScrollPageMouseLeave(mobilePage.page)"
      >
        <img
          :src="mobilePage.src"
          :alt="`Page ${mobilePage.page + 1}`"
          class="min-h-0 w-full max-w-full object-contain select-none transition-transform duration-150 md:h-auto md:max-h-dvh md:w-auto md:cursor-zoom-in"
          :class="[
            desktopZoomEnabled ? 'md:cursor-zoom-out' : '',
            activeDesktopZoomPage === mobilePage.page
              ? 'md:relative md:z-60'
              : desktopZoomedPages.includes(mobilePage.page)
                ? 'md:relative md:z-50'
                : '',
          ]"
          :style="getDesktopImageStyle(mobilePage.page)"
          loading="lazy"
          draggable="false"
          @mouseenter="handleScrollPageImageEnter(mobilePage.page, $event)"
          @click="toggleDesktopZoom(mobilePage.page, $event)"
          @mousemove="updateZoomOriginForPage(mobilePage.page, $event)"
          @load="onMainImageLoaded(mobilePage.page)"
        />
      </div>

      <div
        class="fixed bottom-4 right-4 z-50 rounded bg-black/70 px-2.5 py-1 font-mono text-xs font-bold text-white"
        :style="{ marginBottom: 'env(safe-area-inset-bottom)' }"
      >
        {{ currentPage + 1 }} / {{ totalPages }}
      </div>
    </div>

    <div v-if="!scrollReadingForManga" class="hidden md:block flex-1 min-h-0 p-2 md:p-4">
      <div class="relative w-full h-full" @mouseleave="resetDesktopZoom">
        <img
          :key="currentPage"
          :src="pageUrl"
          :alt="`Page ${currentPage + 1}`"
          class="absolute inset-0 w-full h-full object-contain transition-transform duration-150 cursor-zoom-in"
          :class="desktopZoomedPages.includes(currentPage) ? 'cursor-zoom-out' : ''"
          :style="getDesktopImageStyle(currentPage)"
          @click="toggleDesktopZoom(currentPage, $event)"
          @mousemove="updateZoomOriginForPage(currentPage, $event)"
          @load="onMainImageLoaded(currentPage)"
        />
        <div
          v-if="showPageSkeleton"
          class="pointer-events-none absolute inset-0 flex items-center justify-center bg-black"
        >
          <div class="h-full w-full animate-pulse bg-surface-container-lowest/40" />
        </div>
      </div>
    </div>

    <div
      v-if="!scrollReadingForManga"
      class="fixed right-2 top-16 z-60 hidden items-center bg-surface-container-lowest border-2 border-primary-container p-0.5 shadow-[0_0_40px_rgba(0,0,0,0.8)] sm:right-3 md:right-6 md:top-20 md:flex"
      :style="{ marginTop: 'env(safe-area-inset-top)' }"
    >
      <div class="flex items-center gap-0.5 sm:gap-1">
        <button
          class="flex min-h-9 items-center bg-primary-container px-2 py-1.5 text-[9px] font-black uppercase tracking-[0.12em] text-on-primary-container transition-all hover:bg-primary disabled:opacity-30 disabled:cursor-not-allowed sm:min-h-10 sm:px-3 sm:text-[10px] md:px-4 md:text-xs md:tracking-[0.16em]"
          :disabled="!hasVisualLeft"
          @click="goVisualLeft(snapAfterPageChange)"
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
          @click="goVisualRight(snapAfterPageChange)"
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
