<script setup lang="ts">
defineProps<{
  label: string
  labelColor?: string // e.g. 'text-secondary' or 'text-primary'
}>()
</script>

<template>
  <div class="bg-surface-container-low p-4 aspect-square flex flex-col justify-between border border-white/5">
    <span class="text-[10px] font-bold uppercase" :class="labelColor || 'text-secondary'">
      {{ label }}
    </span>
    <div class="flex items-end">
      <slot />
    </div>
  </div>
</template>
