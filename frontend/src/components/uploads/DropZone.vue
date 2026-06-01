<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    openPickerOnClick?: boolean
  }>(),
  {
    openPickerOnClick: true,
  },
)

const emit = defineEmits<{
  filesSelected: [files: File[]]
  openDialog: []
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
let dragCounter = 0

function handleDragenter(event: DragEvent) {
  event.preventDefault()
  dragCounter++
  isDragging.value = true
}

function handleDragover(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy'
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

  const files = droppedFiles(event.dataTransfer)
  if (files.length > 0) emit('filesSelected', files)
}

function droppedFiles(dataTransfer: DataTransfer | null): File[] {
  if (!dataTransfer) return []

  const itemFiles = Array.from(dataTransfer.items ?? [])
    .filter((item) => item.kind === 'file')
    .map((item) => item.getAsFile())
    .filter((file): file is File => file !== null)

  if (itemFiles.length > 0) return itemFiles
  return Array.from(dataTransfer.files ?? [])
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
    @dragenter="handleDragenter"
    @dragover="handleDragover"
    @dragleave="handleDragleave"
    @drop="handleDrop"
  >
    <input
      ref="fileInputRef"
      class="hidden"
      type="file"
      multiple
      accept=".cbz,.mp3,.wav,.flac,.ogg,.m4a,.aac,.mp4,.mkv,.webm,.mov,.avi,.m4v"
      @change="handleInputChange"
    />
    <Transition name="fade">
      <div
        v-if="isDragging"
        class="absolute inset-0 bg-primary-container/20 flex flex-col items-center justify-center pointer-events-none z-10"
      >
        <span class="material-symbols-outlined text-primary text-5xl mb-2">file_download</span>
        <p class="text-primary font-bold text-sm uppercase tracking-tight">Release to upload</p>
      </div>
    </Transition>
    <div
      class="absolute inset-0 bg-primary-container/5 opacity-0 group-hover:opacity-100 transition-opacity"
      :class="{ '!opacity-100': isDragging }"
    ></div>
    <span
      class="material-symbols-outlined text-primary text-6xl mb-4 group-hover:scale-110 transition-transform"
      >cloud_upload</span
    >
    <p class="text-on-surface font-bold tracking-tight uppercase">
      Drop files here or click to browse
    </p>
    <p class="text-on-surface-variant text-xs mt-2 font-mono">SUPPORTED: .CBZ // AUDIO // VIDEO</p>
    <div
      class="absolute top-4 right-4 text-[8px] font-mono text-secondary-fixed-dim/40 leading-none text-right"
    >
      PROTOCOL: X-TRANSFER_04<br />ENCRYPTION: AES_256
    </div>
  </div>
</template>
