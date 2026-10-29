<script setup lang="ts">
import VaultHeader from '../components/VaultHeader.vue'
import FolderCard from '../components/FolderCard.vue'
import VaultCard from '../components/VaultCard.vue'
import LibraryMoveDialog from '../components/library/LibraryMoveDialog.vue'
import LibraryTransfersPanel from '../components/library/LibraryTransfersPanel.vue'

import { getMangaCoverUrl } from '../api/manga'
import { useLibraryView } from '../composables/useLibraryView'

const {
  mangaStore,
  audioStore,
  videoStore,
  uploadStore,
  folderStore,
  mediaFilter,
  searchQuery,
  scrollContainer,
  actionError,
  actionPending,
  showActionDialog,
  showUploadDialog,
  showFolderDialog,
  showRenameFolderDialog,
  showDeleteFolderDialog,
  showTransfersPanel,
  showMoveDialog,
  fileInputRef,
  newFolderInputRef,
  renameFolderInputRef,
  newFolderName,
  renameFolderName,
  renamingFolder,
  deletingFolder,
  deletingMedia,
  expandedBreadcrumbs,
  filteredManga,
  filteredAudio,
  filteredVideo,
  totalMedia,
  currentFolder,
  folderCoverChapters,
  filteredFolders,
  clearFilters,
  renameFolder,
  moveFolder,
  deleteFolder,
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
  folderName,
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
  openFolderDialog,
  closeFolderDialog,
  closeRenameFolderDialog,
  closeDeleteFolderDialog,
  openFilePicker,
  createFolderFromDialog,
  renameFolderFromDialog,
  deleteFromFolderDialog,
  openTransfersPanel,
  closeTransfersPanel,
  handleFileInputChange,
  rejectUploadQueue,
  processAll,
} = useLibraryView()
</script>

