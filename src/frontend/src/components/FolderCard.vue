<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Folder, MangaRow } from '../types/api'
import { getMangaCoverUrl } from '../api/manga'

const props = defineProps<{ folder: Folder; coverChapter?: MangaRow }>()
interface MenuAnchor {
  top: number
  right: number
}
const emit = defineEmits<{
  open: [folder: Folder]
  rename: [folder: Folder]
  move: [folder: Folder]
  delete: [folder: Folder, anchor: MenuAnchor]
}>()
const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)
const menuButton = ref<HTMLButtonElement | null>(null)
const coverFailed = ref(false)
const coverUrl = computed(() =>
  props.coverChapter ? getMangaCoverUrl(props.coverChapter.id) : null,
)
const hasReadingHistory = computed(() =>
  props.coverChapter?.updatedAt ? Number.isFinite(Date.parse(props.coverChapter.updatedAt)) : false,
)
watch(coverUrl, () => {
  coverFailed.value = false
})

function closeOnOutside(event: PointerEvent) {
  if (!menuRef.value?.contains(event.target as Node)) menuOpen.value = false
}
function closeOnEscape(event: KeyboardEvent) {
  if (event.key === 'Escape' && menuOpen.value) {
    menuOpen.value = false
    menuButton.value?.focus()
  }
}
onMounted(() => {
  window.addEventListener('pointerdown', closeOnOutside)
  window.addEventListener('keydown', closeOnEscape)
})
onBeforeUnmount(() => {
  window.removeEventListener('pointerdown', closeOnOutside)
  window.removeEventListener('keydown', closeOnEscape)
})
function act(action: 'rename' | 'move' | 'delete') {
  menuOpen.value = false
  if (action === 'rename') emit('rename', props.folder)
  else if (action === 'move') emit('move', props.folder)
  else {
    const rect = menuButton.value!.getBoundingClientRect()
    emit('delete', props.folder, { top: rect.bottom, right: window.innerWidth - rect.right })
  }
}
</script>

<template>
  <article class="group relative flex h-full flex-col pt-2">
    <span
      class="absolute inset-x-3 top-0 h-1 bg-surface-container-highest"
      aria-hidden="true"
    ></span>
    <span
      class="absolute inset-x-1.5 top-1 h-1 bg-surface-container-high"
      aria-hidden="true"
    ></span>
    <button
      type="button"
      class="flex h-full w-full cursor-pointer flex-col bg-surface-container-low text-left transition-colors hover:bg-surface-container-high focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      :aria-label="`Open folder ${folder.name}`"
      @click="emit('open', folder)"
    >
      <div
        class="relative flex aspect-3/4 w-full items-center justify-center overflow-hidden bg-surface-container-lowest"
      >
        <img
          v-if="coverUrl && !coverFailed"
          :src="coverUrl"
          alt=""
          class="absolute inset-0 h-full w-full object-cover opacity-80 grayscale transition-[filter,opacity] group-hover:opacity-100 group-hover:grayscale-0"
          loading="lazy"
          decoding="async"
          @error="coverFailed = true"
        />
        <div
          v-else
          class="flex flex-col items-center gap-3 px-3 text-center text-on-surface-variant"
        >
          <span class="text-[10px]">{{
            coverFailed
              ? 'Cover unavailable'
              : folder.objectCount === 0
                ? 'Empty folder'
                : 'No manga available'
          }}</span>
        </div>
        <span
          v-if="coverUrl && !coverFailed && hasReadingHistory"
          class="absolute bottom-2 left-2 bg-surface-container-lowest/90 px-2 py-1 text-[9px] text-on-surface"
          >Latest read</span
        >
      </div>
      <div class="flex w-full min-w-0 flex-1 flex-col p-3">
        <h3
          class="mb-3 line-clamp-2 break-words text-xs font-bold uppercase leading-tight tracking-tight text-on-surface group-hover:text-primary md:text-sm"
          :title="folder.name"
        >
          {{ folder.name }}
        </h3>
        <div class="mt-auto flex items-center justify-between gap-2">
          <span class="text-[10px] text-on-surface-variant"
            >{{ folder.objectCount.toLocaleString() }}
            {{ folder.objectCount === 1 ? 'item' : 'items' }}</span
          >
          <span class="material-symbols-outlined text-base text-primary" aria-hidden="true"
            >chevron_right</span
          >
        </div>
      </div>
    </button>
    <div ref="menuRef" class="absolute right-2 top-4">
      <button
        ref="menuButton"
        type="button"
        class="flex h-8 w-8 cursor-pointer items-center justify-center bg-surface-container-high/90 text-on-surface hover:bg-surface-container-highest hover:text-primary focus-visible:outline-2 focus-visible:outline-primary"
        :aria-label="`Actions for ${folder.name}`"
        :aria-expanded="menuOpen"
        @click="menuOpen = !menuOpen"
      >
        <span class="material-symbols-outlined text-lg" aria-hidden="true">more_vert</span>
      </button>
      <div
        v-if="menuOpen"
        class="absolute right-0 top-9 z-20 min-w-40 bg-surface-container-highest/95 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
      >
        <button
          v-for="action in ['rename', 'move', 'delete'] as const"
          :key="action"
          type="button"
          class="flex min-h-11 w-full cursor-pointer items-center gap-3 px-3 py-2 text-left text-xs font-bold uppercase tracking-wider hover:bg-surface-container-low focus-visible:outline-2 focus-visible:outline-primary"
          :class="action === 'delete' ? 'text-error' : 'text-on-surface'"
          @click="act(action)"
        >
          <span class="material-symbols-outlined text-base" aria-hidden="true">{{
            action === 'rename' ? 'edit' : action === 'move' ? 'drive_file_move' : 'delete'
          }}</span
          >{{ action }}
        </button>
      </div>
    </div>
  </article>
</template>
