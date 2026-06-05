<script setup lang="ts">
import TransferItem from '../uploads/TransferItem.vue'
import type { useUploadStore } from '../../stores/upload'

defineProps<{
  uploadStore: ReturnType<typeof useUploadStore>
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <section class="fixed inset-0 z-50 flex justify-end">
    <button
      class="absolute inset-0 bg-surface-container-lowest/50 backdrop-blur-sm"
      type="button"
      aria-label="Close transfers panel"
      @click="emit('close')"
    ></button>

    <aside
      class="relative h-dvh w-full max-w-md bg-surface-container-low border-l border-outline/30 flex flex-col pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)]"
    >
      <div class="p-6 border-b border-outline/30 flex items-center justify-between">
        <h3 class="text-xs font-black uppercase tracking-widest text-primary">Active_Transfers</h3>
        <button
          class="text-on-surface-variant hover:text-on-surface"
          type="button"
          @click="emit('close')"
        >
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <p
          v-if="uploadStore.loadingTransfers"
          class="text-xs text-on-surface-variant uppercase tracking-widest"
        >
          Loading transfers...
        </p>
        <p
          v-else-if="uploadStore.transfersError"
          class="text-xs text-error uppercase tracking-widest"
        >
          {{ uploadStore.transfersError }}
        </p>
        <p
          v-else-if="uploadStore.visibleTransfers.length === 0"
          class="text-xs text-on-surface-variant uppercase tracking-widest"
        >
          No active transfers
        </p>
        <TransferItem
          v-for="transfer in uploadStore.visibleTransfers"
          :key="transfer.id"
          :name="transfer.name"
          :progress="transfer.progress"
          :sizeInfo="transfer.sizeInfo"
          :status="transfer.status"
          :eta="transfer.eta"
          :speed="transfer.speed"
          :storagePath="transfer.storagePath"
          :canMoveUp="uploadStore.canMoveTransferUp(transfer.id)"
          :canMoveDown="uploadStore.canMoveTransferDown(transfer.id)"
          @cancel="uploadStore.cancelTransfer(transfer.id)"
          @moveUp="uploadStore.moveTransferUp(transfer.id)"
          @moveDown="uploadStore.moveTransferDown(transfer.id)"
        />
      </div>

      <div class="p-6 bg-surface-container-lowest border-t border-outline/30">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-[9px] text-secondary uppercase font-bold">Total Bandwidth</p>
            <p class="text-lg font-black text-primary leading-none mt-1">
              {{ uploadStore.totalBandwidth }}
            </p>
          </div>
          <div>
            <p class="text-[9px] text-secondary uppercase font-bold">Files in Queue</p>
            <p class="text-lg font-black text-on-surface leading-none mt-1">
              {{ uploadStore.visibleTransfers.length }}
            </p>
          </div>
        </div>
      </div>
    </aside>
  </section>
</template>
