<script setup lang="ts">
import { computed } from 'vue'
import type { MangaRow, AudioRow } from '../types/api'
const AUDIO_COVERS = Object.values(
  import.meta.glob('../assets/audio-cover/*.png', { eager: true, query: '?url', import: 'default' })
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
}>()

const audioCover = computed(() => {
  if (!props.audio) return null
  return AUDIO_COVERS[props.audio.id % AUDIO_COVERS.length]
})

function handleClick() {
  if (props.manga) {
    emit('click', props.manga)
    return
  }

  if (props.audio) {
    emit('click', props.audio)
  }
}
</script>

<template>
  <div
    class="group bg-surface-container-low transition-none flex flex-col hover:bg-surface-container-high cursor-pointer"
    @click="handleClick"
  >
    <div class="aspect-[3/4] overflow-hidden bg-surface-container-lowest relative flex items-center justify-center">
      <template v-if="manga">
        <img :src="coverUrl" class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity grayscale group-hover:grayscale-0" />
      </template>
      <template v-else>
        <img :src="audioCover!" class="w-full h-full object-cover opacity-60 group-hover:opacity-90 transition-opacity grayscale group-hover:grayscale-0" />
      </template>

      <div class="absolute top-0 right-0 p-2">
        <span class="bg-secondary-container text-on-secondary-container px-2 py-1 text-[9px] font-black uppercase">{{ manga?.fileExtension || audio?.fileExtension }}</span>
      </div>
    </div>

    <div class="p-2 md:p-4 relative">
      <p class="hidden md:block text-[9px] text-secondary font-bold tracking-widest uppercase mb-1">{{ category }}</p>
      <h3 class="font-bold text-xs md:text-lg leading-tight mb-1 md:mb-2 tracking-tight group-hover:text-primary transition-colors uppercase">{{ manga?.title || audio?.title }}</h3>
      <div class="flex justify-between items-center pt-1 md:pt-2 border-t border-outline-variant/15">
        <span class="text-[9px] md:text-[10px] font-mono text-outline">{{ stats }}</span>
        <span class="material-symbols-outlined text-sm text-primary">arrow_forward</span>
      </div>
    </div>
  </div>
</template>
