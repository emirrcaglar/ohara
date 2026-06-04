<script setup lang="ts">
defineProps<{
  title: string;
  subtitle: string;
  image: string;
  badge?: string;
  badgeType?: 'primary' | 'secondary';
}>();
</script>

<template>
  <div class="bg-surface-container-low h-48 relative group overflow-hidden">
    <img
      class="absolute inset-0 w-full h-full object-cover opacity-40 group-hover:scale-105 transition-transform duration-500"
      :src="image"
      referrerpolicy="no-referrer"
    />
    <div class="absolute bottom-0 left-0 p-4 w-full bg-gradient-to-t from-background to-transparent">
      <span class="text-[10px] text-secondary font-bold uppercase tracking-[0.3em]">{{ subtitle }}</span>
      <h3 class="text-lg font-black uppercase tracking-tighter">{{ title }}</h3>
    </div>
    <div
      v-if="badge"
      class="absolute top-4 right-4 px-2 py-1 text-[10px] font-black uppercase"
      :class="badgeType === 'primary' ? 'bg-primary-container text-on-primary-container' : 'bg-surface-container-highest text-white'"
    >
      {{ badge }}
    </div>
  </div>
</template>

