<script setup lang="ts">
defineProps<{
  title: string
  subLabel: string
  image: string
  bitrate: string
}>()
</script>

<template>
  <div class="relative flex-1 bg-surface-container-lowest group overflow-hidden border-0">
    <img :src="image" class="w-full h-full object-cover opacity-80 grayscale group-hover:grayscale-0 transition-all duration-700" />

    <div class="absolute top-8 left-8 text-left">
      <div class="inline-block bg-secondary-container px-3 py-1 mb-2">
        <span class="text-[10px] font-black text-on-secondary-container tracking-widest uppercase">{{ subLabel }}</span>
      </div>
      <h2 class="text-4xl font-black tracking-tighter text-white uppercase leading-none">{{ title }}</h2>
    </div>

    <div class="absolute bottom-8 right-8 text-right font-mono">
      <p class="text-primary text-sm font-bold">BITRATE: {{ bitrate }}</p>
      <p class="text-white/40 text-[10px]">BUFFERING_OPTIMIZED</p>
    </div>

    <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
      <button class="w-24 h-24 bg-primary-container text-on-primary-container flex items-center justify-center hover:scale-105 transition-transform">
        <span class="material-symbols-outlined text-5xl" style="font-variation-settings: 'FILL' 1;">play_arrow</span>
      </button>
    </div>
  </div>
</template>
