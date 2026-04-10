<script setup lang="ts">
import DropZone from '../components/uploads/DropZone.vue'
import TransferItem from '../components/uploads/TransferItem.vue'
import SystemInput from '../components/uploads/SystemInput.vue'
import SystemParameters from '../components/uploads/SystemParameters.vue'
</script>

<template>
  <div class="flex-1 flex gap-0 p-0 overflow-hidden">
    <section class="flex-1 p-12 overflow-y-auto">
      <div class="max-w-4xl mx-auto space-y-12">
        <div class="space-y-6">
          <h2 class="text-4xl font-black uppercase tracking-tighter">Ingest_Queue</h2>
          <DropZone />
        </div>

        <div class="grid grid-cols-2 gap-8">
          <SystemInput
            label="Destination_Path"
            value="/mnt/storage/media/unprocessed/"
            icon="folder_open"
          />

          <SystemInput
            label="Metadata_Profile"
            value="AUTO_DETECT_SCRAPER_V2"
            icon="expand_more"
          />
        </div>

        <SystemParameters />

        </div>
    </section>

    <aside class="w-80 bg-surface-container-low flex flex-col border-l border-outline-variant/10">
      <div class="p-6 border-b border-outline-variant/10">
        <h3 class="text-xs font-black uppercase text-primary">Active_Transfers</h3>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <TransferItem
          name="Neon_Genesis_V01.cbz"
          :progress="47"
          sizeInfo="420.5 MB / 892.0 MB"
          eta="02m 45s"
          speed="8.4 MB/s"
          status="active"
        />
        <TransferItem
          name="Neuromancer_Digital.epub"
          :progress="100"
          sizeInfo="892.0 MB / 892.0 MB"
          status="complete"
          storagePath="/books/scifi/"
        />
        </div>
    </aside>
  </div>
</template>
