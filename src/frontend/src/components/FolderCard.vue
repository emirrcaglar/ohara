<script setup lang="ts">
import { ref } from 'vue'
import type { CatalogFolder } from '../types/api'

const props = defineProps<{
  folder: CatalogFolder
}>()

interface MenuAnchor {
  top: number
  right: number
}

const emit = defineEmits<{
  open: [folder: CatalogFolder]
  rename: [folder: CatalogFolder]
  move: [folder: CatalogFolder]
  delete: [folder: CatalogFolder, anchor: MenuAnchor]
}>()

const menuOpen = ref(false)

function openFolder() {
  if (menuOpen.value) {
    menuOpen.value = false
    return
  }

  emit('open', props.folder)
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function menuAnchor(event: MouseEvent): MenuAnchor {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  return {
    top: rect.top,
    right: window.innerWidth - rect.right,
  }
}

function emitAction(action: 'rename' | 'move' | 'delete', event: MouseEvent) {
  menuOpen.value = false

  if (action === 'rename') {
    emit('rename', props.folder)
  } else if (action === 'move') {
    emit('move', props.folder)
  } else {
    emit('delete', props.folder, menuAnchor(event))
  }
}
</script>

<template>
  <div
    class="group relative cursor-pointer bg-surface-container-low p-5 text-left transition-none hover:bg-surface-container-high md:p-6"
    role="button"
    tabindex="0"
    @click="openFolder"
    @keydown.enter.prevent="openFolder"
    @keydown.space.prevent="openFolder"
  >
    <div class="mb-8 flex items-start justify-between gap-4">
      <span class="material-symbols-outlined text-5xl text-primary-container">folder</span>

      <div class="flex items-start gap-2">
        <div class="relative">
          <button
            class="flex h-7 w-7 items-center justify-center bg-surface-container-high/90 text-on-surface transition-colors hover:bg-surface-container-highest"
            type="button"
            aria-label="Open folder actions"
            @click.stop="toggleMenu"
          >
            <span class="material-symbols-outlined text-base">more_vert</span>
          </button>

          <div
            v-if="menuOpen"
            class="absolute right-0 top-8 z-20 min-w-32 border border-outline-variant/25 bg-surface-container-highest py-1 shadow-xl"
            @click.stop
          >
            <button
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-[10px] font-bold uppercase tracking-wider text-on-surface transition-colors hover:bg-surface-container-high"
              type="button"
              @click="emitAction('rename', $event)"
            >
              <span class="material-symbols-outlined text-sm">edit</span>
              Rename
            </button>
            <button
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-[10px] font-bold uppercase tracking-wider text-on-surface transition-colors hover:bg-surface-container-high"
              type="button"
              @click="emitAction('move', $event)"
            >
              <span class="material-symbols-outlined text-sm">drive_file_move</span>
              Move
            </button>
            <button
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-[10px] font-bold uppercase tracking-wider text-error transition-colors hover:bg-error/10"
              type="button"
              @click="emitAction('delete', $event)"
            >
              <span class="material-symbols-outlined text-sm">delete</span>
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <p class="mb-1 text-[9px] font-bold uppercase tracking-widest text-secondary">CATALOG</p>
    <h3
      class="text-lg font-bold uppercase leading-tight tracking-tight text-on-surface transition-colors group-hover:text-primary"
    >
      {{ folder.name }}
    </h3>

    <div class="mt-4 flex items-center justify-between bg-surface-container-lowest px-3 py-2">
      <span class="font-mono text-[10px] text-on-surface-variant">
        OBJECTS: {{ folder.objectCount }}
      </span>
      <span class="material-symbols-outlined text-sm text-primary-container">folder_open</span>
    </div>
  </div>
</template>
