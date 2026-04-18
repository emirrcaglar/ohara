<script setup lang="ts">
import { ref } from 'vue'
import DropZone from '../components/uploads/DropZone.vue'
import TransferItem from '../components/uploads/TransferItem.vue'
import SystemInput from '../components/uploads/SystemInput.vue'
import SystemToggle from '../components/uploads/SystemToggle.vue'

const destinationPath = ref('/mnt/storage/media/unprocessed/')
const metadataProfile = ref('AUTO_DETECT_SCRAPER_V2')
const autoExtract = ref(true)
const verifyHash = ref(true)
const overwriteExisting = ref(false)

const transfers = ref([
  {
    id: 1,
    name: 'Neon_Genesis_V01.cbz',
    progress: 47,
    sizeInfo: '420.5 MB / 892.0 MB',
    eta: '02m 45s',
    speed: '8.4 MB/s',
    status: 'active' as const
  },
  {
    id: 2,
    name: 'Ghost_In_The_Shell_1995.mp4',
    progress: 25,
    sizeInfo: '1.2 GB / 4.8 GB',
    eta: '12m 10s',
    speed: '12.1 MB/s',
    status: 'active' as const
  },
  {
    id: 3,
    name: 'Neuromancer_Digital.epub',
    progress: 100,
    sizeInfo: '',
    status: 'complete' as const,
    storagePath: '/books/scifi/'
  }
])
</script>

<template>
  <div class="flex-1 flex gap-0 p-0 overflow-hidden">
    <section class="flex-1 p-12 overflow-y-auto">
      <div class="max-w-4xl mx-auto space-y-12">
        <div class="space-y-6">
          <div class="flex justify-between items-end">
            <h2 class="text-4xl font-black uppercase tracking-tighter">Ingest_Queue</h2>
            <span class="text-secondary text-xs font-bold tracking-widest bg-secondary-container/10 px-3 py-1">READY_FOR_DATA</span>
          </div>

          <DropZone />
        </div>

        <div class="grid grid-cols-2 gap-8">
          <SystemInput
            label="Destination_Path"
            :value="destinationPath"
            icon="folder_open"
          />

          <SystemInput
            label="Metadata_Profile"
            :value="metadataProfile"
            icon="expand_more"
          />
        </div>

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
            <SystemToggle
              label="Auto-Extract Compressed"
              :isActive="autoExtract"
            />
            <SystemToggle
              label="Verify Hash (Checksum)"
              :isActive="verifyHash"
            />
            <SystemToggle
              label="Overwrite Existing"
              :isActive="overwriteExisting"
            />
          </div>
        </div>
      </div>
    </section>

    <aside class="w-80 bg-surface-container-low flex flex-col border-l border-white/5">
      <div class="p-6 border-b border-white/10">
        <div class="flex justify-between items-center">
          <h3 class="text-xs font-black uppercase tracking-widest text-primary">Active_Transfers</h3>
          <div class="flex items-center gap-1">
            <span class="w-2 h-2 bg-secondary rounded-full animate-pulse"></span>
            <span class="text-[10px] font-bold text-secondary">LIVE</span>
          </div>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <TransferItem
          v-for="transfer in transfers"
          :key="transfer.id"
          :name="transfer.name"
          :progress="transfer.progress"
          :sizeInfo="transfer.sizeInfo || undefined"
          :status="transfer.status"
          :eta="transfer.eta"
          :speed="transfer.speed"
          :storagePath="transfer.storagePath"
        />
      </div>

      <div class="p-6 bg-surface-container-lowest border-t border-white/10">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-[9px] text-secondary uppercase font-bold">Total Bandwidth</p>
            <p class="text-lg font-black text-primary leading-none mt-1">20.5 MB/s</p>
          </div>
          <div>
            <p class="text-[9px] text-secondary uppercase font-bold">Files in Queue</p>
            <p class="text-lg font-black text-on-surface leading-none mt-1">{{ transfers.length }}</p>
          </div>
        </div>

        <button class="w-full mt-6 bg-primary-container text-on-primary-container py-3 font-black uppercase text-xs tracking-widest hover:bg-primary transition-colors">
          Process_All
        </button>
      </div>
    </aside>
  </div>
</template>
