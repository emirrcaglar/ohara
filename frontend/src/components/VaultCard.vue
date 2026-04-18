<script setup lang="ts">
import type { MangaRow, AudioRow } from '../types/api'

defineProps<{
  manga?: MangaRow
  audio?: AudioRow
  coverUrl?: string
  stats?: string
  category?: string
}>()
</script>

<template>
  <div class="group bg-surface-container-low transition-none flex flex-col hover:bg-surface-container-high cursor-pointer">
    <div class="aspect-[3/4] overflow-hidden bg-surface-container-lowest relative flex items-center justify-center">
      <template v-if="manga">
        <img :src="coverUrl" class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity grayscale group-hover:grayscale-0" />
      </template>
      <template v-else>
        <div class="w-full h-full absolute top-0 left-0 bg-gradient-to-br from-surface-container-lowest to-surface-container-high"></div>
        <span class="material-symbols-outlined text-6xl text-outline-variant group-hover:text-secondary-container transition-colors">graphic_eq</span>
      </template>

      <div class="absolute top-0 right-0 p-2">
        <span class="bg-secondary-container text-on-secondary-container px-2 py-1 text-[9px] font-black uppercase">{{ manga?.fileExtension || audio?.fileExtension }}</span>
      </div>
    </div>

    <div class="p-4 relative">
      <p class="text-[9px] text-secondary font-bold tracking-widest uppercase mb-1">{{ category }}</p>
      <h3 class="font-bold text-lg leading-tight mb-2 tracking-tight group-hover:text-primary transition-colors uppercase">{{ manga?.title || audio?.title }}</h3>
      <div class="flex justify-between items-center pt-2 border-t border-outline-variant/15">
        <span class="text-[10px] font-mono text-outline">{{ stats }}</span>
        <span class="material-symbols-outlined text-sm text-primary">arrow_forward</span>
      </div>
    </div>
  </div>
</template>
