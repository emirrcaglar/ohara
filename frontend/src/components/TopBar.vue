<script setup lang="ts">
defineEmits(['toggleSidebar'])
</script>

<template>
  <header class="flex md:hidden items-center px-4 h-16 w-full bg-surface-dim z-30">
    <button
      class="text-on-surface-variant hover:text-on-surface p-1"
      aria-label="Toggle sidebar"
      @click="$emit('toggleSidebar')"
    >
      <span class="material-symbols-outlined">menu</span>
    </button>
    <span class="ml-4 text-xl font-black text-primary-container uppercase tracking-tighter">
      OHARA
    </span>
  </header>
</template>
