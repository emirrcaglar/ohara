<script setup lang="ts">
import { computed, ref } from 'vue'
import type { MangaRow, AudioRow } from '../types/api'
const AUDIO_COVERS = Object.values(
  import.meta.glob('../assets/audio-cover/*.png', {
    eager: true,
    query: '?url',
    import: 'default',
  }),
) as string[]

const props = defineProps<{
  manga?: MangaRow
  audio?: AudioRow
  coverUrl?: string
  stats?: string
  category?: string
}>()

const emit = defineEmits<{
  click: [item: MangaRow | AudioRow]
  delete: [item: MangaRow]
}>()

const menuOpen = ref(false)

const audioCover = computed(() => {
  if (!props.audio) return null
  return AUDIO_COVERS[props.audio.id % AUDIO_COVERS.length]
})

function handleClick() {
  if (menuOpen.value) {
    menuOpen.value = false
    return
  }

  if (props.manga) {
    emit('click', props.manga)
    return
  }

  if (props.audio) {
    emit('click', props.audio)
  }
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function deleteManga() {
  if (!props.manga) return

  menuOpen.value = false
  emit('delete', props.manga)
}
</script>

<template>
  <div
    class="group h-full bg-surface-container-low transition-none flex flex-col hover:bg-surface-container-high cursor-pointer"
    @click="handleClick"
  >
    <div
      class="aspect-3/4 overflow-hidden bg-surface-container-lowest relative flex items-center justify-center"
    >
      <template v-if="manga">
        <img
          :src="coverUrl"
          class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity grayscale group-hover:grayscale-0"
        />
      </template>
      <template v-else>
        <img
          :src="audioCover!"
          class="w-full h-full object-cover opacity-60 group-hover:opacity-90 transition-opacity grayscale group-hover:grayscale-0"
        />
      </template>

      <div class="absolute top-0 right-0 p-2 flex items-start gap-2">
        <span
          class="bg-secondary-container text-on-secondary-container px-2 py-1 text-[9px] font-black uppercase"
          >{{ manga?.fileExtension || audio?.fileExtension }}</span
        >

        <div v-if="manga" class="relative">
          <button
            class="flex h-7 w-7 items-center justify-center bg-surface-container-high/90 text-on-surface transition-colors hover:bg-surface-container-highest"
            type="button"
            aria-label="Open manga actions"
            @click.stop="toggleMenu"
          >
            <span class="material-symbols-outlined text-base">more_vert</span>
          </button>

          <div
            v-if="menuOpen"
            class="absolute right-0 top-8 z-20 min-w-28 border border-outline-variant/25 bg-surface-container-highest py-1 shadow-xl"
            @click.stop
          >
            <button
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-[10px] font-bold uppercase tracking-wider text-error transition-colors hover:bg-error/10"
              type="button"
              @click="deleteManga"
            >
              <span class="material-symbols-outlined text-sm">delete</span>
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="p-2 md:p-4 relative flex flex-1 flex-col">
      <p class="hidden md:block text-[9px] text-secondary font-bold tracking-widest uppercase mb-1">
        {{ category }}
      </p>
      <h3
        class="font-bold text-xs md:text-lg leading-tight mb-1 md:mb-2 tracking-tight group-hover:text-primary transition-colors uppercase"
      >
        {{ manga?.title || audio?.title }}
      </h3>
      <div
        class="mt-auto flex justify-between items-center pt-1 md:pt-2 border-t border-outline-variant/15"
      >
        <span class="text-[9px] md:text-[10px] font-mono text-outline">{{ stats }}</span>
        <span class="material-symbols-outlined text-sm text-primary">arrow_forward</span>
      </div>
    </div>
  </div>
</template>
