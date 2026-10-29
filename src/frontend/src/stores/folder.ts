import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { Folder } from '../types/api'
import {
  createFolder as createFolderRequest,
  deleteFolder as deleteFolderRequest,
  fetchAllFolders,
  fetchFolderChildren,
  updateFolder as updateFolderRequest,
} from '../api/folder'
import { compareByNaturalText } from '../utils/naturalSort'

const compareFolders = compareByNaturalText<Folder>((folder) => folder.name)

export const useFolderStore = defineStore('folder', () => {
  const folders = ref<Folder[]>([])
  const path = ref<Folder[]>([])
  const allFolders = ref<Folder[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  let childrenRequest = 0

  async function fetchChildren(parentId: number | null) {
    const request = ++childrenRequest
    loading.value = true
    error.value = null
    try {
      const response = await fetchFolderChildren(parentId)
      if (request !== childrenRequest) return
      folders.value = [...response.items].sort(compareFolders)
      path.value = response.path
    } catch (e) {
      if (request !== childrenRequest) return
      error.value = e instanceof Error ? e.message : 'Failed to fetch folder'
    } finally {
      if (request === childrenRequest) loading.value = false
    }
  }

  async function fetchAll() {
    error.value = null
    try {
      const response = await fetchAllFolders()
      allFolders.value = [...response.items].sort(compareFolders)
      return allFolders.value
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch folders'
      throw e
    }
  }

  async function createFolder(name: string, parentId: number | null) {
    error.value = null
    const folder = await createFolderRequest(name, parentId)
    folders.value = [...folders.value, folder].sort(compareFolders)
    allFolders.value = [...allFolders.value, folder].sort(compareFolders)
    return folder
  }

  async function updateFolder(id: number, name: string, parentId: number | null) {
    error.value = null
    const folder = await updateFolderRequest(id, name, parentId)
    allFolders.value = allFolders.value
      .map((item) => (item.id === id ? folder : item))
      .sort(compareFolders)
    folders.value = folders.value
      .map((item) => (item.id === id ? folder : item))
      .sort(compareFolders)
    return folder
  }

  async function deleteFolder(id: number) {
    error.value = null
    await deleteFolderRequest(id)
    folders.value = folders.value.filter((folder) => folder.id !== id)
    allFolders.value = allFolders.value.filter((folder) => folder.id !== id)
  }

  return {
    folders,
    path,
    allFolders,
    loading,
    error,
    fetchChildren,
    fetchAll,
    createFolder,
    updateFolder,
    deleteFolder,
  }
})
