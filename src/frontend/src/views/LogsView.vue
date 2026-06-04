<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { createLogStream } from '../api/logs'
import type { LogEntry } from '../types/api'

const entries = ref<LogEntry[]>([])
const isStreaming = ref(false)
const logContainer = ref<HTMLElement | null>(null)
let eventSource: EventSource | null = null

const entryCount = computed(() => entries.value.length)

function formatTime(isoString: string): string {
  const d = new Date(isoString)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  const ms = String(d.getMilliseconds()).padStart(3, '0')
  return `[${hh}:${mm}:${ss}:${ms}]`
}

function levelClass(level: string): string {
  switch (level) {
    case 'INFO':
      return 'text-primary'
    case 'WARN':
      return 'text-primary-container'
    case 'ERR':
      return 'text-secondary-container font-bold'
    case 'SYS':
      return 'text-secondary'
    default:
      return 'text-outline'
  }
}

function rowClass(level: string): string {
  const base = 'flex gap-4 py-0.5 group hover:bg-surface-container-high transition-colors'
  if (level === 'ERR')
    return base + ' bg-secondary-container/5 border-l-2 border-secondary-container'
  return base
}

function msgClass(level: string): string {
  return level === 'ERR' ? 'text-secondary-container' : 'text-on-surface'
}

async function scrollToBottom() {
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

function startStream() {
  if (eventSource) return
  eventSource = createLogStream((entry) => {
    entries.value.push(entry)
    scrollToBottom()
  })
  isStreaming.value = true
}

function stopStream() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  isStreaming.value = false
}

function toggleStream() {
  if (isStreaming.value) {
    stopStream()
  } else {
    startStream()
  }
}

function clearLogs() {
  entries.value = []
}

function exportCsv() {
  const rows = ['time,level,message']
  for (const e of entries.value) {
    const time = new Date(e.time).toISOString()
    const msg = e.message.replace(/"/g, '""')
    rows.push(`"${time}","${e.level}","${msg}"`)
  }
  const blob = new Blob([rows.join('\n')], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ohara-logs-${Date.now()}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  startStream()
})

onUnmounted(() => {
  stopStream()
})
</script>

<template>
  <div class="flex-1 flex flex-col min-h-0">
    <!-- Controls Bar -->
    <div
      class="bg-surface-container-low px-6 py-3 flex justify-between items-center border-b border-surface-container-high"
    >
      <!-- Left: stream status + entry count -->
      <div class="flex items-center gap-3">
        <!-- Streaming indicator -->
        <div class="flex items-center gap-2">
          <span v-if="isStreaming" class="w-2 h-2 bg-secondary-container animate-pulse"></span>
          <span v-else class="w-2 h-2 bg-outline"></span>
          <span
            :class="
              isStreaming
                ? 'text-[10px] font-bold text-secondary uppercase tracking-widest'
                : 'text-[10px] font-bold text-outline uppercase tracking-widest'
            "
          >
            {{ isStreaming ? 'Live Stream Active' : 'Stream Paused' }}
          </span>
        </div>

        <!-- Separator -->
        <div class="h-4 w-px bg-surface-container-highest"></div>

        <!-- Entry count -->
        <span class="text-[10px] font-mono text-outline uppercase"> {{ entryCount }} ENTRIES </span>
      </div>

      <!-- Right: action buttons -->
      <div class="flex items-center gap-2">
        <!-- Clear Logs -->
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-1.5 bg-surface-container-highest text-on-surface text-[10px] font-bold uppercase tracking-tighter hover:bg-surface-bright transition-colors active:scale-95"
          @click="clearLogs"
        >
          <span class="material-symbols-outlined text-base leading-none">delete_sweep</span>
          Clear Logs
        </button>

        <!-- Export CSV -->
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-1.5 bg-surface-container-highest text-on-surface text-[10px] font-bold uppercase tracking-tighter hover:bg-surface-bright transition-colors active:scale-95"
          @click="exportCsv"
        >
          <span class="material-symbols-outlined text-base leading-none">download</span>
          Export CSV
        </button>

        <!-- Pause / Resume Feed -->
        <button
          type="button"
          class="flex items-center gap-2 px-4 py-1.5 bg-primary-container text-on-primary-container text-[10px] font-bold uppercase tracking-tighter hover:bg-primary transition-colors active:scale-95"
          @click="toggleStream"
        >
          <span class="material-symbols-outlined text-base leading-none">
            {{ isStreaming ? 'pause' : 'play_arrow' }}
          </span>
          {{ isStreaming ? 'Pause Feed' : 'Resume Feed' }}
        </button>
      </div>
    </div>

    <!-- Terminal Console -->
    <div class="flex-1 bg-surface-container-lowest overflow-hidden flex flex-col min-h-0">
      <!-- Terminal header bar -->
      <div
        class="flex items-center justify-between px-4 py-1 bg-surface-container-low border-b border-surface-container-high text-[9px] text-outline font-mono"
      >
        <span>OHARA_KERNEL_FEED_V1.LOG</span>
        <span>UTF-8 // MONO_GRID</span>
      </div>

      <!-- Scrollable log list -->
      <div ref="logContainer" class="flex-1 overflow-y-auto p-4 font-mono text-xs leading-relaxed">
        <!-- Empty state -->
        <div v-if="entries.length === 0" class="h-full flex items-center justify-center">
          <span class="text-outline text-[10px] font-mono uppercase tracking-widest">
            AWAITING LOG STREAM...
          </span>
        </div>

        <!-- Log entries -->
        <template v-else>
          <div v-for="(entry, index) in entries" :key="index" :class="rowClass(entry.level)">
            <!-- Timestamp -->
            <span class="text-outline-variant shrink-0 w-28">
              {{ formatTime(entry.time) }}
            </span>

            <!-- Level badge -->
            <span class="shrink-0 w-16" :class="levelClass(entry.level)">
              {{ entry.level }}
            </span>

            <!-- Message -->
            <span :class="msgClass(entry.level)">
              {{ entry.message }}
            </span>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
