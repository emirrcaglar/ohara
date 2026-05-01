<script setup lang="ts">
import { Settings, Power } from 'lucide-vue-next';
</script>

<template>
  <header class="flex items-center justify-between px-8 h-16 w-full bg-surface-dim z-30">
    <div class="flex items-center gap-4">
      <span class="text-xl font-black text-primary-container px-4 uppercase tracking-tighter">OHARA_MEDIA</span>
      <div class="h-4 w-[2px] bg-surface-container-highest"></div>
      <span class="text-[10px] tracking-[0.3em] text-white/40 uppercase">DASHBOARD_NOT_YET_IMPLEMENTED</span>
    </div>

    <div class="flex items-center gap-6">
      <div class="relative hidden md:flex items-center">
        <input
          class="bg-surface-container-low border-0 text-[10px] tracking-widest uppercase py-2 px-4 w-64 focus:ring-1 focus:ring-primary-container text-white placeholder-white/20"
          placeholder="SEARCH_SYSTEM..."
          type="text"
        />
      </div>
      <div class="flex items-center gap-4">
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

