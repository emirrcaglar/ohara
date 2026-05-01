<script setup lang="ts">
const logs = [
  { timestamp: 'NOT_YET_IMPLEMENTED', source: 'NOT_YET_IMPLEMENTED', status: 'NOT_YET_IMPLEMENTED', checksum: 'NOT_YET_IMPLEMENTED' },
];
</script>

<template>
  <div class="bg-surface-container-low flex flex-col flex-1 overflow-hidden">
    <div class="p-6 border-b border-surface-container-highest flex justify-between items-center">
      <h2 class="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant">Ingest_Logs</h2>
      <span class="text-[10px] text-white/40 tracking-widest uppercase">Last Update: NOT_YET_IMPLEMENTED</span>
    </div>
    <div class="overflow-y-auto">
      <table class="w-full text-left text-xs">
        <thead>
          <tr class="text-secondary font-bold uppercase tracking-widest border-b border-surface-container-highest">
            <th class="px-6 py-4">Timestamp</th>
            <th class="px-6 py-4">Source_ID</th>
            <th class="px-6 py-4">Status</th>
            <th class="px-6 py-4">Checksum</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-surface-container-highest/30">
          <tr
            v-for="log in logs"
            :key="log.timestamp"
            class="hover:bg-surface-container-high transition-none"
          >
            <td class="px-6 py-4 font-mono text-white/60">{{ log.timestamp }}</td>
            <td class="px-6 py-4 font-bold tracking-tight">{{ log.source }}</td>
            <td class="px-6 py-4">
              <span
                class="font-black"
                :class="log.status === 'COMPLETE' ? 'text-secondary-container' : 'text-primary-container'"
              >
                {{ log.status }}
              </span>
            </td>
            <td class="px-6 py-4 font-mono text-white/40">{{ log.checksum }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

