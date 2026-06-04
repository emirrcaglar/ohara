import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { CatalogFolder } from '../types/api'
import {
  createCatalogFolder,
  deleteCatalogFolder,
  fetchAllCatalogs,
  fetchCatalog,
  updateCatalogFolder,
} from '../api/catalog'

export const useCatalogStore = defineStore('catalog', () => {
  const folders = ref<CatalogFolder[]>([])
  const path = ref<CatalogFolder[]>([])
  const allFolders = ref<CatalogFolder[]>([])
  const catalogFolders = computed(() => allFolders.value)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchChildren(parentId: number | null) {
    loading.value = true
    error.value = null
    try {
      const response = await fetchCatalog(parentId)
      folders.value = response.items
      path.value = response.path
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch catalog'
    } finally {
      loading.value = false
    }
  }

  async function fetchAll() {
    error.value = null
    try {
      const response = await fetchAllCatalogs()
      allFolders.value = response.items
      return response.items
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch catalogs'
      throw e
    }
  }

  async function createFolder(name: string, parentId: number | null) {
    error.value = null
    try {
      const folder = await createCatalogFolder(name, parentId)
      folders.value = [...folders.value, folder].sort((a, b) => a.name.localeCompare(b.name))
      allFolders.value = [...allFolders.value, folder].sort((a, b) => a.name.localeCompare(b.name))
      return folder
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create folder'
      throw e
    }
  }

  async function updateFolder(id: number, name: string, parentId: number | null) {
    error.value = null
    try {
      const folder = await updateCatalogFolder(id, name, parentId)
      allFolders.value = allFolders.value.map((item) => (item.id === id ? folder : item))
      return folder
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to update folder'
      throw e
    }
  }

  async function deleteFolder(id: number) {
    error.value = null
    try {
      await deleteCatalogFolder(id)
      folders.value = folders.value.filter((folder) => folder.id !== id)
      allFolders.value = allFolders.value.filter((folder) => folder.id !== id)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete folder'
      throw e
    }
  }

  return {
    folders,
    path,
    allFolders,
    catalogFolders,
    loading,
    error,
    fetchChildren,
    fetchAll,
    createFolder,
    updateFolder,
    deleteFolder,
  }
})
