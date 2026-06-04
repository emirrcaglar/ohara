import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMangaStore } from '../stores/manga'
import { useAudioStore } from '../stores/audio'
import { useVideoStore } from '../stores/video'
import { usePlayerStore } from '../stores/player'
import { useUploadStore } from '../stores/upload'
import { useCatalogStore } from '../stores/catalog'
import type { MangaRow, AudioRow, VideoRow, CatalogFolder } from '../types/api'
import { folderSlug } from './useLibrarySlugs'
import { videoStats } from './useLibraryFormatting'

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
  const catalogStore = useCatalogStore()

  const selectedTab = ref<'ALL' | 'CBZ' | 'AUDIO' | 'VIDEO'>('ALL')
  const showActionDialog = ref(false)
  const showUploadDialog = ref(false)
  const showCatalogDialog = ref(false)
  const showRenameCatalogDialog = ref(false)
  const showDeleteCatalogDialog = ref(false)
  const showTransfersPanel = ref(false)
  const showMoveDialog = ref(false)
  const fileInputRef = ref<HTMLInputElement | null>(null)
  const newCatalogInputRef = ref<HTMLInputElement | null>(null)
  const newCatalogName = ref('')
  const renameCatalogName = ref('')
  const renamingFolder = ref<CatalogFolder | null>(null)
  const deletingFolder = ref<CatalogFolder | null>(null)
  const deletingMedia = ref<MangaRow | VideoRow | null>(null)
  const deleteDialogAnchor = ref<MenuAnchor | null>(null)
  const pendingMoveItem = ref<MangaRow | AudioRow | VideoRow | null>(null)
  const pendingMoveFolder = ref<CatalogFolder | null>(null)
  const pendingMoveExcludedIds = ref<Set<number>>(new Set())
  const expandedBreadcrumbs = ref<Set<string>>(new Set())

  const currentCatalogId = computed(() => {
    const folder = route.query.folder
    if (typeof folder !== 'string') return null

    const id = Number(folder)
    if (Number.isInteger(id) && id > 0) return id

    return catalogStore.catalogFolders.find((catalog) => folderSlug(catalog) === folder)?.id ?? null
  })

  const catalogManga = computed(() =>
    mangaStore.items.filter((item) => item.catalogId === currentCatalogId.value),
  )

  const catalogAudio = computed(() =>
    audioStore.items.filter((item) => item.catalogId === currentCatalogId.value),
  )

  const catalogVideo = computed(() =>
    videoStore.items.filter((item) => item.catalogId === currentCatalogId.value),
  )

  const filteredManga = computed(() => {
    if (selectedTab.value === 'ALL' || selectedTab.value === 'CBZ') return catalogManga.value
    return []
  })

  const filteredAudio = computed(() => {
    if (selectedTab.value === 'ALL' || selectedTab.value === 'AUDIO') return catalogAudio.value
    return []
  })

  const filteredVideo = computed(() => {
    if (selectedTab.value === 'ALL' || selectedTab.value === 'VIDEO') return catalogVideo.value
    return []
  })

  const totalMedia = computed(
    () => filteredManga.value.length + filteredAudio.value.length + filteredVideo.value.length,
  )

  const hasVisibleMedia = computed(
    () => filteredManga.value.length + filteredAudio.value.length + filteredVideo.value.length > 0,
  )

  const floatingButtonsBottomClass = computed(() => {
    return playerStore.currentTrack
      ? 'bottom-[calc(7rem+env(safe-area-inset-bottom))]'
      : 'bottom-[calc(1.5rem+env(safe-area-inset-bottom))]'
  })

  watch(
    currentCatalogId,
    (catalogId) => {
      catalogStore.fetchChildren(catalogId)
    },
    { immediate: true },
  )

  watch(showCatalogDialog, async (isOpen) => {
    if (!isOpen) return

    await nextTick()
    newCatalogInputRef.value?.focus()
  })

  onMounted(() => {
    mangaStore.fetchLibrary()
    audioStore.fetchLibrary()
    videoStore.fetchLibrary()
    catalogStore.fetchAll()
    uploadStore.fetchPendingTransfers()
    window.addEventListener('keydown', handleGlobalKeydown)
  })

  onBeforeUnmount(() => {
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

  function openFolder(folder: CatalogFolder) {
    router.push({ path: '/library', query: { folder: folderSlug(folder) } })
  }

  function openRoot() {
    router.push({ path: '/library' })
  }

  async function createFolder(name: string) {
    const trimmedName = name.trim()
    if (!trimmedName) return

    await catalogStore.createFolder(trimmedName, currentCatalogId.value)
  }

  function catalogName(folder: CatalogFolder) {
    return catalogPath(folder)
      .map((item) => item.name)
      .join(' / ')
  }

  function catalogPath(folder: CatalogFolder) {
    const byId = new Map(catalogStore.catalogFolders.map((item) => [item.id, item]))
    const path: CatalogFolder[] = []
    let current: CatalogFolder | undefined = folder

    while (current) {
      path.unshift(current)
      current = current.parentId ? byId.get(current.parentId) : undefined
    }

    return path
  }

  const moveDestinationOptions = computed(() => {
    const currentDestinationId =
      pendingMoveFolder.value?.parentId ?? pendingMoveItem.value?.catalogId ?? null
    const folderOptions = catalogStore.catalogFolders
      .filter((folder) => !pendingMoveExcludedIds.value.has(folder.id))
      .map((folder) => ({ id: folder.id as number | null, folder }))

    return [
      { id: null as number | null, folder: null as CatalogFolder | null },
      ...folderOptions,
    ].filter((option) => option.id !== currentDestinationId)
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
    if (deletingFolder.value) return 'Catalog'
    if (deletingMedia.value && 'pageCount' in deletingMedia.value) return 'Manga_Archive'
    if (deletingMedia.value) return 'Video_Stream'
    return 'Object'
  })

  const deletingImpactText = computed(() => {
    if (deletingFolder.value) {
      return 'Deletes this catalog and child catalogs. Media inside becomes uncataloged.'
    }

    if (deletingMedia.value && 'pageCount' in deletingMedia.value) {
      return 'Removes its index, reading progress, and cached pages.'
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

  function breadcrumbKey(folder: CatalogFolder | null) {
    return folder ? `folder-${folder.id}` : 'root'
  }

  function breadcrumbParts(folder: CatalogFolder | null) {
    return folder ? catalogPath(folder) : []
  }

  function visibleBreadcrumbParts(folder: CatalogFolder | null) {
    const parts = breadcrumbParts(folder)
    const key = breadcrumbKey(folder)

    if (expandedBreadcrumbs.value.has(key) || parts.length <= 2) return parts
    return parts.slice(-2)
  }

  function isBreadcrumbCollapsed(folder: CatalogFolder | null) {
    return breadcrumbParts(folder).length > visibleBreadcrumbParts(folder).length
  }

  function toggleBreadcrumb(folder: CatalogFolder | null) {
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
    folder: CatalogFolder | null,
    excludedIds = new Set<number>(),
  ) {
    await catalogStore.fetchAll()
    pendingMoveItem.value = item
    pendingMoveFolder.value = folder
    pendingMoveExcludedIds.value = excludedIds
    expandedBreadcrumbs.value = new Set()
    showMoveDialog.value = true
  }

  function closeMoveDialog() {
    showMoveDialog.value = false
    pendingMoveItem.value = null
    pendingMoveFolder.value = null
    pendingMoveExcludedIds.value = new Set()
    expandedBreadcrumbs.value = new Set()
  }

  async function selectMoveDestination(catalogId: number | null) {
    if (pendingMoveFolder.value) {
      await catalogStore.updateFolder(
        pendingMoveFolder.value.id,
        pendingMoveFolder.value.name,
        catalogId,
      )
      await catalogStore.fetchChildren(currentCatalogId.value)
      closeMoveDialog()
      return
    }

    const item = pendingMoveItem.value
    if (!item) return

    if ('pageCount' in item) {
      await mangaStore.moveManga(item.id, catalogId)
    } else if ('artist' in item) {
      await audioStore.moveAudio(item.id, catalogId)
    } else {
      await videoStore.moveVideo(item.id, catalogId)
    }

    await catalogStore.fetchChildren(currentCatalogId.value)
    closeMoveDialog()
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
      showCatalogDialog.value ||
      showRenameCatalogDialog.value ||
      showDeleteCatalogDialog.value
    ) {
      closeFloatingDialogs()
      return
    }

    openActionDialog()
  }

  function closeFloatingDialogs() {
    closeActionDialog()
    closeUploadDialog()
    closeCatalogDialog()
    closeRenameCatalogDialog()
    closeDeleteCatalogDialog()
  }

  function openUploadDialog() {
    closeCatalogDialog()
    closeRenameCatalogDialog()
    closeDeleteCatalogDialog()
    showUploadDialog.value = true
  }

  function closeUploadDialog() {
    showUploadDialog.value = false
  }

  function openCatalogDialog() {
    closeUploadDialog()
    closeRenameCatalogDialog()
    closeDeleteCatalogDialog()
    showCatalogDialog.value = true
  }

  function closeCatalogDialog() {
    showCatalogDialog.value = false
    newCatalogName.value = ''
  }

  function closeRenameCatalogDialog() {
    showRenameCatalogDialog.value = false
    renameCatalogName.value = ''
    renamingFolder.value = null
  }

  function openDeleteDialog(
    folder: CatalogFolder | null,
    media: MangaRow | VideoRow | null,
    anchor: MenuAnchor,
  ) {
    closeUploadDialog()
    closeCatalogDialog()
    closeRenameCatalogDialog()
    deletingFolder.value = folder
    deletingMedia.value = media
    deleteDialogAnchor.value = anchor
    showDeleteCatalogDialog.value = true
  }

  function closeDeleteCatalogDialog() {
    showDeleteCatalogDialog.value = false
    deletingFolder.value = null
    deletingMedia.value = null
    deleteDialogAnchor.value = null
  }

  function openFilePicker() {
    fileInputRef.value?.click()
  }

  async function createFolderFromCatalogDialog() {
    await createFolder(newCatalogName.value)
    closeFloatingDialogs()
  }

  async function renameFolderFromCatalogDialog() {
    const folder = renamingFolder.value
    const name = renameCatalogName.value.trim()
    if (!folder || !name || name === folder.name) {
      closeRenameCatalogDialog()
      return
    }

    await catalogStore.updateFolder(folder.id, name, folder.parentId)
    await catalogStore.fetchChildren(currentCatalogId.value)
    closeRenameCatalogDialog()
  }

  async function deleteFromCatalogDialog() {
    if (deletingFolder.value) {
      await catalogStore.deleteFolder(deletingFolder.value.id)
      await Promise.all([
        catalogStore.fetchChildren(currentCatalogId.value),
        mangaStore.fetchLibrary(),
        audioStore.fetchLibrary(),
        videoStore.fetchLibrary(),
      ])
      closeDeleteCatalogDialog()
      return
    }

    if (deletingMedia.value && 'pageCount' in deletingMedia.value) {
      await mangaStore.removeManga(deletingMedia.value.id)
      closeDeleteCatalogDialog()
      return
    }

    if (deletingMedia.value) {
      await videoStore.removeVideo(deletingMedia.value.id)
      closeDeleteCatalogDialog()
    }
  }

  function openTransfersPanel() {
    showTransfersPanel.value = true
    uploadStore.fetchPendingTransfers()
  }

  function closeTransfersPanel() {
    showTransfersPanel.value = false
  }

  function handleFilesSelected(files: File[]) {
    uploadStore.enqueue(files, currentCatalogId.value)
  }

  function handleFileInputChange(event: Event) {
    const target = event.target as HTMLInputElement
    if (!target.files || target.files.length === 0) return

    handleFilesSelected(Array.from(target.files))
    target.value = ''
  }

  function clearQueue() {
    uploadStore.clearQueue()
  }

  function rejectUploadQueue() {
    clearQueue()
    closeUploadDialog()
  }

  function processAll() {
    closeFloatingDialogs()
    uploadStore.setOnComplete(() => {
      void Promise.all([
        mangaStore.fetchLibrary(),
        audioStore.fetchLibrary(),
        videoStore.fetchLibrary(),
        catalogStore.fetchChildren(currentCatalogId.value),
      ])
    })
    uploadStore.processAll()
    openTransfersPanel()
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key !== 'Escape') return

    if (
      showActionDialog.value ||
      showUploadDialog.value ||
      showCatalogDialog.value ||
      showRenameCatalogDialog.value ||
      showDeleteCatalogDialog.value
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
    pendingMoveFolder,
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
    openMoveDialog,
    closeMoveDialog,
    selectMoveDestination,
    videoStats,
    handleMangaClick,
    playAudio,
    openVideo,
    moveMedia,
    deleteManga,
    deleteVideo,
    openActionDialog,
    closeActionDialog,
    toggleActionDialog,
    closeFloatingDialogs,
    openUploadDialog,
    openCatalogDialog,
    closeCatalogDialog,
    closeRenameCatalogDialog,
    openDeleteDialog,
    closeDeleteCatalogDialog,
    openFilePicker,
    createFolderFromCatalogDialog,
    renameFolderFromCatalogDialog,
    deleteFromCatalogDialog,
    openTransfersPanel,
    closeTransfersPanel,
    handleFileInputChange,
    clearQueue,
    rejectUploadQueue,
    processAll,
  }
}