<template>
  <div class="h-full flex flex-col">
    <main ref="scrollContainer" class="min-h-0 flex-1 overflow-y-auto">
      <section class="min-h-full p-4 pb-40 md:p-6 md:pb-40 flex-1 bg-surface">
        <nav
          v-if="folderStore.path.length"
          aria-label="Folder breadcrumb"
          class="mb-4 flex items-center gap-2 overflow-x-auto whitespace-nowrap pb-2"
        >
          <button
            class="px-2 py-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant transition-colors hover:text-primary"
            type="button"
            @click="openRoot"
          >
            Library
          </button>
          <template v-for="folder in folderStore.path" :key="folder.id">
            <span class="material-symbols-outlined text-[14px] text-primary-container">
              chevron_right
            </span>
            <button
              class="bg-surface-container-low px-2 py-1 text-xs font-bold uppercase tracking-widest text-primary-container transition-colors hover:bg-surface-container-high"
              type="button"
              :aria-current="folder.id === currentFolder?.id ? 'page' : undefined"
              @click="openFolder(folder)"
            >
              {{ folder.name }}
            </button>
          </template>
        </nav>

        <VaultHeader
          :title="currentFolder?.name ?? 'Library'"
          :folder-count="folderStore.folders.length"
          :media-count="totalMedia"
          @create-folder="openFolderDialog"
        />

        <div class="mb-6 grid grid-cols-[minmax(0,1fr)_auto] gap-2">
          <label
            class="flex items-center gap-2 bg-surface-container-low px-3 focus-within:shadow-[inset_0_-2px_0_var(--color-primary-container)]"
          >
            <span class="material-symbols-outlined text-lg text-primary" aria-hidden="true"
              >search</span
            >
            <input
              v-model="searchQuery"
              type="search"
              aria-label="Search this folder"
              placeholder="Search this folder…"
              class="min-w-0 flex-1 bg-transparent py-2.5 text-sm text-on-surface outline-none placeholder:text-on-surface-variant"
            />
          </label>
          <label class="flex items-center bg-surface-container-low px-2">
            <span class="sr-only">Media type</span>
            <select
              v-model="mediaFilter"
              aria-label="Media type"
              class="bg-transparent py-2 text-xs text-on-surface focus-visible:outline-2 focus-visible:outline-primary"
            >
              <option value="ALL">All media</option>
              <option value="CBZ">Manga</option>
              <option value="AUDIO">Audio</option>
              <option value="VIDEO">Video</option>
            </select>
          </label>
        </div>

        <p
          v-if="
            actionError &&
            !showFolderDialog &&
            !showRenameFolderDialog &&
            !showDeleteFolderDialog &&
            !showMoveDialog
          "
          role="alert"
          class="mb-4 bg-error-container p-4 text-sm text-on-error-container"
        >
          {{ actionError }}
        </p>

        <div
          v-if="
            mangaStore.loading || audioStore.loading || videoStore.loading || folderStore.loading
          "
          role="status"
          class="bg-surface-container-low p-8 text-sm text-on-surface-variant"
        >
          Loading folder…
        </div>
        <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>
        <div v-else-if="audioStore.error" class="text-error">{{ audioStore.error }}</div>
        <div v-else-if="videoStore.error" class="text-error">{{ videoStore.error }}</div>
        <div v-else-if="folderStore.error" class="text-error">
          {{ folderStore.error }}
        </div>

        <div v-else class="space-y-6">
          <section v-if="filteredFolders.length" aria-labelledby="folders-heading">
            <div class="mb-3 flex items-center gap-2">
              <h2
                id="folders-heading"
                class="text-sm font-bold uppercase tracking-widest text-on-surface"
              >
                Folders <span class="ml-2 text-secondary">{{ filteredFolders.length }}</span>
              </h2>
            </div>
            <div class="grid grid-cols-[repeat(auto-fill,minmax(min(100%,10rem),1fr))] gap-2">
              <FolderCard
                v-for="folder in filteredFolders"
                :key="folder.id"
                :folder="folder"
                :cover-chapter="folderCoverChapters.get(folder.id)"
                @open="openFolder"
                @rename="renameFolder"
                @move="moveFolder"
                @delete="deleteFolder"
              />
            </div>
          </section>
          <section v-if="hasVisibleMedia" aria-labelledby="files-heading" class="space-y-3">
            <h2
              id="files-heading"
              class="text-sm font-bold uppercase tracking-widest text-on-surface"
            >
              Media files
              <span class="ml-2 text-secondary">{{
                filteredManga.length + filteredAudio.length + filteredVideo.length
              }}</span>
            </h2>
            <div
              v-if="hasVisibleMedia"
              class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8"
            >
              <VaultCard
                v-for="manga in filteredManga"
                :key="`manga-${manga.id}`"
                :manga="manga"
                :coverUrl="getMangaCoverUrl(manga.id)"
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
          <div
            v-if="!filteredFolders.length && !hasVisibleMedia"
            class="bg-surface-container-low p-8 md:p-12"
          >
            <span class="material-symbols-outlined mb-5 text-5xl text-primary" aria-hidden="true"
              >folder_open</span
            >
            <template v-if="searchQuery || mediaFilter !== 'ALL'">
              <h2 class="text-2xl font-bold uppercase tracking-tight">No matching items</h2>
              <p class="mt-3 text-sm text-on-surface-variant">
                Try another search or show all media in this folder.
              </p>
              <button
                type="button"
                class="mt-6 bg-primary-container px-4 py-3 text-xs font-bold uppercase text-on-primary-container hover:bg-primary"
                @click="clearFilters"
              >
                Clear filters
              </button>
            </template>
            <template v-else>
              <h2 class="text-2xl font-bold uppercase tracking-tight">Room for your collection</h2>
              <p class="mt-3 text-sm text-on-surface-variant">
                Create a folder or upload media to get started here.
              </p>
              <button
                type="button"
                class="mt-6 bg-primary-container px-4 py-3 text-xs font-bold uppercase text-on-primary-container hover:bg-primary"
                @click="openFilePicker"
              >
                Upload files
              </button>
            </template>
          </div>
        </div>
      </section>

      <input
        ref="fileInputRef"
        class="hidden"
        type="file"
        multiple
        accept=".cbz,.mp3,.wav,.flac,.ogg,.m4a,.aac,.mp4,.mkv,.webm,.mov,.avi,.m4v"
        @change="handleFileInputChange"
      />

      <button
        v-if="
          showActionDialog ||
          showUploadDialog ||
          showFolderDialog ||
          showRenameFolderDialog ||
          showDeleteFolderDialog
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
            <div v-if="uploadStore.draftItems.length" class="mt-2 bg-surface-container-low p-2">
              <div class="mb-2 flex items-center justify-between">
                <span class="text-[8px] font-black uppercase tracking-widest text-secondary">
                  In_Queue
                </span>
                <span class="font-mono text-[8px] font-bold text-primary-container">
                  {{ String(uploadStore.draftItems.length).padStart(2, '0') }}
                </span>
              </div>

              <ul class="max-h-40 space-y-2 overflow-y-auto">
                <li
                  v-for="item in uploadStore.draftItems"
                  :key="item.id"
                  class="flex items-center justify-between gap-2 text-xs text-on-surface"
                  :title="item.name"
                >
                  <span class="truncate">{{ item.name }}</span>
                  <span
                    v-if="item.status === 'exists'"
                    class="shrink-0 text-[10px] text-on-surface-variant"
                    >Already exists</span
                  >
                </li>
              </ul>
            </div>

            <p
              v-if="uploadStore.rejectedItems.length > 0"
              class="mt-2 text-[9px] font-bold uppercase tracking-widest text-error"
            >
              Rejected: {{ uploadStore.rejectedItems.join(', ') }}
            </p>

            <div v-if="uploadStore.draftItems.length" class="mt-2 grid grid-cols-2 gap-2">
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
            v-if="showFolderDialog"
            class="absolute bottom-48 right-0 w-64 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            @submit.prevent="createFolderFromDialog"
            :aria-busy="actionPending"
          >
            <p
              v-if="actionError"
              role="alert"
              class="mb-3 bg-error-container p-3 text-sm text-on-error-container"
            >
              {{ actionError }}
            </p>
            <label
              for="new-folder-name"
              class="block text-[10px] font-black uppercase tracking-widest text-secondary"
            >
              New folder
            </label>
            <input
              id="new-folder-name"
              ref="newFolderInputRef"
              v-model="newFolderName"
              class="mt-2 w-full bg-surface-container-high px-3 py-3 text-sm font-bold uppercase tracking-tight text-on-surface outline-none focus:border-b-2 focus:border-primary-container"
              type="text"
              autocomplete="off"
              placeholder="FOLDER_NAME"
            />
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="closeFolderDialog"
              >
                Reject
              </button>
              <button
                class="bg-primary-container px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary disabled:opacity-50"
                type="submit"
                :disabled="actionPending || !newFolderName.trim()"
              >
                Confirm
              </button>
            </div>
          </form>

          <form
            v-if="showRenameFolderDialog"
            class="absolute bottom-48 right-0 w-64 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            @submit.prevent="renameFolderFromDialog"
            :aria-busy="actionPending"
          >
            <p
              v-if="actionError"
              role="alert"
              class="mb-3 bg-error-container p-3 text-sm text-on-error-container"
            >
              {{ actionError }}
            </p>
            <label
              for="rename-folder-name"
              class="block text-[10px] font-black uppercase tracking-widest text-secondary"
            >
              Rename folder
            </label>
            <p
              class="mt-1 truncate text-[9px] font-bold uppercase tracking-widest text-on-surface-variant"
            >
              {{ renamingFolder?.name }}
            </p>
            <input
              id="rename-folder-name"
              ref="renameFolderInputRef"
              v-model="renameFolderName"
              class="mt-2 w-full bg-surface-container-high px-3 py-3 text-sm font-bold uppercase tracking-tight text-on-surface outline-none focus:border-b-2 focus:border-primary-container"
              type="text"
              autocomplete="off"
              placeholder="FOLDER_NAME"
            />
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                class="bg-surface-container-high px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
                type="button"
                @click="closeRenameFolderDialog"
              >
                Reject
              </button>
              <button
                class="bg-primary-container px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-primary-container transition-colors hover:bg-primary disabled:opacity-50"
                type="submit"
                :disabled="actionPending || !renameFolderName.trim()"
              >
                Confirm
              </button>
            </div>
          </form>

          <form
            v-if="showDeleteFolderDialog"
            class="fixed z-40 w-72 bg-surface-container-highest/80 p-2 shadow-[0_0_28px_rgba(14,14,14,0.4)] backdrop-blur"
            :style="deleteDialogStyle"
            @submit.prevent="deleteFromFolderDialog"
            :aria-busy="actionPending"
          >
            <p
              v-if="actionError"
              role="alert"
              class="mb-3 bg-error-container p-3 text-sm text-on-error-container"
            >
              {{ actionError }}
            </p>
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
                @click="closeDeleteFolderDialog"
              >
                Reject
              </button>
              <button
                class="bg-error px-3 py-2 text-[10px] font-black uppercase tracking-widest text-on-error transition-colors hover:bg-secondary-container hover:text-on-secondary-container disabled:opacity-50"
                type="submit"
                :disabled="actionPending || (!deletingFolder && !deletingMedia)"
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
              @click="openFilePicker"
            >
              <span class="material-symbols-outlined text-lg">upload_file</span>
              Upload_File
            </button>
            <button
              class="mt-2 flex w-full items-center gap-3 bg-surface-container-high px-3 py-3 text-left text-[10px] font-black uppercase tracking-widest text-on-surface transition-colors hover:bg-surface-container-low"
              type="button"
              @click="openFolderDialog"
            >
              <span class="material-symbols-outlined text-lg text-primary-container"
                >create_new_folder</span
              >
              New_Folder
            </button>
          </div>

          <button
            class="h-14 w-14 bg-primary-container text-on-primary-container text-3xl font-black leading-none transition-colors hover:bg-primary"
            type="button"
            aria-label="Open folder actions"
            @click="toggleActionDialog"
          >
            +
          </button>
        </div>
      </div>

      <LibraryMoveDialog
        v-if="showMoveDialog"
        :movingSubjectName="movingSubjectName"
        :error="actionError"
        :pending="actionPending"
        :moveDestinationOptions="moveDestinationOptions"
        :expandedBreadcrumbs="expandedBreadcrumbs"
        :folderName="folderName"
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
