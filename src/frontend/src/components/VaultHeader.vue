<script setup lang="ts">
defineProps<{
  title: string
  folderCount: number
  mediaCount: number
}>()
defineEmits<{ createFolder: [] }>()
</script>

<template>
  <header class="mb-5 flex flex-wrap items-center justify-between gap-4">
    <div class="min-w-0">
      <h1
        class="break-words text-2xl font-bold uppercase leading-tight tracking-tight text-on-surface md:text-3xl"
      >
        {{ title }}
      </h1>
      <p class="mt-1 text-xs text-on-surface-variant">
        {{ folderCount.toLocaleString() }} folders <span class="mx-1" aria-hidden="true">/</span>
        {{ mediaCount.toLocaleString() }} files
      </p>
    </div>
    <button
      type="button"
      class="flex min-h-10 shrink-0 items-center gap-2 bg-surface-container-high px-3 text-xs font-semibold text-on-surface transition-colors hover:bg-surface-container-highest focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      @click="$emit('createFolder')"
    >
      <span class="material-symbols-outlined text-lg text-primary" aria-hidden="true"
        >create_new_folder</span
      >New folder
    </button>
  </header>
</template>
