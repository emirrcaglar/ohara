import type { CatalogFolder } from '../types/api'

export function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function folderSlug(folder: CatalogFolder) {
  return slugify(folder.name) || String(folder.id)
}
