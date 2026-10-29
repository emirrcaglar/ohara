import type { Folder, MangaRow } from '../types/api'
import { compareManga } from './mangaMetadata'

// Prefer the latest read chapter in each subtree, falling back to the first
// readable manga in the same order used by the library.
export function selectFolderCoverChapters(chapters: MangaRow[], folders: Folder[]) {
  const parents = new Map(folders.map((folder) => [folder.id, folder.parentId]))
  const covers = new Map<number, MangaRow>()

  for (const chapter of chapters) {
    const timestamp = chapter.updatedAt ? Date.parse(chapter.updatedAt) : NaN
    if (chapter.pageCount <= 0) continue

    let folderId = chapter.folderId
    const visited = new Set<number>()
    while (folderId !== null && !visited.has(folderId)) {
      visited.add(folderId)
      const previous = covers.get(folderId)
      const previousTimestamp = previous?.updatedAt ? Date.parse(previous.updatedAt) : NaN
      const hasReadingHistory = Number.isFinite(timestamp)
      const previousHasHistory = Number.isFinite(previousTimestamp)
      if (
        !previous ||
        (hasReadingHistory &&
          (!previousHasHistory ||
            timestamp > previousTimestamp ||
            (timestamp === previousTimestamp && chapter.id > previous.id))) ||
        (!hasReadingHistory &&
          !previousHasHistory &&
          (compareManga(chapter, previous) < 0 ||
            (compareManga(chapter, previous) === 0 && chapter.id < previous.id)))
      ) {
        covers.set(folderId, chapter)
      }
      folderId = parents.get(folderId) ?? null
    }
  }

  return covers
}
