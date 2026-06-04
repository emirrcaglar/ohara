<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next';

const metrics = [
  { label: 'CPU_LOAD', value: 'NOT_YET_IMPLEMENTED', progress: 0 },
  { label: 'RAM_USAGE', value: 'NOT_YET_IMPLEMENTED', progress: 0 },
  { label: 'NET_THROUGHPUT', value: 'NOT_YET_IMPLEMENTED', progress: 0 },
];
</script>

<template>
  <div class="bg-surface-container-low p-6 flex flex-col gap-8 relative overflow-hidden">
    <div class="flex justify-between items-start">
      <h2 class="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant">System_Telemetry</h2>
      <AlertTriangle class="text-secondary-container w-4 h-4 fill-secondary-container" />
    </div>

    <div class="space-y-8">
      <div v-for="metric in metrics" :key="metric.label" class="space-y-2">
        <div class="flex justify-between text-[10px] font-bold tracking-tighter uppercase">
          <span>{{ metric.label }}</span>
          <span class="text-primary-container">{{ metric.value }}</span>
        </div>
        <div class="h-3 bg-surface-container-highest w-full relative">
          <div
            class="h-full bg-primary-container transition-all duration-500"
            :style="{ width: `${metric.progress}%` }"
          ></div>
          <div class="absolute top-0 right-0 h-full w-1 bg-secondary-container"></div>
        </div>
      </div>
    </div>
  </div>
</template>

