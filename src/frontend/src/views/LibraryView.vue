<script setup lang="ts">
import VaultHeader from '../components/VaultHeader.vue'
import VaultCard from '../components/VaultCard.vue'
import LibraryMoveDialog from '../components/library/LibraryMoveDialog.vue'
import LibraryTransfersPanel from '../components/library/LibraryTransfersPanel.vue'

import { getMangaPageUrl } from '../api/manga'
import { useLibraryView } from '../composables/useLibraryView'

const {
  mangaStore,
  audioStore,
  videoStore,
  uploadStore,
  catalogStore,
  selectedTab,
  showActionDialog,
  showUploadDialog,
  showCatalogDialog,
  showRenameCatalogDialog,
  showDeleteCatalogDialog,
  showTransfersPanel,
  showMoveDialog,
  fileInputRef,
  newCatalogInputRef,
  newCatalogName,
  renameCatalogName,
  renamingFolder,
  deletingFolder,
  deletingMedia,
  expandedBreadcrumbs,
  filteredManga,
  filteredAudio,
  filteredVideo,
  totalMedia,
  hasVisibleMedia,
  floatingButtonsBottomClass,
  moveDestinationOptions,
  movingSubjectName,
  deletingSubjectName,
  deletingSubjectType,
  deletingImpactText,
  deleteDialogStyle,
  openFolder,
  openRoot,
  catalogName,
  breadcrumbKey,
  breadcrumbParts,
  visibleBreadcrumbParts,
  isBreadcrumbCollapsed,
  toggleBreadcrumb,
  closeMoveDialog,
  selectMoveDestination,
  videoStats,
  handleMangaClick,
  playAudio,
  openVideo,
  moveMedia,
  deleteManga,
  deleteVideo,
  toggleActionDialog,
  closeFloatingDialogs,
  openUploadDialog,
  openCatalogDialog,
  closeCatalogDialog,
  closeRenameCatalogDialog,
  closeDeleteCatalogDialog,
  openFilePicker,
  createFolderFromCatalogDialog,
  renameFolderFromCatalogDialog,
  deleteFromCatalogDialog,
  openTransfersPanel,
  closeTransfersPanel,
  handleFileInputChange,
  rejectUploadQueue,
  processAll,
} = useLibraryView()
</script>

