<script setup lang="ts">
defineProps<{
  label: string
  isActive: boolean
}>()
</script>

<template>
  <div class="flex items-center justify-between border-b border-outline-variant/10 pb-4 last:border-0 last:pb-0">
    <span
      class="text-xs font-bold uppercase tracking-widest transition-colors"
      :class="isActive ? 'text-on-surface' : 'text-on-surface-variant'"
    >
      {{ label }}
    </span>

    <div
      class="w-12 h-6 bg-surface-container-highest p-1 flex transition-all duration-75"
      :class="isActive ? 'justify-end' : 'justify-start'"
    >
      <div
        class="w-4 h-4 transition-colors"
        :class="isActive ? 'bg-primary-container' : 'bg-outline-variant/40'"
      ></div>
    </div>
  </div>
</template>
