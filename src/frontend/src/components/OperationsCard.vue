<script setup lang="ts">
import { ArrowRight, RefreshCw } from 'lucide-vue-next';
</script>

<template>
  <div class="bg-surface-container-low p-6 flex flex-col gap-4">
    <h2 class="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant">Operations</h2>
    <div class="grid grid-cols-1 gap-2">
      <button class="bg-primary-container text-on-primary-container font-black py-4 px-6 uppercase text-xs tracking-widest hover:bg-primary transition-none flex justify-between items-center group">
        Process Data
        <ArrowRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
      </button>
      <button class="border-2 border-primary-container text-primary-container font-black py-4 px-6 uppercase text-xs tracking-widest hover:bg-primary-container hover:text-on-primary-container transition-none flex justify-between items-center">
        Sync Nodes
        <RefreshCw class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>