<template>
  <div class="h-full flex flex-col">
    <main class="flex-1 overflow-y-auto">
      <section class="min-h-full p-4 md:p-8 flex-1 bg-surface">
        <div class="mb-8 flex items-center gap-2 overflow-x-auto whitespace-nowrap pb-2">
          <button
            class="px-2 py-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant transition-colors hover:text-primary"
            type="button"
            @click="openRoot"
          >
            CENTRAL
          </button>
          <template v-for="folder in catalogStore.path" :key="folder.id">
            <span class="material-symbols-outlined text-[14px] text-primary-container">
              chevron_right
            </span>
            <button
              class="bg-surface-container-low px-2 py-1 text-xs font-bold uppercase tracking-widest text-primary-container transition-colors hover:bg-surface-container-high"
              type="button"
              @click="openFolder(folder)"
            >
              {{ folder.name }}
            </button>
          </template>
        </div>

        <VaultHeader
          v-model="selectedTab"
          :totalManga="totalMedia"
          :catalogs="catalogStore.folders"
          @openCatalog="openFolder"
        />

        <div
          v-if="
            mangaStore.loading || audioStore.loading || videoStore.loading || catalogStore.loading
          "
          class="text-secondary"
        >
          Loading...
        </div>
        <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>
        <div v-else-if="audioStore.error" class="text-error">{{ audioStore.error }}</div>
        <div v-else-if="videoStore.error" class="text-error">{{ videoStore.error }}</div>
        <div v-else-if="catalogStore.error" class="text-error">
          {{ catalogStore.error }}
        </div>

        <div v-else class="space-y-12">
          <section class="space-y-6">
            <div
              v-if="hasVisibleMedia"
              class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5 md:gap-6"
            >
              <VaultCard
                v-for="manga in filteredManga"
                :key="`manga-${manga.id}`"
                :manga="manga"
                :coverUrl="getMangaPageUrl(manga.id, 0)"
                category="MANGA_ARCHIVE"
                :stats="`${manga.currentPage} / ${manga.pageCount} PAGES`"
                @click="handleMangaClick"
                @move="moveMedia"
                @delete="deleteManga"
              />

              <VaultCard
                v-for="audio in filteredAudio"
                :key="`audio-${audio.id}`"
                :audio="audio"
                category="AUDIO_ARCHIVE"
                :stats="`${Math.floor(audio.duration / 60)}:${String(audio.duration % 60).padStart(2, '0')} MIN`"
                @click="() => playAudio(audio)"
                @move="moveMedia"
              />

              <VaultCard
                v-for="video in filteredVideo"
                :key="`video-${video.id}`"
                :video="video"
                category="VIDEO_ARCHIVE"
                :stats="videoStats(video)"
                @click="() => openVideo(video)"
                @move="moveMedia"
                @delete="deleteVideo"
              />
            </div>
          </section>
        </div>
      </section>

      <button
        v-if="
          showActionDialog ||
          showUploadDialog ||
          showCatalogDialog ||
          showRenameCatalogDialog ||
          showDeleteCatalogDialog
        "
        class="fixed inset-0 z-30 cursor-default bg-transparent"
        type="button"
        aria-label="Close floating actions"
        @click="closeFloatingDialogs"
      ></button>

      <div class="fixed right-6 z-40 flex flex-col gap-3" :class="floatingButtonsBottomClass">
        <button
          class="relative flex h-14 w-14 items-center justify-center overflow-hidden bg-surface-container-high transition-colors hover:bg-surface-container-highest"
          type="button"
          aria-label="Open active transfers panel"
          @click="openTransfersPanel"
        >
          <i class="pi pi-upload text-s text-on-surface" aria-hidden="true"></i>
        </button>

        <div class="relative">
          <div
            v-if="showUploadDialog"
            class="absolute bottom-48 right-0 w-64 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
          >
            <input
              ref="fileInputRef"
              class="hidden"
              type="file"
              multiple
              accept=".cbz,.mp3,.wav,.flac,.ogg,.m4a,.aac,.mp4,.mkv,.webm,.mov,.avi,.m4v"
              @change="handleFileInputChange"
            />

            <button
              class="flex h-28 w-full items-center justify-center bg-surface-container-low text-5xl font-black leading-none text-primary-container transition-colors hover:bg-surface-container-high"
              type="button"
              aria-label="Select upload files"
              @click="openFilePicker"
            >
              +
            </button>

            <div v-if="uploadStore.queuedItems.length" class="mt-2 bg-surface-container-low p-2">
              <div class="mb-2 flex items-center justify-between">
                <span class="text-[8px] font-black uppercase tracking-widest text-secondary">
                  In_Queue
                </span>
                <span class="font-mono text-[8px] font-bold text-primary-container">
                  {{ String(uploadStore.queuedItems.length).padStart(2, '0') }}
                </span>
              </div>

              <div class="space-y-1.5">
                <div
                  v-for="(item, index) in uploadStore.queuedItems"
                  :key="item.id"
                  class="grid grid-cols-[1fr_auto] items-center gap-2"
                >
                  <div class="h-1.5 bg-surface-container-high">
                    <div
                      class="h-full bg-primary-container"
                      :class="[
                        index % 4 === 0
                          ? 'w-full'
                          : index % 4 === 1
                            ? 'w-10/12'
                            : index % 4 === 2
                              ? 'w-8/12'
                              : 'w-11/12',
                      ]"
                    ></div>
                  </div>
                  <span class="h-1.5 w-1.5 bg-secondary-container"></span>
                </div>
              </div>
            </div>

            <p
              v-if="uploadStore.rejectedItems.length > 0"
              class="mt-2 text-[9px] font-bold uppercase tracking-widest text-error"
            >
              Rejected: {{ uploadStore.rejectedItems.join(', ') }}
            </p>

            <div v-if="uploadStore.queuedItems.length" class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="rejectUploadQueue"
              >
                Reject
              </button>
              <button
                class="bg-primary-container px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary"
                type="button"
                @click="processAll"
              >
                Confirm
              </button>
            </div>
          </div>

          <form
            v-if="showCatalogDialog"
            class="absolute bottom-48 right-0 w-64 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            @submit.prevent="createFolderFromCatalogDialog"
          >
            <label class="block text-[10px] font-black uppercase tracking-widest text-secondary">
              New_Catalog
            </label>
            <input
              ref="newCatalogInputRef"
              v-model="newCatalogName"
              class="mt-2 w-full bg-surface-container-high px-3 py-3 text-sm font-bold uppercase tracking-tight text-on-surface outline-none focus:border-b-2 focus:border-primary-container"
              type="text"
              autocomplete="off"
              placeholder="CATALOG_NAME"
            />
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="closeCatalogDialog"
              >
                Reject
              </button>
              <button
                class="bg-primary-container px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary disabled:opacity-50"
                type="submit"
                :disabled="!newCatalogName.trim()"
              >
                Confirm
              </button>
            </div>
          </form>

          <form
            v-if="showRenameCatalogDialog"
            class="absolute bottom-48 right-0 w-64 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            @submit.prevent="renameFolderFromCatalogDialog"
          >
            <label class="block text-[10px] font-black uppercase tracking-widest text-secondary">
              Rename_Catalog
            </label>
            <p
              class="mt-1 truncate text-[9px] font-bold uppercase tracking-widest text-on-surface-variant"
            >
              {{ renamingFolder?.name }}
            </p>
            <input
              v-model="renameCatalogName"
              class="mt-2 w-full bg-surface-container-high px-3 py-3 text-sm font-bold uppercase tracking-tight text-on-surface outline-none focus:border-b-2 focus:border-primary-container"
              type="text"
              autocomplete="off"
              placeholder="CATALOG_NAME"
            />
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="closeRenameCatalogDialog"
              >
                Reject
              </button>
              <button
                class="bg-primary-container px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary disabled:opacity-50"
                type="submit"
                :disabled="!renameCatalogName.trim()"
              >
                Confirm
              </button>
            </div>
          </form>

          <form
            v-if="showDeleteCatalogDialog"
            class="fixed z-40 w-72 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            :style="deleteDialogStyle"
            @submit.prevent="deleteFromCatalogDialog"
          >
            <label class="block text-[10px] font-black uppercase tracking-widest text-error">
              Delete_{{ deletingSubjectType }}
            </label>
            <div class="mt-2 bg-surface-container-low p-3">
              <p class="truncate text-sm font-black uppercase tracking-tight text-on-surface">
                {{ deletingSubjectName }}
              </p>
              <p
                class="mt-2 text-[9px] font-bold uppercase leading-relaxed tracking-widest text-on-surface-variant"
              >
                {{ deletingImpactText }}
              </p>
            </div>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="closeDeleteCatalogDialog"
              >
                Reject
              </button>
              <button
                class="bg-error px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-error transition-colors hover:bg-secondary-container hover:text-on-secondary-container disabled:opacity-50"
                type="submit"
                :disabled="!deletingFolder && !deletingMedia"
              >
                Delete
              </button>
            </div>
          </form>

          <div
            v-if="showActionDialog"
            class="absolute bottom-16 right-0 w-48 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
          >
            <button
              class="flex w-full items-center gap-3 bg-primary-container px-3 py-3 text-left text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary"
              type="button"
              @click="openUploadDialog"
            >
              <span class="material-symbols-outlined text-lg">upload_file</span>
              Upload_File
            </button>
            <button
              class="mt-2 flex w-full items-center gap-3 bg-surface-container-high px-3 py-3 text-left text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
              type="button"
              @click="openCatalogDialog"
            >
              <span class="material-symbols-outlined text-lg text-primary-container"
                >create_new_folder</span
              >
              New_Catalog
            </button>
          </div>

          <button
            class="h-14 w-14 bg-primary-container text-on-primary-container text-3xl font-black leading-none transition-colors hover:bg-primary"
            type="button"
            aria-label="Open catalog actions"
            @click="toggleActionDialog"
          >
            +
          </button>
        </div>
      </div>

      <LibraryMoveDialog
        v-if="showMoveDialog"
        :movingSubjectName="movingSubjectName"
        :moveDestinationOptions="moveDestinationOptions"
        :expandedBreadcrumbs="expandedBreadcrumbs"
        :catalogName="catalogName"
        :breadcrumbKey="breadcrumbKey"
        :breadcrumbParts="breadcrumbParts"
        :visibleBreadcrumbParts="visibleBreadcrumbParts"
        :isBreadcrumbCollapsed="isBreadcrumbCollapsed"
        @close="closeMoveDialog"
        @selectDestination="selectMoveDestination"
        @toggleBreadcrumb="toggleBreadcrumb"
      />

      <LibraryTransfersPanel
        v-if="showTransfersPanel"
        :uploadStore="uploadStore"
        @close="closeTransfersPanel"
      />
    </main>
  </div>
</template>
