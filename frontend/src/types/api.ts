export interface MangaRow {
  id: number
  title: string
  path: string
  pageCount: number
  currentPage: number
  fileExtension: string
}

export interface MangaInfo {
  id: number
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
  title: string
  artist: string
  album: string
  duration: number
  fileExtension: string
}

export interface AudioLibraryResponse {
  items: AudioRow[]
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
