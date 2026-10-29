import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMangaStore } from '../stores/manga'
import { useAudioStore } from '../stores/audio'
import { useVideoStore } from '../stores/video'
import { usePlayerStore } from '../stores/player'
import { useUploadStore } from '../stores/upload'
import { useFolderStore } from '../stores/folder'
import type { MangaRow, AudioRow, VideoRow, Folder } from '../types/api'
import { folderSlug } from './useLibrarySlugs'
import { videoStats } from './useLibraryFormatting'
import { matchesMangaSearch } from '../utils/mangaMetadata'
import { selectFolderCoverChapters } from '../utils/folderCovers'
import { selectUploadFiles } from '../utils/uploadFiles'

interface MenuAnchor {
  top: number
  right: number
}

export function useLibraryView() {
  const route = useRoute()
  const router = useRouter()
  const mangaStore = useMangaStore()
  const audioStore = useAudioStore()
  const videoStore = useVideoStore()
  const playerStore = usePlayerStore()
  const uploadStore = useUploadStore()
  const folderStore = useFolderStore()

  const mediaFilter = ref<'ALL' | 'CBZ' | 'AUDIO' | 'VIDEO'>('ALL')
  const searchQuery = ref('')
  const scrollContainer = ref<HTMLElement | null>(null)
  const actionError = ref<string | null>(null)
  const actionPending = ref(false)

  async function performAction(action: () => Promise<void>) {
    if (actionPending.value) return
    actionPending.value = true
    actionError.value = null
    try {
      await action()
    } catch (error) {
      actionError.value = error instanceof Error ? error.message : 'Unable to complete this action.'
    } finally {
      actionPending.value = false
    }
  }
  const showActionDialog = ref(false)
  const showUploadDialog = ref(false)
  const showFolderDialog = ref(false)
  const showRenameFolderDialog = ref(false)
  const showDeleteFolderDialog = ref(false)
  const showTransfersPanel = ref(false)
  const showMoveDialog = ref(false)
  const fileInputRef = ref<HTMLInputElement | null>(null)
  const newFolderInputRef = ref<HTMLInputElement | null>(null)
  const renameFolderInputRef = ref<HTMLInputElement | null>(null)
  const newFolderName = ref('')
  const renameFolderName = ref('')
  const renamingFolder = ref<Folder | null>(null)
  const deletingFolder = ref<Folder | null>(null)
  const deletingMedia = ref<MangaRow | VideoRow | null>(null)
  const deleteDialogAnchor = ref<MenuAnchor | null>(null)
  const pendingMoveItem = ref<MangaRow | AudioRow | VideoRow | null>(null)
  const pendingMoveFolder = ref<Folder | null>(null)
  const pendingMoveExcludedIds = ref<Set<number>>(new Set())
  const expandedBreadcrumbs = ref<Set<string>>(new Set())

  const currentFolderId = computed(() => {
    const folder = route.query.folder
    if (typeof folder !== 'string') return null

    const id = Number(folder)
    if (Number.isInteger(id) && id > 0) return id

    return folderStore.allFolders.find((item) => folderSlug(item) === folder)?.id ?? -1
  })

  const folderManga = computed(() =>
    mangaStore.items.filter((item) => item.folderId === currentFolderId.value),
  )

  const folderAudio = computed(() =>
    audioStore.items.filter((item) => item.folderId === currentFolderId.value),
  )

  const folderVideo = computed(() =>
    videoStore.items.filter((item) => item.folderId === currentFolderId.value),
  )

  const filteredManga = computed(() => {
    if (mediaFilter.value === 'ALL' || mediaFilter.value === 'CBZ') {
      return folderManga.value.filter((item) => matchesMangaSearch(item, searchQuery.value))
    }
    return []
  })

  const filteredAudio = computed(() => {
    if (mediaFilter.value === 'ALL' || mediaFilter.value === 'AUDIO') {
      return folderAudio.value.filter((item) =>
        `${item.title} ${item.artist} ${item.album}`
          .toLowerCase()
          .includes(searchQuery.value.trim().toLowerCase()),
      )
    }
    return []
  })

  const filteredVideo = computed(() => {
    if (mediaFilter.value === 'ALL' || mediaFilter.value === 'VIDEO') {
      return folderVideo.value.filter((item) =>
        item.title.toLowerCase().includes(searchQuery.value.trim().toLowerCase()),
      )
    }
    return []
  })

  const totalMedia = computed(
    () => folderManga.value.length + folderAudio.value.length + folderVideo.value.length,
  )

  const currentFolder = computed(() => folderStore.path.at(-1))
  const folderCoverChapters = computed(() =>
    selectFolderCoverChapters(mangaStore.items, folderStore.allFolders),
  )
  const filteredFolders = computed(() =>
    folderStore.folders.filter((folder) =>
      folder.name.toLowerCase().includes(searchQuery.value.trim().toLowerCase()),
    ),
  )

  function clearFilters() {
    searchQuery.value = ''
    mediaFilter.value = 'ALL'
  }

  const hasVisibleMedia = computed(
    () => filteredManga.value.length + filteredAudio.value.length + filteredVideo.value.length > 0,
  )

  const floatingButtonsBottomClass = computed(() => {
    return playerStore.currentTrack
      ? 'bottom-[calc(7rem+env(safe-area-inset-bottom))]'
      : 'bottom-[calc(1.5rem+env(safe-area-inset-bottom))]'
  })

  watch(
    currentFolderId,
    async (folderId) => {
      await folderStore.fetchChildren(folderId)
      await nextTick()
      if (scrollContainer.value) scrollContainer.value.scrollTop = 0
    },
    { immediate: true },
  )

  watch(showFolderDialog, async (isOpen) => {
    if (!isOpen) return

    await nextTick()
    newFolderInputRef.value?.focus()
  })

  watch(showRenameFolderDialog, async (isOpen) => {
    if (!isOpen) return
    await nextTick()
    renameFolderInputRef.value?.focus()
    renameFolderInputRef.value?.select()
  })

  onMounted(() => {
    uploadStore.setOnComplete(refreshAfterUpload)
    mangaStore.fetchLibrary()
    audioStore.fetchLibrary()
    videoStore.fetchLibrary()
    void folderStore.fetchAll().catch(() => {})
    uploadStore.fetchPendingTransfers()
    window.addEventListener('keydown', handleGlobalKeydown)
  })

  onBeforeUnmount(() => {
    uploadStore.setOnComplete(null)
    window.removeEventListener('keydown', handleGlobalKeydown)
  })

  function openManga(manga: MangaRow) {
    router.push({
      path: '/reader',
      query: {
        manga: manga.id,
        page: manga.currentPage || 0,
        total: manga.pageCount,
      },
    })
  }

  function playAudio(audio: AudioRow) {
    playerStore.setQueue(audioStore.items)
    playerStore.play(audio)
  }

  function openVideo(video: VideoRow) {
    router.push({
      path: `/video/${video.id}`,
    })
  }

  function openFolder(folder: Folder) {
    clearFilters()
    router.push({ path: '/library', query: { folder: String(folder.id) } })
  }

  function openRoot() {
    clearFilters()
    router.push({ path: '/library' })
  }

  async function createFolder(name: string) {
    const trimmedName = name.trim()
    if (!trimmedName) return

    await folderStore.createFolder(trimmedName, currentFolderId.value)
  }

  function renameFolder(folder: Folder) {
    closeFloatingDialogs()
    actionError.value = null
    renamingFolder.value = folder
    renameFolderName.value = folder.name
    showRenameFolderDialog.value = true
  }

  function moveFolder(folder: Folder) {
    void openMoveDialog(null, folder)
  }

  function deleteFolder(folder: Folder, anchor: MenuAnchor) {
    openDeleteDialog(folder, null, anchor)
  }

  function folderName(folder: Folder) {
    return folderPath(folder)
      .map((item) => item.name)
      .join(' / ')
  }

  function folderPath(folder: Folder) {
    const byId = new Map(folderStore.allFolders.map((item) => [item.id, item]))
    const path: Folder[] = []
    let current: Folder | undefined = folder

    while (current) {
      path.unshift(current)
      current = current.parentId ? byId.get(current.parentId) : undefined
    }

    return path
  }

  const moveDestinationOptions = computed(() => {
    const currentDestinationId =
      pendingMoveFolder.value?.parentId ?? pendingMoveItem.value?.folderId ?? null
    const folderOptions = folderStore.allFolders
      .filter((folder) => !pendingMoveExcludedIds.value.has(folder.id))
      .map((folder) => ({ id: folder.id as number | null, folder }))

    return [{ id: null as number | null, folder: null as Folder | null }, ...folderOptions].filter(
      (option) => option.id !== currentDestinationId,
    )
  })

  const movingSubjectName = computed(() => {
    if (pendingMoveFolder.value) return pendingMoveFolder.value.name
    return pendingMoveItem.value?.title ?? ''
  })

  const deletingSubjectName = computed(() => {
    if (deletingFolder.value) return deletingFolder.value.name
    return deletingMedia.value?.title ?? ''
  })

  const deletingSubjectType = computed(() => {
    if (deletingFolder.value) return 'Folder'
    if (deletingMedia.value && 'pageCount' in deletingMedia.value) return 'Manga_Archive'
    if (deletingMedia.value) return 'Video_Stream'
    return 'Object'
  })

  const deletingImpactText = computed(() => {
    if (deletingFolder.value) {
      return 'Deletes this folder and its subfolders. Media inside moves to the library root.'
    }

    if (deletingMedia.value && 'pageCount' in deletingMedia.value) {
      return 'Removes its index and reading progress.'
    }

    return 'Removes its index but keeps the video file on disk.'
  })

  const deleteDialogStyle = computed(() => {
    if (!deleteDialogAnchor.value) return {}

    return {
      top: `${Math.max(16, Math.min(deleteDialogAnchor.value.top, window.innerHeight - 240))}px`,
      right: `${Math.max(16, deleteDialogAnchor.value.right)}px`,
    }
  })

  function breadcrumbKey(folder: Folder | null) {
    return folder ? `folder-${folder.id}` : 'root'
  }

  function breadcrumbParts(folder: Folder | null) {
    return folder ? folderPath(folder) : []
  }

  function visibleBreadcrumbParts(folder: Folder | null) {
    const parts = breadcrumbParts(folder)
    const key = breadcrumbKey(folder)

    if (expandedBreadcrumbs.value.has(key) || parts.length <= 2) return parts
    return parts.slice(-2)
  }

  function isBreadcrumbCollapsed(folder: Folder | null) {
    return breadcrumbParts(folder).length > visibleBreadcrumbParts(folder).length
  }

  function toggleBreadcrumb(folder: Folder | null) {
    const key = breadcrumbKey(folder)
    const next = new Set(expandedBreadcrumbs.value)

    if (next.has(key)) {
      next.delete(key)
    } else {
      next.add(key)
    }

    expandedBreadcrumbs.value = next
  }

  async function openMoveDialog(
    item: MangaRow | AudioRow | VideoRow | null,
    folder: Folder | null,
  ) {
    await performAction(async () => {
      await folderStore.fetchAll()
      pendingMoveItem.value = item
      pendingMoveFolder.value = folder
      const excludedIds = new Set<number>()
      if (folder) {
        excludedIds.add(folder.id)
        let changed = true
        while (changed) {
          changed = false
          for (const item of folderStore.allFolders) {
            if (
              item.parentId !== null &&
              excludedIds.has(item.parentId) &&
              !excludedIds.has(item.id)
            ) {
              excludedIds.add(item.id)
              changed = true
            }
          }
        }
      }
      pendingMoveExcludedIds.value = excludedIds
      expandedBreadcrumbs.value = new Set()
      showMoveDialog.value = true
    })
  }

  function closeMoveDialog() {
    showMoveDialog.value = false
    pendingMoveItem.value = null
    pendingMoveFolder.value = null
    pendingMoveExcludedIds.value = new Set()
    expandedBreadcrumbs.value = new Set()
  }

  async function selectMoveDestination(folderId: number | null) {
    await performAction(async () => {
      if (pendingMoveFolder.value) {
        await folderStore.updateFolder(
          pendingMoveFolder.value.id,
          pendingMoveFolder.value.name,
          folderId,
        )
        await folderStore.fetchChildren(currentFolderId.value)
        closeMoveDialog()
        return
      }

      const item = pendingMoveItem.value
      if (!item) return

      if ('pageCount' in item) {
        await mangaStore.moveManga(item.id, folderId)
      } else if ('artist' in item) {
        await audioStore.moveAudio(item.id, folderId)
      } else {
        await videoStore.moveVideo(item.id, folderId)
      }

      await folderStore.fetchChildren(currentFolderId.value)
      closeMoveDialog()
    })
  }

  function handleMangaClick(item: MangaRow | AudioRow | VideoRow) {
    if ('pageCount' in item) {
      openManga(item)
    }
  }

  function moveMedia(item: MangaRow | AudioRow | VideoRow) {
    void openMoveDialog(item, null)
  }

  function deleteManga(item: MangaRow | VideoRow, anchor: MenuAnchor) {
    if (!('pageCount' in item)) return

    openDeleteDialog(null, item, anchor)
  }

  function deleteVideo(item: MangaRow | VideoRow, anchor: MenuAnchor) {
    if ('pageCount' in item) return

    openDeleteDialog(null, item, anchor)
  }

  function openActionDialog() {
    showActionDialog.value = true
  }

  function closeActionDialog() {
    showActionDialog.value = false
  }

  function toggleActionDialog() {
    if (
      showActionDialog.value ||
      showUploadDialog.value ||
      showFolderDialog.value ||
      showRenameFolderDialog.value ||
      showDeleteFolderDialog.value
    ) {
      closeFloatingDialogs()
      return
    }

    openActionDialog()
  }

  function closeFloatingDialogs() {
    closeActionDialog()
    closeUploadDialog()
    closeFolderDialog()
    closeRenameFolderDialog()
    closeDeleteFolderDialog()
  }

  function closeUploadDialog() {
    showUploadDialog.value = false
  }

  function openFolderDialog() {
    actionError.value = null
    closeUploadDialog()
    closeRenameFolderDialog()
    closeDeleteFolderDialog()
    showFolderDialog.value = true
  }

  function closeFolderDialog() {
    showFolderDialog.value = false
    newFolderName.value = ''
  }

  function closeRenameFolderDialog() {
    showRenameFolderDialog.value = false
    renameFolderName.value = ''
    renamingFolder.value = null
  }

  function openDeleteDialog(
    folder: Folder | null,
    media: MangaRow | VideoRow | null,
    anchor: MenuAnchor,
  ) {
    actionError.value = null
    closeUploadDialog()
    closeFolderDialog()
    closeRenameFolderDialog()
    deletingFolder.value = folder
    deletingMedia.value = media
    deleteDialogAnchor.value = anchor
    showDeleteFolderDialog.value = true
  }

  function closeDeleteFolderDialog() {
    showDeleteFolderDialog.value = false
    deletingFolder.value = null
    deletingMedia.value = null
    deleteDialogAnchor.value = null
  }

  async function openFilePicker() {
    closeFloatingDialogs()
    const files = await selectUploadFiles(() => fileInputRef.value?.click())
    if (files.length) {
      await handleFilesSelected(files)
      showUploadDialog.value = true
    }
  }

  async function createFolderFromDialog() {
    await performAction(async () => {
      await createFolder(newFolderName.value)
      closeFloatingDialogs()
    })
  }

  async function renameFolderFromDialog() {
    await performAction(async () => {
      const folder = renamingFolder.value
      const name = renameFolderName.value.trim()
      if (!folder || !name || name === folder.name) {
        closeRenameFolderDialog()
        return
      }

      await folderStore.updateFolder(folder.id, name, folder.parentId)
      await folderStore.fetchChildren(currentFolderId.value)
      closeRenameFolderDialog()
    })
  }

  async function deleteFromFolderDialog() {
    await performAction(async () => {
      if (deletingFolder.value) {
        await folderStore.deleteFolder(deletingFolder.value.id)
        await Promise.all([
          folderStore.fetchChildren(currentFolderId.value),
          folderStore.fetchAll(),
          mangaStore.fetchLibrary(),
          audioStore.fetchLibrary(),
          videoStore.fetchLibrary(),
        ])
        closeDeleteFolderDialog()
        return
      }

      if (deletingMedia.value && 'pageCount' in deletingMedia.value) {
        await mangaStore.removeManga(deletingMedia.value.id)
        closeDeleteFolderDialog()
        return
      }

      if (deletingMedia.value) {
        await videoStore.removeVideo(deletingMedia.value.id)
        closeDeleteFolderDialog()
      }
    })
  }

  function openTransfersPanel() {
    showTransfersPanel.value = true
    uploadStore.fetchPendingTransfers()
  }

  function closeTransfersPanel() {
    showTransfersPanel.value = false
  }

  async function handleFilesSelected(files: File[]) {
    await uploadStore.enqueue(files, currentFolderId.value)
  }

  async function handleFileInputChange(event: Event) {
    const target = event.target as HTMLInputElement
    if (!target.files || target.files.length === 0) return

    await handleFilesSelected(Array.from(target.files))
    showUploadDialog.value = true
    target.value = ''
  }

  function clearQueue() {
    uploadStore.clearQueue()
  }

  function rejectUploadQueue() {
    clearQueue()
    closeUploadDialog()
  }

  function refreshAfterUpload() {
    void Promise.all([
      mangaStore.fetchLibrary(),
      audioStore.fetchLibrary(),
      videoStore.fetchLibrary(),
      folderStore.fetchChildren(currentFolderId.value),
      folderStore.fetchAll(),
    ])
  }

  function processAll() {
    closeFloatingDialogs()
    uploadStore.processAll()
    openTransfersPanel()
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key !== 'Escape') return

    if (
      showActionDialog.value ||
      showUploadDialog.value ||
      showFolderDialog.value ||
      showRenameFolderDialog.value ||
      showDeleteFolderDialog.value
    ) {
      closeFloatingDialogs()
    }

    if (showMoveDialog.value) {
      closeMoveDialog()
    }

    if (showTransfersPanel.value) {
      closeTransfersPanel()
    }
  }

  return {
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
    pendingMoveFolder,
    expandedBreadcrumbs,
    filteredManga,
    filteredAudio,
    filteredVideo,
    totalMedia,
    currentFolder,
    folderCoverChapters,
    filteredFolders,
    clearFilters,
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
    openMoveDialog,
    closeMoveDialog,
    selectMoveDestination,
    videoStats,
    handleMangaClick,
    playAudio,
    openVideo,
    moveMedia,
    renameFolder,
    moveFolder,
    deleteFolder,
    deleteManga,
    deleteVideo,
    openActionDialog,
    closeActionDialog,
    toggleActionDialog,
    closeFloatingDialogs,
    openFolderDialog,
    closeFolderDialog,
    closeRenameFolderDialog,
    openDeleteDialog,
    closeDeleteFolderDialog,
    openFilePicker,
    createFolderFromDialog,
    renameFolderFromDialog,
    deleteFromFolderDialog,
    openTransfersPanel,
    closeTransfersPanel,
    handleFileInputChange,
    clearQueue,
    rejectUploadQueue,
    processAll,
  }
}
