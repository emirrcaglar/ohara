<script setup lang="ts">
import SystemToggle from './SystemToggle.vue'

</script>

<template>
  <div class="bg-surface-container-low p-8 border-l-4 border-secondary">
    <div class="flex justify-between items-start mb-6">
      <div>
        <h3 class="text-xl font-black uppercase tracking-tighter">System_Parameters</h3>
        <p class="text-xs text-on-surface-variant mt-1">Configure processing overhead and thread priority</p>
      </div>
      <span class="material-symbols-outlined text-on-surface-variant cursor-pointer hover:text-primary">
        settings_input_component
      </span>
    </div>

    <div class="space-y-6">
      <SystemToggle label="Auto-Extract Compressed" :isActive="true" />
      <SystemToggle label="Verify Hash (Checksum)" :isActive="true" />
      <SystemToggle label="Overwrite Existing" :isActive="false" />
    </div>
  </div>
</template>
