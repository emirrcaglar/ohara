import type { Folder } from '../types/api'

export function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function folderSlug(folder: Folder) {
  return slugify(folder.name) || String(folder.id)
}
