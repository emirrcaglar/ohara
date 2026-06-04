<script setup lang="ts">
import type { CatalogFolder } from '../../types/api'

interface MoveDestinationOption {
  id: number | null
  folder: CatalogFolder | null
}

defineProps<{
  movingSubjectName: string
  moveDestinationOptions: MoveDestinationOption[]
  expandedBreadcrumbs: Set<string>
  catalogName: (folder: CatalogFolder) => string
  breadcrumbKey: (folder: CatalogFolder | null) => string
  breadcrumbParts: (folder: CatalogFolder | null) => CatalogFolder[]
  visibleBreadcrumbParts: (folder: CatalogFolder | null) => CatalogFolder[]
  isBreadcrumbCollapsed: (folder: CatalogFolder | null) => boolean
}>()

const emit = defineEmits<{
  close: []
  selectDestination: [catalogId: number | null]
  toggleBreadcrumb: [folder: CatalogFolder | null]
}>()
</script>

<template>
  <section class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-surface-container-lowest/70 backdrop-blur-sm"
      type="button"
      aria-label="Close catalog move dialog"
      @click="emit('close')"
    ></button>

    <div
      class="relative w-full max-w-2xl bg-surface-container-highest/80 p-3 shadow-[0_0_40px_rgba(14,14,14,0.4)] backdrop-blur"
      role="dialog"
      aria-modal="true"
      aria-labelledby="move-catalog-title"
    >
      <div class="bg-surface-container-low p-5 md:p-6">
        <div class="flex items-start justify-between gap-6">
          <div>
            <p class="text-[9px] font-black uppercase tracking-widest text-secondary">Catalogs</p>
            <h3
              id="move-catalog-title"
              class="mt-1 text-2xl font-black uppercase italic tracking-tighter text-on-surface"
            >
              Move_Target
            </h3>
          </div>
          <button
            class="bg-surface-container-high px-3 py-2 text-on-surface-variant transition-colors hover:bg-surface-container-lowest hover:text-on-surface"
            type="button"
            aria-label="Close catalog move dialog"
            @click="emit('close')"
          >
            <span class="material-symbols-outlined text-lg">close</span>
          </button>
        </div>

        <div class="mt-5 bg-surface-container-lowest p-4">
          <p class="text-[9px] font-black uppercase tracking-widest text-secondary">Object</p>
          <p
            class="mt-1 truncate text-lg font-black uppercase tracking-tight text-primary-container"
          >
            {{ movingSubjectName }}
          </p>
        </div>
      </div>

      <div class="mt-3 max-h-[60vh] overflow-y-auto bg-surface-container-low p-3">
        <p
          v-if="moveDestinationOptions.length === 0"
          class="bg-surface-container-lowest p-5 text-xs font-bold uppercase tracking-widest text-on-surface-variant"
        >
          No catalog destinations available
        </p>

        <div v-else class="space-y-2">
          <div
            v-for="option in moveDestinationOptions"
            :key="option.id ?? 'root'"
            class="grid gap-2 bg-surface-container-lowest p-2 md:grid-cols-[1fr_auto]"
          >
            <button
              class="bg-surface-container-high p-4 text-left transition-colors hover:bg-primary-container hover:text-on-primary-container"
              type="button"
              @click="emit('selectDestination', option.id)"
            >
              <p class="text-[9px] font-black uppercase tracking-widest text-secondary">
                Destination
              </p>
              <p class="mt-1 text-base font-black uppercase tracking-tight">
                {{ option.folder?.name ?? 'CENTRAL' }}
              </p>

              <div
                class="mt-3 flex flex-wrap items-center gap-1 text-[10px] font-bold uppercase tracking-widest text-on-surface-variant"
                :title="option.folder ? catalogName(option.folder) : 'CENTRAL'"
              >
                <span class="bg-surface-container-low px-2 py-1 text-primary-container">
                  CENTRAL
                </span>
                <template v-if="isBreadcrumbCollapsed(option.folder)">
                  <span class="material-symbols-outlined text-[13px] text-primary-container">
                    chevron_right
                  </span>
                  <span class="bg-surface-container-low px-2 py-1">…</span>
                </template>
                <template v-for="part in visibleBreadcrumbParts(option.folder)" :key="part.id">
                  <span class="material-symbols-outlined text-[13px] text-primary-container">
                    chevron_right
                  </span>
                  <span class="bg-surface-container-low px-2 py-1">{{ part.name }}</span>
                </template>
              </div>
            </button>

            <button
              v-if="breadcrumbParts(option.folder).length > 2"
              class="bg-surface-container-high px-3 py-3 text-[10px] font-black uppercase tracking-widest text-primary-container transition-colors hover:bg-surface-container-highest"
              type="button"
              @click="emit('toggleBreadcrumb', option.folder)"
            >
              {{ expandedBreadcrumbs.has(breadcrumbKey(option.folder)) ? 'Collapse' : 'Expand' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
