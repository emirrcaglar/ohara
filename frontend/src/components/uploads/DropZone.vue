<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  openPickerOnClick?: boolean
}>(), {
  openPickerOnClick: true
})

const emit = defineEmits<{
  filesSelected: [files: File[]]
  openDialog: []
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)

function openPicker() {
  fileInputRef.value?.click()
}

function handleInputChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return

  emit('filesSelected', Array.from(target.files))

  // Reset value so selecting the same file again still triggers change.
  target.value = ''
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    handleZoneClick()
  }
}

function handleZoneClick() {
  if (props.openPickerOnClick) {
    openPicker()
    return
  }

  emit('openDialog')
}
</script>

<template>
  <div
    class="group relative aspect-[21/9] border-2 border-dashed border-outline-variant/30 bg-surface-container-lowest hover:border-primary-container transition-all flex flex-col items-center justify-center cursor-pointer overflow-hidden"
    role="button"
    tabindex="0"
    @click="handleZoneClick"
    @keydown="handleKeydown"
  >
    <input
      ref="fileInputRef"
      class="hidden"
      type="file"
      multiple
      accept=".cbz,.epub,.mp4"
      @change="handleInputChange"
    />
    <div class="absolute inset-0 bg-primary-container/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
    <span class="material-symbols-outlined text-primary text-6xl mb-4 group-hover:scale-110 transition-transform">cloud_upload</span>
    <p class="text-on-surface font-bold tracking-tight uppercase">Drop files here or click to browse</p>
    <p class="text-on-surface-variant text-xs mt-2 font-mono">SUPPORTED: .CBZ // .EPUB // .MP4</p>
    <div class="absolute top-4 right-4 text-[8px] font-mono text-secondary-fixed-dim/40 leading-none text-right">
      PROTOCOL: X-TRANSFER_04<br/>ENCRYPTION: AES_256
    </div>
  </div>
</template>
