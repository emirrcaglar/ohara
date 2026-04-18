<script setup lang="ts">
defineProps<{
  totalManga?: number
  modelValue?: 'ALL' | 'CBZ' | 'AUDIO'
}>()

defineEmits<{
  'update:modelValue': [value: 'ALL' | 'CBZ' | 'AUDIO']
}>()

const tabs = ['ALL', 'CBZ', 'AUDIO'] as const
</script>

<template>
  <div class="flex items-end justify-between mb-12">
    <div>
      <h2 class="text-4xl font-black tracking-tighter uppercase mb-4">Central_Vault</h2>
      <div class="flex gap-2">
        <button
          v-for="tab in tabs"
          :key="tab"
          @click="$emit('update:modelValue', tab)"
          class="px-6 py-2 font-bold text-xs uppercase transition-colors"
          :class="modelValue === tab ? 'bg-primary-container text-on-primary-container' : 'bg-surface-hover text-secondary hover:text-white'"
        >
          {{ tab }}
        </button>
      </div>
    </div>
    <div class="text-right hidden sm:block">
      <p class="text-[10px] text-secondary font-bold uppercase tracking-widest mb-1">Index_Status</p>
      <p class="text-2xl font-mono text-on-surface">{{ totalManga?.toLocaleString() || 0 }} <span class="text-sm text-outline">UNITS</span></p>
    </div>
  </div>
</template>