<script setup lang="ts">
import type { CatalogFolder } from '../types/api'

type LibraryTab = 'ALL' | 'CBZ' | 'AUDIO' | 'VIDEO'

defineProps<{
  totalManga?: number
  modelValue?: LibraryTab
  catalogs?: CatalogFolder[]
}>()

defineEmits<{
  'update:modelValue': [value: LibraryTab]
  openCatalog: [folder: CatalogFolder]
}>()

const tabs = ['ALL', 'CBZ', 'AUDIO', 'VIDEO'] as const
</script>

<template>
  <div class="flex flex-wrap items-end justify-between gap-4 mb-4 md:mb-6">
    <div>
      <div class="flex flex-wrap items-start gap-2">
        <button
          v-for="tab in tabs"
          :key="tab"
          @click="$emit('update:modelValue', tab)"
          class="px-3 md:px-6 py-2 font-bold text-xs uppercase transition-colors"
          :class="
            modelValue === tab
              ? 'bg-primary-container text-on-primary-container'
              : 'bg-surface-hover text-secondary hover:text-white'
          "
        >
          {{ tab }}
        </button>

        <button
          v-for="catalog in catalogs"
          :key="catalog.id"
          class="flex items-center gap-2 bg-surface-hover px-3 py-2 text-xs font-bold uppercase leading-none text-primary-container transition-colors hover:text-white md:px-6"
          type="button"
          @click="$emit('openCatalog', catalog)"
        >
          <span>{{ catalog.name }}</span>
          <span
            class="flex h-4 w-4 shrink-0 items-center justify-center rounded-full font-mono text-xs font-black leading-none text-primary-container"
          >
            -{{ catalog.objectCount }}-
          </span>
        </button>
      </div>
    </div>
    <div class="text-right hidden sm:block">
      <p class="text-[10px] text-secondary font-bold uppercase tracking-widest mb-1">
        Index_Status
      </p>
      <p class="text-2xl font-mono text-on-surface">
        {{ totalManga?.toLocaleString() || 0 }} <span class="text-sm text-outline">UNITS</span>
      </p>
    </div>
  </div>
</template>
