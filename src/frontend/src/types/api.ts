export interface MangaNumber {
  whole: number
  decimal: number | null
}

export interface MangaNumberRange {
  start: MangaNumber
  end: MangaNumber | null
}

export interface MangaMetadata {
  volume: MangaNumberRange | null
  chapter: MangaNumberRange | null
  group: string
  source: string
  language: string
  revision: number | null
  isOneshot: boolean
  edition: string
}

export interface MangaRelease {
  seriesTitle: string
  fileName: string
  metadata: MangaMetadata | null
  parserVersion: string
}

export interface MangaRow extends MangaRelease {
  id: number
  folderId: number | null
  title: string
  path: string
  pageCount: number
  currentPage: number
  fileExtension: string
  updatedAt: string | null
}

export interface MangaInfo extends MangaRelease {
  id: number
  folderId: number | null
  title: string
  path: string
  pageCount: number
  currentPage: number
  pages: MangaPage[]
}

export interface MangaPage {
  index: number
  name: string
  width: number
  height: number
}

export interface AudioRow {
  id: number
  folderId: number | null
  title: string
  artist: string
  album: string
  duration: number
  fileExtension: string
}

export interface VideoRow {
  id: number
  folderId: number | null
  title: string
  duration: number
  width: number
  height: number
  position: number
  completed: boolean
  lastError: string
  fileExtension: string
}

export interface VideoInfo {
  id: number
  folderId: number | null
  title: string
  path: string
  duration: number
  width: number
  height: number
  position: number
  completed: boolean
  lastError: string
  fileExtension: string
}

export interface VideoStateUpdate {
  duration: number
  width: number
  height: number
  position: number
  completed: boolean
  lastError: string
}

export interface Folder {
  id: number
  parentId: number | null
  name: string
  objectCount: number
}

export interface FolderListResponse {
  items: Folder[]
  path: Folder[]
}

export interface AudioLibraryResponse {
  items: AudioRow[]
  total: number
}

export interface VideoLibraryResponse {
  items: VideoRow[]
  total: number
}

export interface MangaLibraryResponse {
  items: MangaRow[]
  total: number
}

export interface LogEntry {
  time: string
  level: string
  message: string
}

export interface PlayerState {
  currentTrack: AudioRow | null
  isPlaying: boolean
  currentTime: number
  duration: number
  volume: number
  queue: AudioRow[]
}
