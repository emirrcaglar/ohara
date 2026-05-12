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
const isDragging = ref(false)
let dragCounter = 0

function handleDragover(event: DragEvent) {
  event.preventDefault()
  if (!isDragging.value) isDragging.value = true
}

function handleDragleave() {
  dragCounter--
  if (dragCounter <= 0) {
    dragCounter = 0
    isDragging.value = false
  }
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  isDragging.value = false
  dragCounter = 0
  if (!event.dataTransfer?.files) return
  const files = Array.from(event.dataTransfer.files)
  if (files.length > 0) emit('filesSelected', files)
}

function handleGlobalDragenter(event: DragEvent) {
  if (event.dataTransfer?.types.includes('Files')) {
    dragCounter++
  }
}

function handleGlobalDragleave(event: DragEvent) {
  if (event.dataTransfer?.types.includes('Files')) {
    handleDragleave()
  }
}

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
  :class="{ '!border-primary-container !bg-primary-container/10': isDragging }"
  @click="handleZoneClick"
  @keydown="handleKeydown"
  @dragover="handleDragover"
  @dragleave="handleDragleave"
  @drop="handleDrop"
>
  <input
    ref="fileInputRef"
    class="hidden"
    type="file"
    multiple
    accept=".cbz,.mp3,.wav,.flac,.ogg,.m4a,.aac"
    @change="handleInputChange"
  />
  <Transition name="fade">
    <div v-if="isDragging" class="absolute inset-0 bg-primary-container/20 flex flex-col items-center justify-center pointer-events-none z-10">
      <span class="material-symbols-outlined text-primary text-5xl mb-2">file_download</span>
      <p class="text-primary font-bold text-sm uppercase tracking-tight">Release to upload</p>
    </div>
  </Transition>
  <div class="absolute inset-0 bg-primary-container/5 opacity-0 group-hover:opacity-100 transition-opacity" :class="{ '!opacity-100': isDragging }"></div>
    <span class="material-symbols-outlined text-primary text-6xl mb-4 group-hover:scale-110 transition-transform">cloud_upload</span>
    <p class="text-on-surface font-bold tracking-tight uppercase">Drop files here or click to browse</p>
    <p class="text-on-surface-variant text-xs mt-2 font-mono">SUPPORTED: .CBZ // .MP3 // .WAV</p>
    <div class="absolute top-4 right-4 text-[8px] font-mono text-secondary-fixed-dim/40 leading-none text-right">
      PROTOCOL: X-TRANSFER_04<br/>ENCRYPTION: AES_256
    </div>
  </div>
</template>
