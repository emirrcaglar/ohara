<script setup lang="ts">
defineProps<{
  title: string
  subtitle: string
  image: string
  isActive?: boolean
}>()
</script>

<template>
  <div class="group flex gap-4 p-4 transition-colors cursor-pointer border-l-4"
    :class="isActive ? 'bg-surface-container-high border-primary' : 'hover:bg-surface-container-highest border-transparent'">

    <div class="w-12 h-12 bg-surface-container-lowest shrink-0">
      <img :src="image" class="w-full h-full object-cover grayscale" :class="isActive ? 'opacity-50' : 'opacity-30 group-hover:opacity-80'" />
    </div>

    <div class="min-w-0">
      <p class="text-xs font-black uppercase tracking-tight truncate" :class="isActive ? 'text-primary' : 'text-on-surface'">
        {{ title }}
      </p>
      <p class="text-[10px] text-white/40 uppercase">{{ subtitle }}</p>
    </div>

    <div class="ml-auto flex items-center">
      <span v-if="isActive" class="material-symbols-outlined text-secondary text-sm">equalizer</span>
      <span v-else class="material-symbols-outlined text-white/40 text-sm opacity-0 group-hover:opacity-100">drag_handle</span>
    </div>
  </div>
</template>
