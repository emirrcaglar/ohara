<script setup lang="ts">
defineProps<{
  label: string
  value: string
  icon: string
}>()
</script>

<template>
  <div class="space-y-4">
    <h3 class="text-xs font-black uppercase tracking-[0.2em] text-secondary">
      {{ label }}
    </h3>

    <div class="bg-surface-container-low p-1 group">
      <div class="bg-surface-container-highest p-4 flex items-center justify-between border-b-2 border-transparent group-focus-within:border-primary-container transition-all">
        <span class="text-sm font-mono text-on-surface truncate pr-4">
          {{ value }}
        </span>

        <span class="material-symbols-outlined text-on-surface-variant text-sm cursor-pointer hover:text-primary">
          {{ icon }}
        </span>
      </div>
    </div>
  </div>
</template>
