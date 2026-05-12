<script setup lang="ts">
import { Settings, Power } from 'lucide-vue-next';

const emit = defineEmits(['toggleSidebar']);
</script>

<template>
  <header class="flex items-center justify-between px-4 md:px-8 h-16 w-full bg-surface-dim z-30">
    <div class="flex items-center gap-4">
      <button class="md:hidden text-on-surface-variant hover:text-on-surface p-1" @click="$emit('toggleSidebar')" aria-label="Toggle sidebar">
        <span class="material-symbols-outlined">menu</span>
      </button>
      <span class="text-xl font-black text-primary-container uppercase tracking-tighter">
        <span class="md:hidden">OHARA</span>
        <span class="hidden md:inline">OHARA_MEDIA</span>
      </span>
      <div class="hidden md:block h-4 w-[2px] bg-surface-container-highest"></div>
      <span class="hidden md:block text-[10px] tracking-[0.3em] text-white/40 uppercase">DASHBOARD_NOT_YET_IMPLEMENTED</span>
    </div>

    <div class="flex items-center gap-3 md:gap-6">
      <div class="relative hidden md:flex items-center">
        <input
          class="bg-surface-container-low border-0 text-[10px] tracking-widest uppercase py-2 px-4 w-64 focus:ring-1 focus:ring-primary-container text-white placeholder-white/20"
          placeholder="SEARCH_SYSTEM..."
          type="text"
        />
      </div>
      <div class="flex items-center gap-2 md:gap-4">
        <button class="text-white/80 hover:text-secondary-container transition-none">
          <Settings class="w-5 h-5" />
        </button>
        <button class="text-white/80 hover:text-secondary-container transition-none">
          <Power class="w-5 h-5" />
        </button>
      </div>
    </div>
  </header>
</template>

