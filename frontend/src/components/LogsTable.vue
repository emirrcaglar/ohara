<script setup lang="ts">
const logs = [
  { timestamp: '12:43:55', source: 'MEDIA_FLUX_A8', status: 'COMPLETE', checksum: '0x88AF...22' },
  { timestamp: '12:40:12', source: 'NODE_ORBIT_01', status: 'COMPLETE', checksum: '0x12CC...B4' },
  { timestamp: '12:35:44', source: 'LOCAL_CACHE_U', status: 'PENDING', checksum: '----' },
  { timestamp: '12:32:01', source: 'SATELLITE_LINK', status: 'COMPLETE', checksum: '0xDD09...1A' },
  { timestamp: '12:28:19', source: 'CORE_ROOT_SRV', status: 'COMPLETE', checksum: '0x55E2...9F' },
];
</script>

<template>
  <div class="bg-surface-container-low flex flex-col flex-1 overflow-hidden">
    <div class="p-6 border-b border-surface-container-highest flex justify-between items-center">
      <h2 class="text-xs font-black uppercase tracking-[0.2em] text-on-surface-variant">Ingest_Logs</h2>
      <span class="text-[10px] text-white/40 tracking-widest uppercase">Last Update: 12:44:01</span>
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

